# User Applications

This directory is where user application manifests and ArgoCD Application resources should be placed.

## Structure

Each application should have its own subdirectory containing either:

- An ArgoCD `Application` resource (for Helm charts or external repos)
- Raw Kubernetes manifests (Deployments, Services, Ingress, etc.)
- A Kustomization if using Kustomize overlays

## Example

```
apps/
  my-app/
    application.yaml    # ArgoCD Application pointing to this app's chart or manifests
    values.yaml         # Helm values (if applicable)
  another-app/
    deployment.yaml
    service.yaml
    ingress.yaml
```

## Notes

- Use `cert-manager.io/cluster-issuer: letsencrypt-staging` annotations on Ingress resources while testing, then switch to `letsencrypt-prod` once validated.
- For secrets, use Sealed Secrets (`kubeseal`) to encrypt sensitive data before committing to this repo.
- Available storage classes: `longhorn` (distributed, default), `nfs-client` (NAS-backed).
