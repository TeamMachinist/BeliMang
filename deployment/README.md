# Belimang Deployment

This directory contains deployment configurations for Belimang application with two deployment strategies:

## Directory Structure

```
deployment/
├── k8s/                    # Production Kubernetes deployment
│   ├── deploy.sh           # K8s deployment script
│   ├── app-deployment.yaml # Application deployment (2 replicas, 3 CPU, 4Gi RAM)
│   ├── postgres-deployment.yaml # PostgreSQL with H3 extension (4 CPU, 12Gi RAM)
│   ├── redis-deployment.yaml    # Redis deployment (2 CPU, 6Gi RAM)
│   ├── postgres-pvc.yaml   # PostgreSQL PVC (50Gi)
│   ├── redis-pvc.yaml      # Redis PVC (20Gi)
│   ├── .env.example        # Production environment template
│   └── shared files...     # Services, secrets, configmaps
│
├── k3s/                    # Local development deployment
│   ├── deploy.sh           # K3s deployment script
│   ├── app-deployment.yaml # Application deployment (1 replica, 1 CPU, 1Gi RAM)
│   ├── postgres-deployment.yaml # PostgreSQL 18 with io_uring (1 CPU, 1Gi RAM)
│   ├── redis-deployment.yaml    # Redis deployment (500m CPU, 512Mi RAM)
│   ├── postgres-pvc.yaml   # PostgreSQL PVC (2Gi)
│   ├── redis-pvc.yaml      # Redis PVC (1Gi)
│   ├── .env.example        # Development environment template
│   └── shared files...     # Services, secrets, configmaps
│
└── README.md               # This file
```

## Quick Start

### K3s (Local Development)

1. **Setup environment**:

   ```bash
   cp deployment/k3s/.env.example deployment/k3s/.env
   # Edit deployment/k3s/.env with your local settings
   ```

2. **Deploy**:

   ```bash
   make k3s-deploy
   ```

3. **Access application**:

   ```bash
   # Application will be available at the K3s LoadBalancer IP
   kubectl get services -n belimang
   ```

4. **Cleanup**:
   ```bash
   make k3s-cleanup
   ```

### K8s (Production)

1. **Setup environment**:

   ```bash
   cp deployment/k8s/.env.example deployment/k8s/.env
   # Edit deployment/k8s/.env with your production settings
   ```

2. **Build and push images**:

   ```bash
   make deploy-images REGISTRY_USER=your-dockerhub-username
   ```

3. **Deploy**:

   ```bash
   make k8s-deploy REGISTRY_USER=your-dockerhub-username
   ```

4. **Cleanup**:
   ```bash
   make k8s-cleanup
   ```

## Resource Requirements

### K3s (Local Development)

- **Target**: Local development machine
- **Total Resources**: ~2.5 CPU, ~2.5Gi RAM, ~3Gi storage
- **PostgreSQL**: 1 CPU, 1Gi RAM, 2Gi storage
- **Redis**: 500m CPU, 512Mi RAM, 1Gi storage
- **Application**: 1 CPU, 1Gi RAM, 1 replica

### K8s (Production - 7 Core, 21GB RAM)

- **Target**: Production server (7 Core, 21GB RAM)
- **Total Resources**: ~9 CPU, ~22Gi RAM, ~70Gi storage
- **PostgreSQL**: 4 CPU, 12Gi RAM, 50Gi storage
- **Redis**: 2 CPU, 6Gi RAM, 20Gi storage
- **Application**: 3 CPU, 4Gi RAM, 2 replicas

## Key Differences

| Feature            | K3s (Development)                       | K8s (Production)                            |
| ------------------ | --------------------------------------- | ------------------------------------------- |
| **Image Strategy** | Local images (`imagePullPolicy: Never`) | Registry images (`imagePullPolicy: Always`) |
| **Storage**        | local-path provisioner                  | Production storage class                    |
| **PostgreSQL**     | Standard postgres:18-alpine             | Custom postgres with H3 extension           |
| **Replicas**       | 1 (single instance)                     | 2 (high availability)                       |
| **Resources**      | Minimal (development)                   | Optimized for 7-core server                 |
| **Environment**    | Development settings                    | Production settings                         |
| **Logging**        | Debug level, simple format              | Info level, JSON format                     |

## Environment Variables

Both deployments use `.env` files with different default values:

### Common Variables

- `ENV`: Environment (development/production)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_ADDR`: Redis connection address
- `JWT_SECRET_KEY`: JWT signing key
- `GOMAXPROCS`: Go runtime CPU limit

### K3s Specific

- `K3S_NODE_IP`: K3s node IP address
- `STORAGE_CLASS=local-path`: K3s storage class

### K8s Specific

- `REGISTRY`: Docker registry URL
- `REGISTRY_USER`: Docker registry username
- `STORAGE_CLASS=fast-ssd`: Production storage class

## Monitoring

### Check Deployment Status

```bash
# For both K3s and K8s
kubectl get all -n belimang
kubectl get pvc -n belimang
```

### View Logs

```bash
# Application logs
kubectl logs -n belimang deployment/belimang-app-deployment -f

# PostgreSQL logs
kubectl logs -n belimang deployment/postgres-deployment -f

# Redis logs
kubectl logs -n belimang deployment/redis-deployment -f
```

### Health Checks

```bash
# Check application health
kubectl get pods -n belimang
curl http://<service-ip>/healthz
```

## Troubleshooting

### Common Issues

1. **Image Pull Errors (K8s)**:

   - Ensure images are pushed to registry
   - Check registry credentials
   - Verify image names in deployment files

2. **Storage Issues**:

   - Check if storage class exists
   - Verify PVC status: `kubectl get pvc -n belimang`
   - Check available storage space

3. **Resource Constraints**:

   - Monitor resource usage: `kubectl top pods -n belimang`
   - Adjust resource limits in deployment files
   - Check node capacity: `kubectl describe nodes`

4. **Health Check Failures**:
   - Check application logs
   - Verify `/healthz` endpoint is accessible
   - Adjust health check timeouts if needed

### Debug Commands

```bash
# Describe problematic pod
kubectl describe pod <pod-name> -n belimang

# Get events
kubectl get events -n belimang --sort-by='.lastTimestamp'

# Port forward for local access
kubectl port-forward service/belimang-app-service 8080:80 -n belimang
```

## Security

Both deployments include:

- Non-root containers with specific user IDs
- Security contexts with capability dropping
- Read-only root filesystems where possible
- Secret management for sensitive data
- Network policies (if supported by cluster)

## Scaling

### K3s

- Designed for single-node development
- Scaling not typically needed

### K8s

- Horizontal scaling supported:
  ```bash
  kubectl scale deployment belimang-app-deployment --replicas=4 -n belimang
  ```
- Vertical scaling by adjusting resource limits
- Database scaling requires additional configuration
