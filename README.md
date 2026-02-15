# Home-Lab Kubernetes Platform

Self-hosted K3s cluster on physical hardware, provisioned with Ansible and managed via ArgoCD GitOps.

**Base OS:** Rocky Linux 9
**Hypervisor:** Proxmox VE (hybrid — control plane virtualized, workers bare-metal)
**IaC:** Ansible (OS/K3s/WireGuard) + Pulumi (Proxmox VMs + Linode edge)
**GitOps:** ArgoCD (App-of-Apps pattern)
**Public Ingress:** Linode NodeBalancer + Nanode → WireGuard tunnel → Home K3s

---

## Hardware

| Hostname | Hardware | IP | Role | Approach |
|----------|----------|----|------|----------|
| `k3s-master-01` | Minisforum UM690 | 192.168.1.100 | K3s Control Plane | Proxmox VM |
| `k3s-node-01` | HP Envy | 192.168.1.101 | K3s Worker | Bare-metal |
| `k3s-node-02` | Dell Optiplex Micro 1 | 192.168.1.102 | K3s Worker | Bare-metal |
| `k3s-node-03` | Dell Optiplex Micro 2 | 192.168.1.103 | K3s Worker | Bare-metal |
| `k3s-node-04` | Dell Optiplex Micro 3 | 192.168.1.104 | K3s Worker | Bare-metal |
| `nas-node-01` | Robo-brain | 192.168.1.105 | NAS / Storage | Bare-metal + Podman |
| `homelab-edge` | Linode Nanode 1GB | (public) | WireGuard + iptables forwarder | Linode ($5/mo) |
| — | Linode NodeBalancer | (public) | Public HTTPS entry point | Linode ($10/mo) |

---

## Stack

| Layer | Tool | Notes |
|-------|------|-------|
| OS | Rocky Linux 9 | Stable RHEL-based, 10-year support |
| Hypervisor | Proxmox VE | Control plane VM for snapshotting |
| VM Provisioning | Pulumi (Go) | bpg/proxmoxve provider, cloud-init templates |
| Edge Infra | Pulumi (Go) + Linode | Nanode ($5/mo) + NodeBalancer ($10/mo) |
| VPN Tunnel | WireGuard | Encrypted tunnel from Linode edge to home network |
| Config Management | Ansible | Node prep, K3s install, NAS setup, WireGuard |
| Kubernetes | K3s | Lightweight, single-binary |
| CNI | Cilium | eBPF-based, replaces kube-proxy, Hubble UI |
| Load Balancer | MetalLB | L2 mode, 192.168.1.200-220 |
| Ingress | NGINX Ingress Controller | DaemonSet mode |
| TLS | cert-manager | Cloudflare DNS-01, Let's Encrypt |
| Storage | Longhorn + NFS | Longhorn default (2 replicas), NFS from NAS |
| GitOps | ArgoCD | App-of-Apps, auto-sync + self-heal |
| Secrets | Sealed Secrets | Git-safe encrypted secrets |
| Monitoring | kube-prometheus-stack + Loki | Prometheus, Grafana, Alertmanager, log aggregation |
| Containers (NAS) | Podman | Standalone containers on NAS node |

---

## Project Structure

```
home-lab/
├── ansible/                    # OS prep + K3s provisioning
│   ├── ansible.cfg
│   ├── inventory/
│   │   ├── hosts.yaml          # Node inventory
│   │   └── group_vars/
│   │       ├── all.yaml        # Global vars (k3s version, packages, etc.)
│   │       ├── control_plane.yaml
│   │       ├── workers.yaml
│   │       └── nas.yaml
│   ├── playbooks/
│   │   ├── site.yaml           # Full run (prep + install)
│   │   ├── prep-nodes.yaml     # OS config, packages, firewall
│   │   ├── install-k3s.yaml    # K3s server + agent install
│   │   └── reset-k3s.yaml     # Tear down cluster
│   └── roles/
│       ├── common/             # Base packages, chrony, sysctl, swap, SELinux
│       ├── prereqs/            # K3s prereqs, firewall ports, iscsid
│       ├── k3s_server/         # K3s server install + kubeconfig fetch
│       ├── k3s_agent/          # K3s agent join
│       ├── nas/                # Podman + NFS server setup
│       ├── wireguard_edge/     # WireGuard + iptables on Linode Nanode
│       └── wireguard_home/     # WireGuard + iptables on home K3s node
│
├── pulumi/                     # Infrastructure provisioning (Go)
│   ├── Pulumi.yaml
│   ├── main.go                 # Orchestrates Proxmox + Linode
│   ├── config.go               # Proxmox VM config loading
│   ├── linode.go               # Linode edge node + NodeBalancer
│   └── go.mod
│
├── kubernetes/                 # GitOps root (ArgoCD watches this)
│   ├── bootstrap/
│   │   └── argocd/             # ArgoCD initial install + root app
│   ├── infrastructure/         # Cluster infrastructure (sync wave ordered)
│   │   ├── cilium/             # Wave 1: CNI
│   │   ├── metallb/            # Wave 2: Load balancer + IP pool
│   │   ├── ingress-nginx/      # Wave 3: Ingress controller
│   │   ├── cert-manager/       # Wave 3: TLS + ClusterIssuers
│   │   ├── longhorn/           # Wave 3: Distributed storage
│   │   ├── sealed-secrets/     # Wave 3: Secret management
│   │   ├── nfs-provisioner/    # Wave 3: NAS storage class
│   │   └── monitoring/         # Wave 4: Prometheus + Loki
│   └── apps/                   # User applications (deployed via ArgoCD)
│
├── scripts/
│   └── bootstrap.sh            # Full cluster orchestrator
│
├── phase1/                     # (Legacy) Hardware inventory + prep checklist
│   ├── inventory.yaml
│   └── prep/PREP-CHECKLIST.md
│
├── phase2/                     # (Legacy) Package requirement lists
│   └── reqs/
│
├── .gitignore
└── README.md
```

---

## Getting Started

### Prerequisites

On your admin workstation:
- Ansible, Pulumi, kubectl, Helm, k9s
- SSH access to all nodes (as `admin` user)

### 1. Provision VMs (if using Proxmox)

```bash
cd pulumi/
pulumi up
```

### 2. Prepare All Nodes

```bash
cd ansible/
ansible-playbook -i inventory/hosts.yaml playbooks/prep-nodes.yaml
```

### 3. Install K3s

```bash
ansible-playbook -i inventory/hosts.yaml playbooks/install-k3s.yaml
```

### 4. Bootstrap ArgoCD

```bash
bash kubernetes/bootstrap/argocd/install.sh
```

ArgoCD will auto-deploy everything in `kubernetes/infrastructure/` via the App-of-Apps.

### 5. Set Up Public Ingress (Linode + WireGuard)

Pulumi provisions the Linode edge infra (Nanode + NodeBalancer) alongside Proxmox VMs.
After `pulumi up`, configure the WireGuard tunnel:

```bash
# Get the edge node's public IP
pulumi stack output edgeNodeIP

# Update inventory with the edge IP
# → ansible/inventory/hosts.yaml: homelab-edge → ansible_host
# → ansible/inventory/group_vars/wg_home.yaml: wg_peer_endpoint

# Generate keys on home side, then edge side, swap public keys
ansible-playbook -i inventory/hosts.yaml playbooks/setup-wireguard.yaml --limit wg_home
# Copy displayed public key → group_vars/edge.yaml → wg_peer_public_key

ansible-playbook -i inventory/hosts.yaml playbooks/setup-wireguard.yaml --limit edge
# Copy displayed public key → group_vars/wg_home.yaml → wg_peer_public_key

# Deploy final configs with correct peer keys
ansible-playbook -i inventory/hosts.yaml playbooks/setup-wireguard.yaml

# Port forward 51820/UDP on your home router to 192.168.1.101 (k3s-node-01)
# Point DNS *.yourdomain.com → NodeBalancer IP (pulumi stack output nodeBalancerIPv4)
```

**Traffic flow:**
```
Internet → NodeBalancer ($10/mo) → Nanode ($5/mo) → WireGuard → Home K3s → NGINX Ingress → Apps
```

### Or Run Everything

```bash
./scripts/bootstrap.sh all
```

### Tear Down

```bash
./scripts/bootstrap.sh reset
```

---

## Planned Services

- Local LLM inference (Ollama)
- Home automation dashboards
- File sharing / media server
- Personal wiki / notes

---

## TODO-READ

- [Rocky Linux for Beginners: Build Your First Homelab with Proxmox, Docker, Kubernetes, Ansible & Modern Home Servers](https://www.amazon.com/Rocky-Linux-Beginners-Homelab-Kubernetes/dp/B0GC51PHBK)

---

## Requirements

- 5 physical machines (specs vary, see Hardware table)
- Rocky Linux 9 on all nodes (or Proxmox VE on hypervisor nodes)
- Ansible + Pulumi + Helm + kubectl on admin workstation
- SSH access from admin workstation to all nodes
- (Optional) Domain + Cloudflare account for TLS certificates
- Linode account + API token for edge node ($15/mo total)
- Home router capable of port forwarding UDP 51820 for WireGuard

---

## License

????
