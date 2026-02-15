#!/bin/bash
set -euo pipefail

# Home Lab K3s Cluster Bootstrap Script
# This script orchestrates the full cluster provisioning:
#   1. (Optional) Provision VMs via Pulumi
#   2. Prepare nodes via Ansible
#   3. Install K3s via Ansible
#   4. Bootstrap ArgoCD for GitOps

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ANSIBLE_DIR="$PROJECT_ROOT/ansible"
PULUMI_DIR="$PROJECT_ROOT/pulumi"
K8S_DIR="$PROJECT_ROOT/kubernetes"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()  { echo -e "${RED}[ERROR]${NC} $*" >&2; }

check_deps() {
    local missing=()
    for cmd in ansible-playbook kubectl helm; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        err "Missing required tools: ${missing[*]}"
        exit 1
    fi
}

provision_vms() {
    if [[ ! -d "$PULUMI_DIR" ]]; then
        warn "Pulumi directory not found, skipping VM provisioning"
        return 0
    fi

    if ! command -v pulumi &>/dev/null; then
        warn "Pulumi not installed, skipping VM provisioning"
        return 0
    fi

    log "Provisioning VMs via Pulumi..."
    cd "$PULUMI_DIR"
    pulumi up --yes
    cd "$PROJECT_ROOT"
    log "VM provisioning complete"
}

prep_nodes() {
    log "Preparing nodes (OS config, packages, firewall)..."
    ansible-playbook -i "$ANSIBLE_DIR/inventory/hosts.yaml" \
        "$ANSIBLE_DIR/playbooks/prep-nodes.yaml"
    log "Node preparation complete"
}

install_k3s() {
    log "Installing K3s cluster..."
    ansible-playbook -i "$ANSIBLE_DIR/inventory/hosts.yaml" \
        "$ANSIBLE_DIR/playbooks/install-k3s.yaml"
    log "K3s installation complete"
}

bootstrap_argocd() {
    if [[ ! -f "$K8S_DIR/bootstrap/argocd/install.sh" ]]; then
        warn "ArgoCD install script not found, skipping"
        return 0
    fi

    log "Bootstrapping ArgoCD..."
    chmod +x "$K8S_DIR/bootstrap/argocd/install.sh"
    bash "$K8S_DIR/bootstrap/argocd/install.sh"
    log "ArgoCD bootstrap complete"
}

verify_cluster() {
    log "Verifying cluster health..."
    if kubectl get nodes &>/dev/null; then
        kubectl get nodes -o wide
        echo
        kubectl get pods -A
        log "Cluster is healthy"
    else
        err "Cannot reach cluster. Check kubeconfig at ~/.kube/config"
        exit 1
    fi
}

usage() {
    cat <<EOF
Usage: $(basename "$0") [COMMAND]

Commands:
  all         Run full bootstrap (default)
  vms         Provision VMs via Pulumi only
  prep        Prepare nodes via Ansible only
  k3s         Install K3s only
  argocd      Bootstrap ArgoCD only
  verify      Verify cluster health
  reset       Tear down K3s cluster

Options:
  -h, --help  Show this help message
EOF
}

reset_cluster() {
    warn "This will DESTROY the K3s cluster on all nodes."
    read -rp "Are you sure? (y/N): " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        log "Aborted"
        exit 0
    fi
    ansible-playbook -i "$ANSIBLE_DIR/inventory/hosts.yaml" \
        "$ANSIBLE_DIR/playbooks/reset-k3s.yaml"
    log "Cluster has been torn down"
}

main() {
    check_deps

    local cmd="${1:-all}"

    case "$cmd" in
        all)
            provision_vms
            prep_nodes
            install_k3s
            verify_cluster
            bootstrap_argocd
            ;;
        vms)     provision_vms ;;
        prep)    prep_nodes ;;
        k3s)     install_k3s ;;
        argocd)  bootstrap_argocd ;;
        verify)  verify_cluster ;;
        reset)   reset_cluster ;;
        -h|--help) usage ;;
        *)
            err "Unknown command: $cmd"
            usage
            exit 1
            ;;
    esac

    log "Done!"
}

main "$@"
