# GCP KMS / Workload Identity Operator Checklist

This checklist describes the minimum steps and example commands to configure Google Cloud KMS for cosign signing from GitHub Actions using Workload Identity Federation (OIDC).

Folder: .github/workflows/sbom-and-kms-sign-gcp.yml

Prerequisites (GCP):
- A GCP project (PROJECT)
- Owner/Org admin who can create Workload Identity Pool and Service Account
- KMS key (asymmetric signing key) created with purpose=ASYMMETRIC_SIGN

High-level steps
1) Create an asymmetric KMS key (purpose=ASYMMETRIC_SIGN) in a KeyRing
2) Create or identify a service account that will perform signing
3) Configure Workload Identity Pool / Provider to allow GitHub OIDC to impersonate the SA
4) Bind the service account to allow Workload Identity impersonation
5) Bind the service account to the KMS key with roles/cloudkms.signerVerifier
6) Add repo secrets in GitHub (see below)
7) Trigger the `sbom-and-kms-sign-gcp.yml` workflow on a release (or workflow_dispatch) to sign

Gcloud example commands (replace PROJECT, REGION, KEYRING, KEY, POOL, PROVIDER, SA_EMAIL):

```bash
# 1) Create keyring and asymmetric key
gcloud kms keyrings create my-keyring --location=global --project=$PROJECT || true
gcloud kms keys create my-key --location=global --keyring=my-keyring --purpose=asymmetric-sign --default-algorithm=ec-sign-p256-sha256 --project=$PROJECT || true

# 2) Create signer service account
gcloud iam service-accounts create signer --project=$PROJECT --display-name "CI signing service account"
SA_EMAIL=signer@${PROJECT}.iam.gserviceaccount.com

# 3) Create workload identity pool & provider (or use existing configured pool for GitHub)
gcloud iam workload-identity-pools create github-pool --project=$PROJECT --location="global" --display-name="GitHub Actions pool"

gcloud iam workload-identity-pools providers create-oidc github-provider \
  --project=$PROJECT --location="global" \
  --workload-identity-pool=github-pool \
  --display-name="GitHub OIDC" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --allowed-audiences="https://github.com/kushin77/rrctl" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor"

# 4) Grant the Workload Provider permission to impersonate the service account
gcloud iam service-accounts add-iam-policy-binding $SA_EMAIL \
  --project=$PROJECT \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://projects/${PROJECT}/locations/global/workloadIdentityPools/github-pool/attribute.actor/*"

# 5) Grant the service account permission to sign with the KMS key
KMS_KEY=projects/${PROJECT}/locations/global/keyRings/my-keyring/cryptoKeys/my-key
gcloud kms keys add-iam-policy-binding "$KMS_KEY" \
  --member="serviceAccount:$SA_EMAIL" \
  --role="roles/cloudkms.signerVerifier"

# 6) GitHub secrets to set in repository settings
# - GCP_WIF_PROVIDER: full resource name of provider (projects/PROJECT/locations/global/workloadIdentityPools/github-pool/providers/github-provider)
# - GCP_SA_EMAIL: signer@PROJECT.iam.gserviceaccount.com
# - GCP_KMS_KEY: projects/PROJECT/locations/global/keyRings/my-keyring/cryptoKeys/my-key

# 7) Trigger a release or run the workflow manually in GitHub Actions
```

Verification and local testing
- Use cosign to verify signed artifacts: `cosign verify --keyless --rekor-url https://rekor.sigstore.dev artifact.tar.gz` (for Rekor-anchored keyless)
- For KMS-signed artifacts (GCP): `cosign verify --key <public key file> artifact.tar.gz` or retrieve the public key from the KMS key and use `cosign verify --key`.

Notes & best-practices
- Use Workload Identity (OIDC) rather than long-lived JSON keys when possible.
- Restrict the signer service account to only the KMS key using IAM conditions if available.
- Use least privilege and enable audit logging for the service account and KMS key operations.
- Keep a copy of the public key (or Rekor log entry) in the release notes for easier verification by consumers.
