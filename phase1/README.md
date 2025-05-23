# 🧱 Phase 1: Professional-Grade Infrastructure Prep TODO

## 🖥️ Section A: OS Installation & Baseline System Setup

- [x] ✅ Download latest Fedora Server ISO (LTS preferred for stability)
- [x] ✅ Flash USB (or PXE boot if automating installs)
- [x] 🚀 Install Fedora on each machine:
  - [ ] Set hostname (naming convention: `cluster-role-host`, e.g. `k8s-cp-rbrain`)
  - [ ] Set static IP or reserve via DHCP
  - [ ] Record Static Reservations and Hostname(s)
  - [ ] Partition disks (consider LVM for flexibility) (?) - maybe I should?
  - [ ] Set timezone, locale, and NTP (chronyd)
  - [ ] Create non-root user (e.g. `opsadmin`)
  - [ ] Harden SSH:
    - [ ] Disable root login
    - [ ] Disable password auth
    - [ ] Allow only ed25519 key-based login (?) - possible

---

## 🔐 Section B: Access Control & SSH Management

- [ ] Generate SSH keypair (ed25519) on control machine
  ```bash
  ssh-keygen -t ed25519 -C "home-lab"
  ```
- [ ] Copy public key to all nodes
  ```bash
  ssh-copy-id opsadmin@<IP>
  ```
- [ ] Create a secure `.ssh/config` entry for each node for quick access
- [ ] Centralize SSH key storage in password manager or backup repo

---

## 🧰 Section C: Post-Install System Hardening & Package Setup

- [ ] Create and run a `prep-fedora.sh` post-install script that:
  - [ ] Updates system packages
  - [ ] Installs base tools: `vim`, `htop`, `tmux`, `curl`, `git`, `bash-completion`
  - [ ] Installs virtualization tools (if node is a hypervisor): `qemu-kvm`, `libvirt`, `bridge-utils`
  - [ ] Enables & starts `libvirtd` where needed
  - [ ] Adds user to `libvirt`, `wheel`, and `kvm` groups
  - [ ] Sets hostname from argument
  - [ ] Masks sleep/hibernate services
  - [ ] Enables and checks firewall with custom zones (firewalld)
  - [ ] Enables automatic security updates (dnf-automatic)
  - [ ] Logs all steps to `/var/log/node-prep.log`
- [ ] Copy and run script on each machine via SSH

---

## 🛡️ Section D: OS Hardening & Compliance (Production-Ready)

- [ ] Enable and configure SELinux (enforcing)
- [ ] Configure firewalld with zones and ingress/egress rules
- [ ] Install and enable `fail2ban`
- [ ] Configure auditd
- [ ] Set up system-wide resource limits and swappiness
- [ ] Configure journald persistent storage

---

## 📋 Section E: Infrastructure & Cluster Inventory

- [ ] Create `inventory.yaml` for Ansible with:
  - Hostname
  - IP Address
  - Role: `control-plane`, `worker`, `storage`, `infra`
  - CPU, RAM, Disk
  - Tags: `hypervisor`, `ssd`, `arm64`, etc.
- [ ] Version control this file in your Git repo (`home-lab/inventory.yaml`)
- [ ] Test inventory with Ansible:
  ```bash
  ansible -i inventory.yaml all -m ping
  ```

---

## 🗂 Section F: Git Project Structure & Documentation

- [ ] Initialize Git repo structure:
  ```
  home-lab/
  ├── inventory.yaml
  ├── ansible/
  │   └── roles/
  ├── terraform/
  ├── scripts/
  ├── k8s/
  ├── docs/
  ├── HOME-LAB.md
  └── TODO.md
  ```
- [ ] Add `.gitignore` and `.editorconfig`
- [ ] Add README with overview of goals
- [ ] Track versions of each machine and OS (in `/docs/hosts.md`)

---

## 📦 Section G: Optional – Prepare for VM Automation

- [ ] Install `cloud-utils`, `virt-install`, `genisoimage`
- [ ] Prepare `cloud-init` base image with SSH key injection
- [ ] Write VM provisioning playbook or shell script
- [ ] Ensure VMs boot with hostname, static IP, and key-based access

---

## 🔁 Section H: Operational Practices to Prep

- [ ] Add initial system state backups (`/etc`, user SSH keys, `hostnamectl`, etc.)
- [ ] Set up local logging directory or centralized syslog if building it
- [ ] Document machine names, MACs, IPs, and roles in a spreadsheet or YAML file
- [ ] Plan for scheduled snapshots or VM backups if using KVM or libvirt

---

## 🧠 Professional Practice Notes

- [ ] Track every IP, hostname, and MAC address
- [ ] Keep consistent usernames and shell environment across nodes
- [ ] Use Git to track your infra code from Day 1
- [ ] Write README-style runbooks for every script or process you automate
- [ ] Test all scripts in isolation first before applying cluster-wide

---

> ✅ Once all of this is done, **Phase 2** is:  
> *Ansible automation to install K3s, disable Traefik, install MetalLB, and prepare kubeconfig access.*

