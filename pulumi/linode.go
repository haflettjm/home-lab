package main

import (
	"fmt"

	"github.com/pulumi/pulumi-linode/sdk/v5/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// LinodeConfig holds all Linode-specific configuration.
type LinodeConfig struct {
	Region       string
	SSHPublicKey string
	RootPassword string
	Label        string
	// WireGuard tunnel endpoint — your home router's public IP or DDNS hostname.
	HomeEndpoint    string
	HomeWgPublicKey string
	// The K3s ingress IP on your home network that WireGuard will forward to.
	HomeIngressIP string
}

// LoadLinodeConfig reads Linode settings from the Pulumi stack config.
func LoadLinodeConfig(ctx *pulumi.Context) (*LinodeConfig, error) {
	cfg := config.New(ctx, "")

	region := cfg.Get("linodeRegion")
	if region == "" {
		region = "us-ord" // Chicago — good central US latency
	}

	label := cfg.Get("linodeLabel")
	if label == "" {
		label = "homelab-edge"
	}

	sshKey := cfg.Require("sshPublicKey")
	rootPass := cfg.RequireSecret("linodeRootPassword")

	homeEndpoint := cfg.Get("homeEndpoint")
	homeWgPubKey := cfg.Get("homeWgPublicKey")
	homeIngressIP := cfg.Get("homeIngressIP")
	if homeIngressIP == "" {
		homeIngressIP = "192.168.1.200" // Default MetalLB ingress IP
	}

	// RequireSecret returns a pulumi.StringOutput; for the config struct we
	// need the plaintext. Pulumi encrypts it in the stack file.
	_ = rootPass

	return &LinodeConfig{
		Region:          region,
		SSHPublicKey:    sshKey,
		RootPassword:    cfg.Get("linodeRootPassword"),
		Label:           label,
		HomeEndpoint:    homeEndpoint,
		HomeWgPublicKey: homeWgPubKey,
		HomeIngressIP:   homeIngressIP,
	}, nil
}

// ProvisionLinodeEdge creates the Linode edge infrastructure:
//   - A Nanode (1GB) running Rocky Linux 9 as a WireGuard VPN endpoint
//   - A NodeBalancer for public HTTPS traffic
//   - NodeBalancer config (port 443 → backend on the Nanode)
//   - NodeBalancer node pointing to the Nanode's private IP
//
// Traffic flow: Internet → NodeBalancer:443 → Nanode:443 → WireGuard tunnel → Home K3s Ingress
func ProvisionLinodeEdge(ctx *pulumi.Context, lcfg *LinodeConfig) error {
	// ── Edge VPS (Nanode 1GB — $5/month) ────────────────────────────
	// This runs WireGuard and nginx to forward traffic through the
	// tunnel back to the home lab K3s cluster.
	edge, err := linode.NewInstance(ctx, "edge-node", &linode.InstanceArgs{
		Label:  pulumi.String(lcfg.Label),
		Image:  pulumi.String("linode/rocky9"),
		Region: pulumi.String(lcfg.Region),
		Type:   pulumi.String("g6-nanode-1"),
		AuthorizedKeys: pulumi.StringArray{
			pulumi.String(lcfg.SSHPublicKey),
		},
		RootPass:  pulumi.String(lcfg.RootPassword),
		PrivateIp: pulumi.Bool(true),
		Booted:    pulumi.Bool(true),
		Tags: pulumi.StringArray{
			pulumi.String("homelab"),
			pulumi.String("edge"),
			pulumi.String("wireguard"),
		},
	})
	if err != nil {
		return fmt.Errorf("creating edge node: %w", err)
	}

	// ── NodeBalancer ($10/month) ────────────────────────────────────
	// Sits in front of the edge node, provides a stable public IP,
	// health checks, and connection throttling.
	nb, err := linode.NewNodeBalancer(ctx, "edge-lb", &linode.NodeBalancerArgs{
		Label:              pulumi.String(fmt.Sprintf("%s-lb", lcfg.Label)),
		Region:             pulumi.String(lcfg.Region),
		ClientConnThrottle: pulumi.Int(20),
		Tags: pulumi.StringArray{
			pulumi.String("homelab"),
			pulumi.String("edge"),
		},
	})
	if err != nil {
		return fmt.Errorf("creating NodeBalancer: %w", err)
	}

	// ── NodeBalancer Config: HTTPS (port 443) ───────────────────────
	// TCP passthrough — TLS termination happens at the home lab's
	// NGINX Ingress Controller, not at the NodeBalancer. This lets
	// cert-manager manage certs inside the cluster.
	nbConfigHTTPS, err := linode.NewNodeBalancerConfig(ctx, "edge-lb-https", &linode.NodeBalancerConfigArgs{
		NodebalancerId: nb.ID(),
		Port:           pulumi.Int(443),
		Protocol:       pulumi.String("tcp"),
		Algorithm:      pulumi.String("roundrobin"),
		Check:          pulumi.String("connection"),
		CheckInterval:  pulumi.Int(30),
		CheckTimeout:   pulumi.Int(10),
		CheckAttempts:  pulumi.Int(3),
		Stickiness:     pulumi.String("table"),
	})
	if err != nil {
		return fmt.Errorf("creating NodeBalancer HTTPS config: %w", err)
	}

	// ── NodeBalancer Config: HTTP (port 80) ─────────────────────────
	// Forwards HTTP to the edge node for Let's Encrypt HTTP-01
	// challenges or HTTP→HTTPS redirects.
	nbConfigHTTP, err := linode.NewNodeBalancerConfig(ctx, "edge-lb-http", &linode.NodeBalancerConfigArgs{
		NodebalancerId: nb.ID(),
		Port:           pulumi.Int(80),
		Protocol:       pulumi.String("tcp"),
		Algorithm:      pulumi.String("roundrobin"),
		Check:          pulumi.String("connection"),
		CheckInterval:  pulumi.Int(30),
		CheckTimeout:   pulumi.Int(10),
		CheckAttempts:  pulumi.Int(3),
		Stickiness:     pulumi.String("table"),
	})
	if err != nil {
		return fmt.Errorf("creating NodeBalancer HTTP config: %w", err)
	}

	// ── Backend Nodes ───────────────────────────────────────────────
	// Point both configs at the edge Nanode's private IP on ports
	// 443 and 80 respectively. Nginx on the Nanode forwards through
	// the WireGuard tunnel to the home lab.
	_, err = linode.NewNodeBalancerNode(ctx, "edge-backend-https", &linode.NodeBalancerNodeArgs{
		NodebalancerId: nb.ID(),
		ConfigId:       nbConfigHTTPS.ID(),
		Label:          pulumi.String("edge-https"),
		Address: edge.PrivateIpAddress.ApplyT(func(ip string) string {
			return fmt.Sprintf("%s:443", ip)
		}).(pulumi.StringOutput),
		Weight: pulumi.Int(100),
	})
	if err != nil {
		return fmt.Errorf("creating HTTPS backend node: %w", err)
	}

	_, err = linode.NewNodeBalancerNode(ctx, "edge-backend-http", &linode.NodeBalancerNodeArgs{
		NodebalancerId: nb.ID(),
		ConfigId:       nbConfigHTTP.ID(),
		Label:          pulumi.String("edge-http"),
		Address: edge.PrivateIpAddress.ApplyT(func(ip string) string {
			return fmt.Sprintf("%s:80", ip)
		}).(pulumi.StringOutput),
		Weight: pulumi.Int(100),
	})
	if err != nil {
		return fmt.Errorf("creating HTTP backend node: %w", err)
	}

	// ── Stack Exports ───────────────────────────────────────────────
	ctx.Export("edgeNodeIP", edge.IpAddress)
	ctx.Export("edgeNodePrivateIP", edge.PrivateIpAddress)
	ctx.Export("nodeBalancerHostname", nb.Hostname)
	ctx.Export("nodeBalancerIPv4", nb.Ipv4)
	ctx.Export("edgeNodeRegion", pulumi.String(lcfg.Region))

	return nil
}
