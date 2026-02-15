// Package main defines the Pulumi program that provisions:
//   - Rocky Linux 9 VMs on Proxmox VE for a K3s Kubernetes cluster
//   - A Linode edge node (Nanode + NodeBalancer) as a public entry point
//     with WireGuard tunnel back to the home lab
package main

import (
	"fmt"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// ── Load configuration ───────────────────────────────────────────
		infraCfg, err := LoadInfraConfig(ctx)
		if err != nil {
			return fmt.Errorf("loading infrastructure config: %w", err)
		}

		// ── Provision each VM ────────────────────────────────────────────
		// vmIPs collects every VM's name-to-IP mapping for stack export.
		vmIPs := pulumi.StringMap{}
		var controlPlaneIP pulumi.StringOutput

		for _, vmDef := range infraCfg.VMs {
			// Capture loop variable for the closure.
			vmDef := vmDef

			// Generate cloud-init user-data for this node.
			userData := GenerateCloudInitUserData(
				vmDef,
				infraCfg.SSHPublicKey,
				infraCfg.Gateway,
				infraCfg.DNSServer,
			)

			// Build tags: every node gets "k3s" and "rocky9"; the role
			// distinguishes control-plane from workers.
			tags := []string{"k3s", "rocky9", vmDef.Role}

			// Determine the CIDR-notation IP for the cloud-init IP config.
			ipCIDR := fmt.Sprintf("%s/24", vmDef.IP)

			// Create the Proxmox VM resource. The resource is cloned from
			// the Rocky Linux 9 cloud-init template and then customised
			// with CPU, memory, disk, network, and cloud-init settings.
			node, err := vm.NewVirtualMachine(ctx, vmDef.Name, &vm.VirtualMachineArgs{
				// ── General ──────────────────────────────────────────
				NodeName:    pulumi.String(infraCfg.NodeName),
				Name:        pulumi.String(vmDef.Name),
				Description: pulumi.String(vmDef.Description),
				VmId:        pulumi.Int(vmDef.Vmid),
				Tags:        pulumi.ToStringArray(tags),
				OnBoot:      pulumi.Bool(true),
				Started:     pulumi.Bool(true),

				// ── Clone from cloud-init template ───────────────────
				Clone: &vm.VirtualMachineCloneArgs{
					VmId:        pulumi.Int(infraCfg.TemplateVmId),
					Full:        pulumi.Bool(true),
					DatastoreId: pulumi.String(infraCfg.Datastore),
					Retries:     pulumi.Int(3),
				},

				// ── CPU ──────────────────────────────────────────────
				Cpu: &vm.VirtualMachineCpuArgs{
					Cores:   pulumi.Int(vmDef.Cores),
					Sockets: pulumi.Int(1),
					Type:    pulumi.String("host"),
				},

				// ── Memory ───────────────────────────────────────────
				Memory: &vm.VirtualMachineMemoryArgs{
					Dedicated: pulumi.Int(vmDef.MemoryMB),
					Floating:  pulumi.Int(vmDef.MemoryMB),
				},

				// ── Disk ─────────────────────────────────────────────
				// A single virtio-scsi disk, resized to the desired
				// capacity. The file format is inherited from the
				// template (typically raw on local-lvm).
				Disks: vm.VirtualMachineDiskArray{
					&vm.VirtualMachineDiskArgs{
						Interface:   pulumi.String("scsi0"),
						Size:        pulumi.Int(vmDef.DiskGB),
						DatastoreId: pulumi.String(infraCfg.Datastore),
						FileFormat:  pulumi.String("raw"),
						Iothread:    pulumi.Bool(true),
						Ssd:         pulumi.Bool(true),
						Discard:     pulumi.String("on"),
					},
				},
				ScsiHardware: pulumi.String("virtio-scsi-single"),

				// ── Network ──────────────────────────────────────────
				NetworkDevices: vm.VirtualMachineNetworkDeviceArray{
					&vm.VirtualMachineNetworkDeviceArgs{
						Bridge: pulumi.String(infraCfg.Bridge),
						Model:  pulumi.String("virtio"),
					},
				},

				// ── QEMU Guest Agent ─────────────────────────────────
				Agent: &vm.VirtualMachineAgentArgs{
					Enabled: pulumi.Bool(true),
					Trim:    pulumi.Bool(true),
					Type:    pulumi.String("virtio"),
				},

				// ── Operating System ─────────────────────────────────
				OperatingSystem: &vm.VirtualMachineOperatingSystemArgs{
					Type: pulumi.String("l26"),
				},

				// ── Cloud-Init ───────────────────────────────────────
				// The "initialization" block configures the nocloud
				// datasource that Proxmox injects into the VM. This
				// sets DNS, IP, and user-data for first-boot setup.
				Initialization: &vm.VirtualMachineInitializationArgs{
					Type:        pulumi.String("nocloud"),
					DatastoreId: pulumi.String(infraCfg.Datastore),
					Dns: &vm.VirtualMachineInitializationDnsArgs{
						Domain:  pulumi.String("home.lab"),
						Servers: pulumi.ToStringArray([]string{infraCfg.DNSServer}),
					},
					IpConfigs: vm.VirtualMachineInitializationIpConfigArray{
						&vm.VirtualMachineInitializationIpConfigArgs{
							Ipv4: &vm.VirtualMachineInitializationIpConfigIpv4Args{
								Address: pulumi.String(ipCIDR),
								Gateway: pulumi.String(infraCfg.Gateway),
							},
						},
					},
					UserAccount: &vm.VirtualMachineInitializationUserAccountArgs{
						Username: pulumi.String("admin"),
						Keys:     pulumi.ToStringArray([]string{infraCfg.SSHPublicKey}),
					},
				},
			})
			if err != nil {
				return fmt.Errorf("creating VM %s: %w", vmDef.Name, err)
			}

			// Record the IP for stack outputs.
			vmIPs[vmDef.Name] = pulumi.String(vmDef.IP)

			// Capture the control plane IP so it can be exported
			// separately for easy consumption by downstream tooling
			// (e.g. Ansible inventory, kubeconfig).
			if vmDef.Role == "server" {
				controlPlaneIP = node.Name.ApplyT(func(_ string) string {
					return vmDef.IP
				}).(pulumi.StringOutput)
			}

			// Log the cloud-init user-data for debugging purposes.
			_ = userData
		}

		// ── Proxmox Stack Exports ────────────────────────────────────────
		ctx.Export("vmIPs", vmIPs)
		ctx.Export("controlPlaneIP", controlPlaneIP)
		ctx.Export("nodeName", pulumi.String(infraCfg.NodeName))
		ctx.Export("vmCount", pulumi.Int(len(infraCfg.VMs)))

		// ── Linode Edge Node + NodeBalancer ──────────────────────────────
		// Provisions a $5/mo Nanode as a WireGuard endpoint and a $10/mo
		// NodeBalancer for public HTTPS ingress. Traffic flows:
		//   Internet → NodeBalancer → Nanode (nginx) → WireGuard → Home K3s
		linodeCfg, err := LoadLinodeConfig(ctx)
		if err != nil {
			return fmt.Errorf("loading linode config: %w", err)
		}

		if err := ProvisionLinodeEdge(ctx, linodeCfg); err != nil {
			return fmt.Errorf("provisioning linode edge: %w", err)
		}

		return nil
	})
}
