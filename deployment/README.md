# Belimang Deployment Guide

This directory contains deployment configurations for the Belimang application optimized for production environments.

## Deployment Options

### 1. Helm Charts (Production) - Recommended

- **Location**: `helm/`
- **Use case**: Production Kubernetes deployments
- **Features**: Templated configurations, auto database initialization, optimized for 60k RPS
- **Target**: 7 Core, 21GB RAM server with machinist-belimang-airmata namespace

### 2. Docker Compose (Development)

- **Location**: Root directory (`compose.dev.yaml`, `compose.yaml`)
- **Use case**: Local development and testing
- **Features**: Hot reload, easy debugging, minimal resource usage

## Directory Structure

```
deployment/
├── helm/                   # Helm charts for production deployment
│   ├── templates/          # Kubernetes manifest templates
│   ├── files/             # SQL initialization scripts
│   ├── Chart.yaml         # Helm chart metadata
│   └── values.yaml        # Configuration values
│
├── k8s/                   # Raw Kubernetes manifests (legacy)
├── k3s/                   # K3s development deployment (legacy)
├── machinist.kubeconfig   # Kubernetes cluster configuration
└── README.md              # This file
```

## Quick Start

### Helm (Production)

1. **Deploy**:

   ```bash
   make helm-install
   ```

2. **Check status**:

   ```bash
   make helm-status
   kubectl get pods -n machinist-belimang-airmata
   ```

3. **Upgrade**:

   ```bash
   make helm-upgrade
   ```

4. **Cleanup**:
   ```bash
   make helm-uninstall
   ```

## Resource Requirements

### Helm Production (7 Core, 21GB RAM)

- **Target**: Production server (7 Core, 21GB RAM)
- **Namespace**: machinist-belimang-airmata
- **PostgreSQL**: 2.5 CPU, 8Gi RAM, 50Gi storage (arfiansr/belimang-postgres with PostGIS)
- **Redis**: 1 CPU, 4Gi RAM, 20Gi storage
- **Application**: 3-4 replicas, 1.5 CPU each, 1.5Gi RAM each (HPA enabled)
- **Total**: Optimized for 60k RPS throughput

## Key Features

### Helm Production Deployment

- **Auto Database Initialization**: SQL scripts automatically executed via initdb ConfigMap
- **PostGIS Support**: Custom PostgreSQL image with PostGIS and H3 extensions
- **HPA Enabled**: Horizontal Pod Autoscaler for application scaling (3-4 replicas)
- **Resource Optimized**: Configured for 60k RPS on 7 Core, 21GB RAM server
- **Production Ready**: Optimized PostgreSQL configuration with io_uring

### Database Schema

The deployment automatically creates the following tables:

- `users` - User management
- `merchants` - Merchant information
- `items` - Product catalog
- `estimates` - Price estimates
- `orders` - Order management
- `order_merchants` - Order-merchant relationships
- `order_items` - Order line items

## Monitoring

### Check Deployment Status

```bash
# Check all resources
kubectl get all -n machinist-belimang-airmata

# Check persistent volumes
kubectl get pvc -n machinist-belimang-airmata

# Check HPA status
kubectl get hpa -n machinist-belimang-airmata
```

### View Logs

```bash
# Application logs
kubectl logs -n machinist-belimang-airmata deployment/belimang -f

# PostgreSQL logs
kubectl logs -n machinist-belimang-airmata deployment/belimang-postgresql -f

# Redis logs
kubectl logs -n machinist-belimang-airmata deployment/belimang-redis -f
```

### Health Checks

```bash
# Check application health
kubectl get pods -n machinist-belimang-airmata

# Port forward for local testing
kubectl port-forward service/belimang 8080:80 -n machinist-belimang-airmata
```

## Troubleshooting

### Common Issues

1. **Pod Not Starting**:

   - Check pod status: `kubectl get pods -n machinist-belimang-airmata`
   - Describe pod: `kubectl describe pod <pod-name> -n machinist-belimang-airmata`
   - Check events: `kubectl get events -n machinist-belimang-airmata --sort-by='.lastTimestamp'`

2. **Database Connection Issues**:

   - Verify PostgreSQL pod is running
   - Check database logs: `kubectl logs deployment/belimang-postgresql -n machinist-belimang-airmata`
   - Verify initdb scripts executed: `kubectl exec -n machinist-belimang-airmata <postgres-pod> -- ls -la /docker-entrypoint-initdb.d/`

3. **Resource Constraints**:
   - Monitor resource usage: `kubectl top pods -n machinist-belimang-airmata`
   - Check HPA status: `kubectl get hpa -n machinist-belimang-airmata`
   - Check node capacity: `kubectl describe nodes`

### Debug Commands

```bash
# Check database tables
kubectl exec -n machinist-belimang-airmata <postgres-pod> -- psql -U postgres -d belimang -c "\dt"

# Test database connection
kubectl exec -n machinist-belimang-airmata <app-pod> -- /app/server health

# Port forward for local testing
kubectl port-forward service/belimang 8080:80 -n machinist-belimang-airmata
```

## Security

Helm deployment includes:

- Non-root containers with security contexts
- Secret management for database credentials
- Resource limits and requests
- Health checks and readiness probes
- Network policies ready configuration

## Scaling

### Horizontal Pod Autoscaler (HPA)

The deployment includes HPA configuration:

```bash
# Check HPA status
kubectl get hpa -n machinist-belimang-airmata

# Manual scaling (if needed)
kubectl scale deployment belimang --replicas=4 -n machinist-belimang-airmata

# Check scaling events
kubectl describe hpa belimang -n machinist-belimang-airmata
```

### Performance Tuning

- **CPU**: Optimized for 60k RPS with 3-4 replicas
- **Memory**: 1.5Gi per replica with efficient Go memory management
- **Database**: PostgreSQL with io_uring and optimized configuration
- **Caching**: Redis for session and query result caching
