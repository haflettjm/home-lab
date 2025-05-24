# 🧾 Home-Lab Node Prep Checklist

Track which machines have been bootstrapped with the base configuration using `bootstrap-node.sh`.

| Hostname        | Role             | Hardware               | Tier     | Prep Status |
|-----------------|------------------|------------------------|----------|-------------|
| `k3s-master-01` | Control Plane    | Minisforum UM690       | `high`   | [x]         |
| `k3s-node-01`   | Worker           | HP Envy                | `medium` | [ ]         |
| `k3s-node-02`   | Worker           | Optiplex Micro 1       | `low`    | [x]         |
| `k3s-node-03`   | Worker           | Optiplex Micro 2       | `low`    | [ ]         |
| `k3s-node-04`   | Worker           | Optiplex Micro 3       | `low`    | [ ]         |
| `nas-node-01`   | NAS / Storage    | Robo-brain             | `nas`    | [ ]         |

---

## ✅ What "Prepped" Means:
- Hostname set
- Base packages + Cockpit installed
- SSH enabled and allows password login
- Time sync (chrony) enabled
- LUKS auto-unlock configured (if encrypted)

> Ran `bootstrap-node.sh` on each machine 

