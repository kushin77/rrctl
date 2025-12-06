# Delivery notes: production signing and SBOM verification

This document provides recommended patterns for signing release artifacts and automating SBOM generation/attachment in CI.

1) Key management options
-------------------------

- KMS-backed cosign keys (recommended): store private keys in cloud KMS (GCP KMS, AWS KMS, Azure Key Vault) and use cosign's `sign` command with `--kms` to sign artifacts from CI runners without exposing private key material.
  - GCP example: `cosign sign --key gcpkms://projects/PROJECT/locations/global/keyRings/RING/cryptoKeys/KEY@latest artifact`
  - AWS example: `cosign sign --key awskms://arn:aws:kms:REGION:ACCOUNT:key/KEYID artifact`

- Keyless (OIDC + Fulcio + Rekor): Good for ephemeral runner scenarios and public provenance. Ensure Rekor uploads succeed and configure retries/backoff.

2) Sample GitHub Actions patterns
--------------------------------

- Workflow triggers: run on `release` (types: published) and `workflow_dispatch` for manual runs.
- Steps:
  1. Build / package artifact (tarball, binaries)
  2. Generate SBOM with syft: `syft <artifact> -o spdx-json=artifact.spdx.json`
  3. Create checksums: `sha256sum artifact > artifact.sha256`
  4. Sign artifact:
     - KMS: `cosign sign --key awskms://... artifact`
     - Keyless: `cosign sign-blob --keyless --output-signature signature.b64 artifact`
  5. Upload artifacts: `actions/upload-artifact` and `gh release upload` (for attaching to releases)

3) Rekor resilience
---------------------

If runners have intermittent outbound network issues, consider:

- Configure cosign with `--rekor-url` to use an alternate Rekor instance
- Add retry loops around `cosign` Rekor uploads; detect transient HTTP 5xx/timeouts and retry with exponential backoff
- As fallback, create and upload the signature file and public key to the release but **note** that signatures without Rekor anchoring are less trustworthy

4) Verification steps for downstream consumers
---------------------------------------------

1. Check checksums: `sha256sum -c <(echo "<sha256>  artifact")`
2. Inspect SBOM with `syft`/`jq`/`spdx` viewers to verify included packages and licenses
3. Verify signature (keyless with Rekor anchored): `cosign verify --keyless --rekor-url https://rekor.sigstore.dev artifact`
4. Verify signature (KMS public key): `cosign verify --key <public-key> artifact`

5) Automation: attach workflow artifacts to releases
---------------------------------------------------

Use a final workflow step (after successful sign) to attach artifacts to the GitHub release for the tag being published. Example (if running on `release` event):

```yaml
- name: Attach artifacts to release
  uses: softprops/action-gh-release@v1
  with:
    files: |
      artifact.tar.gz
      artifact.spdx.json
      artifact.sha256
      signature.b64
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

6) Next steps for this project
------------------------------

- Decide preferred signing strategy (KMS provider or keyless Rekor). If KMS, provide an example using your chosen cloud provider and configure CI secrets for minimal privileges. If keyless, ensure OIDC is enabled for the GitHub Actions runner and monitor Rekor availability.
- Add a workflow job to automatically attach artifacts to the release on publish.
- Add monitoring/alerts for failed Rekor uploads so signatures are re-attempted.

If you want, I can implement a KMS-backed signing variant of the current workflow for GCP/AWS/Azure and a release-attachment step—tell me your preferred KMS provider and I will implement, test, and submit a PR.
