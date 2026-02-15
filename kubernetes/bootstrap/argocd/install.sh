#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Creating argocd namespace..."
kubectl apply -f "${SCRIPT_DIR}/namespace.yaml"

echo "==> Adding ArgoCD Helm repository..."
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

echo "==> Installing ArgoCD via Helm..."
helm install argocd argo/argo-cd \
  --namespace argocd \
  --values "${SCRIPT_DIR}/values.yaml" \
  --wait

echo "==> Waiting for ArgoCD server to be ready..."
kubectl -n argocd rollout status deployment/argocd-server --timeout=300s

echo "==> Applying root App-of-Apps..."
kubectl apply -f "${SCRIPT_DIR}/root-app.yaml"

echo "==> ArgoCD bootstrap complete."
echo ""
echo "To get the initial admin password, run:"
echo "  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d; echo"
echo ""
echo "To port-forward the ArgoCD UI:"
echo "  kubectl -n argocd port-forward svc/argocd-server 8080:443"
