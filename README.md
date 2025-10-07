# Belimang Application

Belimang adalah aplikasi geospatial e-commerce yang dibangun dengan Go, menggunakan PostgreSQL 18 dengan ekstensi H3 untuk geospatial operations dan Redis untuk caching. Aplikasi ini mendukung multiple deployment strategies untuk development dan production.

## 🚀 Quick Start

```bash
# Setup environment untuk K3s development
cp .env.example.k3s .env

# Deploy ke K3s (local development)
make k3s-deploy

# Akses aplikasi
curl http://$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.status.loadBalancer.ingress[0].ip}')/healthz
```

## 📋 Prasyarat

- **Docker & Docker Compose** - Untuk containerization
- **K3s atau Kubernetes cluster** - Untuk deployment
- **kubectl** - Kubernetes CLI
- **Docker Hub account** (untuk K8s production) - Registry untuk images
- **Go 1.25+** (optional) - Untuk development lokal

## 🏗️ Struktur Aplikasi

```
├── cmd/                    # Entry point aplikasi
├── internal/              # Kode aplikasi internal
│   ├── app/              # Domain logic (items, merchant, purchase, user)
│   ├── config/           # Konfigurasi aplikasi
│   ├── infrastructure/   # Database, cache, dan infrastruktur
│   ├── middleware/       # HTTP middleware
│   └── pkg/              # Utilities dan packages
├── deployment/           # Deployment configurations
│   ├── k3s/             # K3s local development
│   ├── k8s/             # Kubernetes production
│   └── README.md        # Deployment documentation
├── migrations/           # Database migrations
├── seeds/               # Database seed data
├── compose.yaml         # Production Docker Compose
├── compose.dev.yaml     # Development Docker Compose
├── Dockerfile           # Production Dockerfile
└── Makefile            # Build automation
```

## ⚙️ Environment Configuration

Aplikasi menggunakan **centralized environment management** dengan file `.env` di root directory:

### Environment Files Structure

```
.env                    # Current environment (copy from examples below)
.env.example           # Base example for local development
.env.example.k3s       # K3s specific configuration
.env.example.k8s       # K8s production configuration
```

### Setup Environment

```bash
# Untuk K3s development
cp .env.example.k3s .env

# Untuk K8s production
cp .env.example.k8s .env

# Edit .env dengan nilai yang sesuai
```

**🔒 Security Note**: Pastikan mengganti nilai sensitif:

- `DB_PASSWORD` - Password database yang aman
- `JWT_SECRET_KEY` - Secret key untuk JWT (minimum 32 karakter)

Untuk panduan lengkap, lihat [Deployment Documentation](deployment/README.md)

## 🛠️ Deployment Options

### 1. K3s (Local Development) - Recommended

```bash
# Setup environment
cp .env.example.k3s .env

# Deploy
make k3s-deploy

# Cleanup
make k3s-cleanup
```

### 2. K8s (Production)

```bash
# Setup environment
cp .env.example.k8s .env

# Build dan push images
make deploy-images REGISTRY_USER=your-dockerhub-username

# Deploy
make k8s-deploy REGISTRY_USER=your-dockerhub-username

# Cleanup
make k8s-cleanup
```

### 3. Docker Compose (Alternative)

```bash
# Development
make up-dev-build
make down-dev

# Production
make up-prod-build
make down-prod
```

## 📊 Resource Allocation

### K3s (Local Development)

- **PostgreSQL**: 1 CPU, 1Gi RAM, 2Gi storage
- **Redis**: 500m CPU, 512Mi RAM, 1Gi storage
- **App**: 1 CPU, 1Gi RAM, 1 replica

### K8s (Production - 7 Core, 21GB RAM)

- **PostgreSQL**: 4 CPU, 12Gi RAM, 50Gi storage (dengan io_uring)
- **Redis**: 2 CPU, 6Gi RAM, 20Gi storage
- **App**: 3 CPU, 4Gi RAM, 2 replicas

## 🔧 Key Features

- **Geospatial Operations**: PostGIS dengan H3 indexing untuk location-based queries
- **Full-Text Search**: pg_trgm untuk optimized ILIKE operations
- **Complex Queries**: Multi-table JOINs dengan distance calculations
- **High Performance**: PostgreSQL 18 dengan io_uring untuk production
- **Caching Layer**: Redis untuk session dan query result caching
- **Scalable Architecture**: Horizontal scaling dengan Kubernetes

## 🔍 Monitoring & Management

### Status Monitoring

```bash
# Cek status deployment
kubectl get all -n belimang

# Cek resource usage
kubectl top pods -n belimang

# Cek logs aplikasi
kubectl logs -f deployment/belimang-app-deployment -n belimang
```

### Health Checks

```bash
# Test aplikasi health
curl http://$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.status.loadBalancer.ingress[0].ip}')/healthz

# Port forward untuk akses lokal
kubectl port-forward service/belimang-app-service 8080:80 -n belimang
```

### Scaling Operations

```bash
# Scale aplikasi
kubectl scale deployment belimang-app-deployment --replicas=3 -n belimang

# Update aplikasi
kubectl set image deployment/belimang-app-deployment belimang-app=belimang-app:v2 -n belimang

# Rollback jika diperlukan
kubectl rollout undo deployment/belimang-app-deployment -n belimang
```

### Monitoring Kubernetes Deployment

```bash
# Cek status pods
kubectl get pods -n belimang

# Cek status services
kubectl get services -n belimang

# Cek logs aplikasi
kubectl logs -f deployment/belimang-app-deployment -n belimang

# Cek logs PostgreSQL
kubectl logs -f deployment/postgres-deployment -n belimang

# Cek logs Redis
kubectl logs -f deployment/redis-deployment -n belimang

# Describe pod untuk troubleshooting
kubectl describe pod <pod-name> -n belimang
```

### Akses Aplikasi Kubernetes

```bash
# Port forward untuk akses lokal
kubectl port-forward service/belimang-app-service 8080:80 -n belimang

# Atau dapatkan external IP (jika LoadBalancer tersedia)
kubectl get service belimang-app-service -n belimang
```

### Scaling Aplikasi

```bash
# Scale aplikasi ke 3 replicas
kubectl scale deployment belimang-app-deployment --replicas=3 -n belimang

# Cek status scaling
kubectl get deployment belimang-app-deployment -n belimang
```

### Update Aplikasi

```bash
# Update image aplikasi
kubectl set image deployment/belimang-app-deployment belimang-app=belimang-app:v2 -n belimang

# Rollback jika diperlukan
kubectl rollout undo deployment/belimang-app-deployment -n belimang

# Cek status rollout
kubectl rollout status deployment/belimang-app-deployment -n belimang
```

## Troubleshooting

### Docker Compose Issues

1. **Port sudah digunakan**

```bash
# Cek port yang digunakan
netstat -tulpn | grep :8080

# Ganti port di .env file
HTTP_PORT=8081
```

2. **Database connection error**

```bash
# Cek logs PostgreSQL
docker-compose logs postgres

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

### Kubernetes Issues

1. **Pod tidak bisa start**

```bash
# Cek events
kubectl get events -n belimang --sort-by='.lastTimestamp'

# Cek pod details
kubectl describe pod <pod-name> -n belimang
```

2. **Image pull error**

```bash
# Pastikan image sudah di-load
docker images | grep belimang

# Load ulang image
minikube image load belimang-app:latest
```

3. **PVC tidak bisa mount**

```bash
# Cek storage class
kubectl get storageclass

# Cek PVC status
kubectl get pvc -n belimang
```

## Security Notes

### Production Deployment

1. **Ganti ConfigMap Values**: Update semua nilai sensitif di `k8s/configmap.yaml` dengan nilai yang aman
2. **Network Policies**: Implementasikan network policies untuk isolasi
3. **RBAC**: Setup Role-Based Access Control
4. **TLS**: Gunakan TLS untuk komunikasi eksternal
5. **Image Security**: Scan images untuk vulnerabilities
6. **Registry Security**: Gunakan private registry untuk production images

### Environment Variables

Pastikan untuk mengganti nilai-nilai berikut di production:

- `JWT_SECRET_KEY`: Gunakan secret key yang kuat
- `DB_PASSWORD`: Gunakan password yang kompleks
- Database credentials lainnya

## Performance Tuning

### PostgreSQL

Konfigurasi PostgreSQL sudah dioptimasi untuk performa:

- `shared_buffers=256MB`
- `effective_cache_size=1GB`
- `work_mem=2MB`
- `max_connections=200`

### Redis

Konfigurasi Redis dengan:

- `maxmemory=512mb`
- `maxmemory-policy=allkeys-lru`
- Persistence dengan AOF dan RDB

### Aplikasi Go

- `GOMAXPROCS=4`: Sesuaikan dengan CPU cores
- `GOMEMLIMIT=1536MiB`: Memory limit untuk GC
- CGO enabled untuk performa database

## Monitoring dan Logging

### Logs

```bash
# Docker Compose
docker-compose logs -f [service-name]

# Kubernetes
kubectl logs -f deployment/[deployment-name] -n belimang
```

### Health Checks

Semua services memiliki health checks:

- **PostgreSQL**: `pg_isready`
- **Redis**: `redis-cli ping`
- **Aplikasi**: `/app/server health`

### Metrics

Untuk monitoring production, pertimbangkan untuk menambahkan:

- Prometheus untuk metrics
- Grafana untuk visualization
- Jaeger untuk tracing
- ELK stack untuk centralized logging
