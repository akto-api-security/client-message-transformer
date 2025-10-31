# Helm Installation Guide

This guide provides instructions for installing the Client Message Transformer service using Helm.

## Prerequisites

- Kubernetes cluster (v1.19+)
- Helm 3.x installed
- kubectl configured to access your cluster
- Access to Kafka brokers (source and destination)

## Quick Start

### Basic Installation

```bash
helm install transformer ./helm/transformer
```

### Installation with Custom Release Name

```bash
helm install my-transformer ./helm/transformer
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `public.ecr.aws/aktosecurity/client-message-transformer` |
| `image.tag` | Container image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `kafka.source.brokers` | Source Kafka broker addresses | `localhost:9092` |
| `kafka.source.topic` | Source Kafka topic | `client-messages` |
| `kafka.source.sasl.enabled` | Enable SASL authentication for source | `false` |
| `kafka.source.sasl.mechanism` | SASL mechanism (PLAIN, SCRAM-SHA-256, SCRAM-SHA-512) | `PLAIN` |
| `kafka.source.sasl.username` | SASL username for source | `""` |
| `kafka.source.sasl.password` | SASL password for source | `""` |
| `kafka.source.sasl.securityProtocol` | Security protocol (SASL_PLAINTEXT, SASL_SSL) | `SASL_PLAINTEXT` |
| `kafka.destination.brokers` | Destination Kafka broker addresses | `localhost:9093` |
| `kafka.destination.topic` | Destination Kafka topic | `akto.api.logs` |
| `kafka.destination.sasl.enabled` | Enable SASL authentication for destination | `false` |
| `kafka.destination.sasl.mechanism` | SASL mechanism for destination | `PLAIN` |
| `kafka.destination.sasl.username` | SASL username for destination | `""` |
| `kafka.destination.sasl.password` | SASL password for destination | `""` |
| `kafka.destination.sasl.securityProtocol` | Security protocol for destination | `SASL_PLAINTEXT` |
| `kafka.consumerGroup` | Kafka consumer group ID | `message-transformer-group` |
| `client.id` | Client ID | `1000000` |
| `logging.level` | Logging level | `INFO` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `250m` |
| `resources.requests.memory` | Memory request | `256Mi` |
| `autoscaling.enabled` | Enable horizontal pod autoscaling | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `3` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization | `80` |

## Installation Methods

### Method 1: Using --set Flag

Override individual values using the `--set` flag:

```bash
helm install transformer ./helm/transformer \
  --set kafka.source.brokers="kafka-broker-1:9092,kafka-broker-2:9092" \
  --set kafka.destination.brokers="kafka-dest:9092" \
  --set kafka.source.topic="my-source-topic" \
  --set kafka.destination.topic="my-dest-topic" \
  --set client.id="2000000"
```

### Method 2: Using Custom Values File

Create a custom values file (e.g., `my-values.yaml`):

```yaml
replicaCount: 2

kafka:
  source:
    brokers: "kafka-source-1:9092,kafka-source-2:9092"
    topic: "client-messages"
  destination:
    brokers: "kafka-dest-1:9092,kafka-dest-2:9092"
    topic: "akto.api.logs"
  consumerGroup: "my-transformer-group"

client:
  id: "2000000"

logging:
  level: "DEBUG"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

Install using the custom values file:

```bash
helm install transformer ./helm/transformer -f my-values.yaml
```

### Method 3: With SASL Authentication

For Kafka clusters requiring SASL authentication:

```yaml
# sasl-values.yaml
kafka:
  source:
    brokers: "kafka-source:9092"
    topic: "client-messages"
    sasl:
      enabled: true
      mechanism: "SCRAM-SHA-256"
      username: "source-user"
      password: "source-password"
      securityProtocol: "SASL_SSL"

  destination:
    brokers: "kafka-dest:9092"
    topic: "akto.api.logs"
    sasl:
      enabled: false
      mechanism: "SCRAM-SHA-256"
      username: "dest-user"
      password: "dest-password"
      securityProtocol: "SASL_SSL"
```

Install with SASL configuration:

```bash
helm install transformer ./helm/transformer -f sasl-values.yaml
```

### Method 4: Using Environment-Specific Configurations

```bash
# Development
helm install transformer ./helm/transformer -f values-dev.yaml

# Staging
helm install transformer ./helm/transformer -f values-staging.yaml

# Production
helm install transformer ./helm/transformer -f values-prod.yaml
```

## Upgrade

To upgrade an existing release with new values:

```bash
helm upgrade transformer ./helm/transformer -f my-values.yaml
```

## Uninstall

To remove the deployment:

```bash
helm uninstall transformer
```

## Verify Installation

Check the deployment status:

```bash
# Check Helm release
helm list

# Check pods
kubectl get pods -l app.kubernetes.io/name=transformer

# Check logs
kubectl logs -l app.kubernetes.io/name=transformer -f

# Get service account
kubectl get serviceaccount transformer
```

## Autoscaling

To enable horizontal pod autoscaling:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

Install with autoscaling:

```bash
helm install transformer ./helm/transformer \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=10
```

## Security Considerations

The chart includes security best practices:
- Runs as non-root user (UID 1000)
- Drops all capabilities
- Read-only root filesystem
- No privilege escalation

## Troubleshooting

### Check Pod Status
```bash
kubectl describe pod -l app.kubernetes.io/name=transformer
```

### View Logs
```bash
kubectl logs -l app.kubernetes.io/name=transformer --tail=100
```

### Check Configuration
```bash
helm get values transformer
```

### Verify Kafka Connectivity
```bash
kubectl exec -it deployment/transformer -- sh
# Test network connectivity to Kafka brokers
```

## Advanced Configuration Examples

### High Availability Setup

```yaml
replicaCount: 3

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                  - transformer
          topologyKey: kubernetes.io/hostname
```

### Node-Specific Deployment

```yaml
nodeSelector:
  workload-type: message-processing

tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "kafka-processing"
    effect: "NoSchedule"
```

## Support

For issues and questions:
- Check logs: `kubectl logs -l app.kubernetes.io/name=transformer`
- Review configuration: `helm get values transformer`
- Verify Kafka connectivity and authentication
