# 🏡 Home-Lab Kubernetes Platform

This repository manages the infrastructure, configuration, and application deployment for a self-hosted Kubernetes cluster built from scratch using physical machines in a home-lab environment.

The project is designed to scale from lightweight workloads to production-grade services, with future expansion into cloud-backed hybrid infrastructure.

---

## Current Phase:
### **Phase(s): 01-02**

---

## 📦 Features

- **K3s-based Kubernetes cluster** with optional migration to kubeadm or cloud-managed K8s
- **Infrastructure-as-Code** using Ansible and Terraform
- **Automated provisioning scripts** for first-time node setup
- **NGINX Ingress Controller** with optional Traefik fallback
- **Persistent storage support** (e.g., Longhorn or NFS)
- **Monitoring stack** with Prometheus, Grafana, and Loki
- **Modular application deployment** via Terraform modules
- **Planned support** for GitOps, secret management, and remote scaling

---

## 🔁 Project Phases

| Phase | Description |
|-------|-------------|
| 1. Inventory Management       | Track machine specs manually or via script. |
| 2. Infrastructure Preparation | Set up OS, SSH, VM layers, and packages. |
| 3. Kubernetes Setup           | Bootstrap cluster with K3s. |
| 4. Networking & Storage       | DNS, Ingress (NGINX), persistent volumes. |
| 5. App Hosting                | Deploy apps using Terraform. |
| 6. Monitoring & Observability | Prometheus, Grafana, logging, etc. |
| 7. Maintenance & Scaling      | Node mgmt, patching, backups, cloud-ready. |

---

## 🧱 Project Structure

```
home-lab/
├── phase1-inventory/         # Hardware discovery, inventory tracking scripts
│   └── inventory.yaml
│
├── phase2-prep/              # OS prep scripts and Ansible roles (SSH, packages, users)
│   ├── ansible/
│   └── prep-fedora.sh
│
├── phase3-k8s-install/       # K3s install, control plane bootstrap, node joining
│   ├── k3s-install.yml
│   └── kubeconfig/
│
├── phase4-networking/        # Ingress, MetalLB, DNS, cert-manager
│   ├── helm-charts/
│   └── scripts/
│       └── setup-networking.sh
│
├── phase5-apps/              # Terraform modules for application deployment
│   ├── terraform/
│   │   ├── modules/
│   │   └── environments/
│   └── k8s/apps/
│
├── phase6-monitoring/        # Monitoring stack (Prometheus, Grafana, Loki)
│   └── manifests/
│
├── phase7-maintenance/       # Backups, scaling, updates, long-term ops
│   ├── upgrade-scripts/
│   └── node-tools/
│
├── scripts/                  # Bootstrap wrapper (infra-up.sh, checks, helpers)
│   └── infra-up.sh
│
├── HOME-LAB.md               # Full system plan and status tracker
└── README.md                 # Top-level project overview
```

---


## 🚀 Getting Started

### 1. Install Fedora Server on All Nodes
- Use the `prep-fedora.sh` script to configure each machine
- Ensure SSH access and hostnames are set correctly

### 2. Define Hardware Inventory
Update `inventory.yaml` with each node’s role, IP, and hardware specs.

### 3. Run Infrastructure Bootstrap

```bash
cd scripts/
./infra-up.sh
```

This will:
- Run system prep via Ansible
- Install K3s and required components
- Deploy base infrastructure (NGINX, MetalLB, cert-manager)
- Launch initial applications via Terraform

---

## 📈 Planned Services

- GitOps via ArgoCD or Flux
- Monitoring with Prometheus stack
- Secret management (Vault, Sealed Secrets)
- Local LLM inference (Ollama, Mistral)
- Public app hosting (NGINX Ingress + Cert-Manager)
- Home automation, dashboards, file sharing

---

## 📚 Documentation

- [HOME-LAB.md](./HOME-LAB.md) — full project plan & architecture
- `ansible/` — node setup and cluster install roles
- `terraform/` — application modules and environment configs

---

## 🛠️ Requirements

- 3–4 physical machines (4+ core, 16GB RAM, 128GB SSD recommended)
- Fedora Server (or similar Linux)
- Ansible + Terraform installed on control node
- SSH access between control node and all cluster machines

---

## 🤝 Contributing

This project is designed for personal experimentation and practical DevOps learning. PRs, suggestions, and issue reports are welcome if you're interested in expanding it.

---

## 📜 License

????
