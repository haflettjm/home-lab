# 📦 Phase 1: Inventory & Node Prep

This phase focuses on preparing all machines in the home-lab cluster with consistent base configuration and capturing their identity in a static Ansible-compatible inventory file.

---

## 🎯 Objectives

- Establish a standard naming convention for all nodes
- Prepare each machine with the required base software
- Enable SSH and Cockpit for management access
- Configure encrypted nodes for LUKS auto-unlock
- Manually define the inventory in `inventory.yaml`

---

## 🧰 Step 1: Bootstrap Each Node

Run the `bootstrap-node.sh` script **on each machine** to standardize:

- Hostname (prompted at runtime)
- SSH server with password login
- Cockpit web UI
- Time sync with `chronyd`
- LUKS auto-unlock (if applicable)

### ▶️ Usage:

```bash
chmod +x bootstrap-node.sh
sudo ./bootstrap-node.sh
```

> ⚠️ Run this as the `admin` user. You'll be prompted to enter the desired hostname.

---

## 🧾 Step 2: Track Progress

Use the [PREP-CHECKLIST.md](./prep/PREP-CHECKLIST.md) file to track which machines have been successfully bootstrapped.

---

## 📒 Step 3: Define Static Inventory

Create `inventory.yaml` with Ansible-compatible structure:

---

## 💡 Next Steps

After all machines are bootstrapped and inventoried:

- Lock in DHCP reservations or static IPs 
- Validate SSH access to each host
- Move into Phase 2: OS Configuration via Ansible

