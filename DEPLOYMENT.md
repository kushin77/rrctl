# DEPLOYMENT.md

Production-ready deployment guide for rrctl.

---

## Table of Contents

1. [Local Installation](#local-installation)
2. [Docker Deployment](#docker-deployment)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Enterprise Features](#enterprise-features)
5. [Monitoring & Observability](#monitoring--observability)
6. [Security Hardening](#security-hardening)
7. [Troubleshooting](#troubleshooting)

---

## Local Installation

### Prerequisites
- Go 1.21+ (for building from source)
- Docker 20.10+ (for Ollama integration)
- Git 2.0+
- 4GB+ RAM available
- 10GB disk space

### Binary Installation (Recommended)

**Signed Releases via Sigstore:**
```bash
# Install cosign
curl -L https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64 \
  -o /usr/local/bin/cosign
chmod +x /usr/local/bin/cosign

# Download release binary
VERSION=v1.0.0
wget https://github.com/elevatediq/rrctl/releases/download/${VERSION}/rrctl-linux-amd64

# Verify signature (automatic download of public key)
cosign verify-blob \
  --signature=rrctl-linux-amd64.sig \
  --certificate=rrctl-linux-amd64.crt \
  rrctl-linux-amd64

# Or verify SBOM integrity
cosign verify-sbom --sbom=rrctl-${VERSION}.sbom.json rrctl-linux-amd64

# Install
chmod +x rrctl-linux-amd64
sudo mv rrctl-linux-amd64 /usr/local/bin/rrctl
```

**Checksums:**
```bash
# Verify SHA256
echo "abc123... rrctl-linux-amd64" | sha256sum -c -
```

### Build from Source

**Prerequisites:**
```bash
# Required for Ollama integration
docker pull ollama/ollama

# Clone repository
git clone https://github.com/elevatediq/rrctl.git
cd rrctl
```

**Build:**
```bash
# Production build
make build

# Development build with symbols
make dev-build

# Binary location
./rrctl --version
```

**Verification:**
```bash
# Test installation
./rrctl health

# Check dependencies
./rrctl version
```

---

## Docker Deployment

### Container Registry

**Public Registry (Docker Hub):**
```bash
docker pull elevatediq/rrctl:latest
docker pull elevatediq/rrctl:v1.0.0  # Pinned version
```

**Private Registry (Enterprise):**
```bash
docker tag elevatediq/rrctl:latest private-registry.example.com/rrctl:v1.0.0
docker push private-registry.example.com/rrctl:v1.0.0
```

### Single Container

**Basic Run:**
```bash
docker run \
  --rm \
  --name rrctl \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/reports:/reports \
  elevatediq/rrctl:latest \
  rca analyze /path/to/repo
```

**With Ollama Integration:**
```bash
docker run \
  --rm \
  --name rrctl \
  --network host \
  -e OLLAMA_HOST=http://localhost:11434 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/reports:/reports \
  elevatediq/rrctl:latest \
  rca analyze --ollama-model mistral /path/to/repo
```

**With Observability:**
```bash
docker run \
  --rm \
  --name rrctl \
  --network observability \
  -e OTEL_EXPORTER_JAEGER_AGENT_HOST=jaeger \
  -e OTEL_EXPORTER_JAEGER_AGENT_PORT=6831 \
  -e OTEL_SERVICE_NAME=rrctl \
  elevatediq/rrctl:latest \
  rca analyze /path/to/repo
```

### Docker Compose

**Complete Stack (Production):**
```yaml
version: '3.8'

services:
  # Distributed tracing
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "6831:6831/udp"  # Jaeger agent
      - "16686:16686"    # UI
    environment:
      COLLECTOR_ZIPKIN_HOST_PORT: ":9411"

  # Ollama model server
  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "11434:11434"
    environment:
      OLLAMA_NUM_GPU: 1  # GPU support if available
      OLLAMA_KEEP_ALIVE: "24h"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11434/api/tags"]
      interval: 10s
      timeout: 5s
      retries: 3

  # Pull Mistral model on startup
  ollama-init:
    image: ollama/ollama:latest
    depends_on:
      ollama:
        condition: service_healthy
    entrypoint: /bin/sh
    command: >
      -c "ollama pull mistral && echo 'Model ready'"
    volumes:
      - ollama_data:/root/.ollama

  # rrctl service
  rrctl:
    image: elevatediq/rrctl:latest
    depends_on:
      ollama-init:
        condition: service_completed_successfully
      jaeger:
        condition: service_started
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./repositories:/repos
      - ./reports:/reports
    environment:
      OLLAMA_HOST: http://ollama:11434
      OTEL_EXPORTER_JAEGER_AGENT_HOST: jaeger
      OTEL_EXPORTER_JAEGER_AGENT_PORT: 6831
      OTEL_SERVICE_NAME: rrctl
      LOG_LEVEL: info
    ports:
      - "8080:8080"  # API endpoint (if enabled)

volumes:
  ollama_data:

networks:
  default:
    name: rrctl-stack
```

**Deploy:**
```bash
docker-compose -f docker-compose.yml up -d
docker-compose logs -f rrctl
```

---

## Kubernetes Deployment

### Helm Chart

**Installation:**
```bash
helm repo add elevatediq https://charts.elevatediq.com
helm repo update

# Default deployment
helm install rrctl elevatediq/rrctl \
  --namespace security \
  --create-namespace

# Production deployment with monitoring
helm install rrctl elevatediq/rrctl \
  --namespace security \
  --create-namespace \
  --values production-values.yaml
```

**Production Values (production-values.yaml):**
```yaml
replicaCount: 3

image:
  repository: elevatediq/rrctl
  tag: v1.0.0
  pullPolicy: IfNotPresent

resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app
                operator: In
                values:
                  - rrctl
          topologyKey: kubernetes.io/hostname

# Ollama integration
ollama:
  enabled: true
  host: ollama-service:11434
  models:
    - mistral
    - neural-chat

# Observability
jaeger:
  enabled: true
  host: jaeger-collector:14250

# Security scanning
scanning:
  parallelism: 4
  timeout: 300s
```

**Manual Deployment:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rrctl-config
  namespace: security
data:
  config.yaml: |
    ollama:
      host: ollama:11434
      models:
        - mistral
    security:
      scan_timeout: 300
    observability:
      jaeger_host: jaeger-collector
      jaeger_port: 14250

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rrctl
  namespace: security
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rrctl
  template:
    metadata:
      labels:
        app: rrctl
    spec:
      serviceAccountName: rrctl
      containers:
      - name: rrctl
        image: elevatediq/rrctl:v1.0.0
        imagePullPolicy: IfNotPresent
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
        env:
        - name: OLLAMA_HOST
          value: "http://ollama:11434"
        - name: OTEL_EXPORTER_JAEGER_AGENT_HOST
          value: jaeger-collector
        - name: OTEL_SERVICE_NAME
          value: rrctl
        - name: LOG_LEVEL
          value: info
        volumeMounts:
        - name: config
          mountPath: /etc/rrctl
      volumes:
      - name: config
        configMap:
          name: rrctl-config

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rrctl
  namespace: security
```

**Deploy:**
```bash
kubectl apply -f deployment.yaml
kubectl logs -n security -l app=rrctl -f
```

### Monitoring

**Prometheus ServiceMonitor:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: rrctl
  namespace: security
spec:
  selector:
    matchLabels:
      app: rrctl
  endpoints:
  - port: metrics
    interval: 30s
```

---

## Enterprise Features

### Secret Management

**Using HashiCorp Vault:**
```bash
# Store credentials securely
vault kv put secret/rrctl/github \
  token=ghp_xxxxx \
  org=my-org

# Retrieve in deployment
export GITHUB_TOKEN=$(vault kv get -field=token secret/rrctl/github)
rrctl scan repo --token $GITHUB_TOKEN
```

**Using AWS Secrets Manager:**
```bash
aws secretsmanager create-secret \
  --name rrctl/github-token \
  --secret-string ghp_xxxxx

# In deployment
rrctl scan repo --token-secret arn:aws:secretsmanager:...
```

### Audit Logging

**Structured Audit Trail:**
```bash
rrctl --audit-log /var/log/rrctl/audit.log \
  --audit-json \
  scan repo
```

**Centralized Logging (ELK Stack):**
```bash
rrctl --log-format json \
  --log-output /dev/stdout \
  scan repo | \
  filebeat send --index rrctl
```

### Multi-Tenancy

**Organization Isolation:**
```bash
rrctl scan repo \
  --org my-org-1 \
  --isolation strict \
  --audit-org
```

### Role-Based Access Control (RBAC)

**Kubernetes RBAC:**
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rrctl-scanner
  namespace: security
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
  resourceNames: ["rrctl-credentials"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rrctl-scanner-binding
  namespace: security
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rrctl-scanner
subjects:
- kind: ServiceAccount
  name: rrctl
  namespace: security
```

---

## Monitoring & Observability

### Jaeger Tracing

**Access Dashboard:**
```
http://localhost:16686
```

**Query Traces:**
- Service: `rrctl`
- Operation: `rca.analyze`, `scan.security`, `git.analyze`
- Tags: `error`, `org`, `repository`

### Prometheus Metrics

**Metrics Exported:**
```
rrctl_rca_analysis_duration_ms
rrctl_rca_analysis_findings_total
rrctl_security_scan_duration_ms
rrctl_security_scan_vulnerabilities_total
rrctl_git_analysis_commits_total
rrctl_command_execution_errors_total
```

**Grafana Dashboard:**
```bash
# Import dashboard
curl -X POST http://localhost:3000/api/dashboards/db \
  -H "Authorization: Bearer $GRAFANA_TOKEN" \
  -d @grafana-rrctl-dashboard.json
```

### Health Checks

**Liveness Probe:**
```bash
rrctl health --check liveness
# Exit code 0 = healthy
```

**Readiness Probe:**
```bash
rrctl health --check readiness
# Verifies Ollama connectivity, storage access
```

**Kubernetes Probes:**
```yaml
livenessProbe:
  exec:
    command:
    - /rrctl
    - health
    - --check=liveness
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  exec:
    command:
    - /rrctl
    - health
    - --check=readiness
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## Security Hardening

### Network Security

**Network Policies:**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: rrctl-network-policy
  namespace: security
spec:
  podSelector:
    matchLabels:
      app: rrctl
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: ollama
    ports:
    - protocol: TCP
      port: 11434
  - to:
    - podSelector:
        matchLabels:
          app: jaeger
    ports:
    - protocol: TCP
      port: 14250
```

### Pod Security Standards

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: rrctl-restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'MustRunAs'
    seLinuxOptions:
      level: 's0:c123,c456'
  readOnlyRootFilesystem: true
```

### Secret Encryption

```bash
# Enable etcd encryption for Kubernetes secrets
kubeadm init --encryption-provider-config=/etc/kubernetes/encryption.yaml

# Or use external secret manager
helm install external-secrets \
  external-secrets/external-secrets \
  --namespace external-secrets-system \
  --create-namespace
```

---

## Troubleshooting

### Common Issues

**Ollama Connection Failed**
```bash
# Check Ollama service
docker ps | grep ollama
curl http://localhost:11434/api/tags

# Check environment variable
echo $OLLAMA_HOST

# Logs
docker logs ollama
```

**High Memory Usage**
```bash
# Check memory limits
docker stats rrctl

# Reduce parallelism
rrctl scan repo --parallel 2

# Reduce model size
ollama pull neural-chat  # Smaller than mistral
```

**Jaeger Metrics Not Appearing**
```bash
# Check connectivity
curl http://jaeger:6831/

# Verify environment variables
docker exec rrctl env | grep OTEL

# Enable debug logging
rrctl scan repo --log-level debug
```

**Permission Denied on Docker Socket**
```bash
# Fix socket permissions
docker exec -u root rrctl chown nobody:nogroup /var/run/docker.sock

# Or run with proper group
docker run --group-add $(getent group docker | cut -d: -f3) ...
```

### Debug Mode

```bash
# Enable all debug output
rrctl --debug scan repo

# Capture traces to file
rrctl --trace-output trace.json scan repo

# Profile memory
rrctl --mem-profile mem.prof scan repo
go tool pprof mem.prof
```

### Log Collection

```bash
# Collect logs for debugging
docker logs rrctl > rrctl.log 2>&1
docker logs ollama > ollama.log 2>&1
docker logs jaeger > jaeger.log 2>&1

# Send to support
tar -czf debug-logs.tar.gz *.log
```

---

## Performance Tuning

### CPU and Memory

**Kubernetes Resource Limits:**
```yaml
resources:
  requests:
    cpu: 500m         # Minimum
    memory: 512Mi
  limits:
    cpu: 2000m        # Maximum
    memory: 2Gi
```

**Docker Resource Limits:**
```bash
docker run \
  --cpus 2 \
  --memory 2g \
  elevatediq/rrctl:latest
```

### Parallelism

```bash
# Increase parallel scanning (use with caution)
rrctl scan repo --parallel 8

# CPU-bounded: parallelism = num_cpus
# Memory-bounded: parallelism = available_memory / 256MB
```

### Caching

```bash
# Enable result caching
rrctl scan repo \
  --cache-enabled \
  --cache-ttl 24h \
  --cache-dir /var/cache/rrctl
```

---

## Getting Help

- **Documentation**: https://docs.elevatediq.com/rrctl
- **Issues**: https://github.com/elevatediq/rrctl/issues
- **Email**: support@elevatediq.com
- **Slack**: [ElevatedIQ Community](https://community.elevatediq.com)
