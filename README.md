# Belimang Application

Belimang adalah aplikasi geospatial e-commerce yang dibangun dengan Go, menggunakan PostgreSQL 18 dengan ekstensi PostGIS dan H3 untuk geospatial operations dan Redis untuk caching. Aplikasi ini menggunakan Helm untuk deployment production.

## 🚀 Quick Start

```bash
# Deploy production environment dengan Helm
make helm-install

# Akses aplikasi
kubectl get pods -n machinist-belimang-airmata
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
│   ├── helm/            # Helm charts untuk production
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

### 1. Helm (Production) - Recommended

```bash
# Deploy production environment
make helm-install

# Upgrade deployment
make helm-upgrade

# Uninstall
make helm-uninstall

# Check status
make helm-status
```

### 2. Docker Compose (Development)

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
kubectl get all -n machinist-belimang-airmata

# Cek resource usage
kubectl top pods -n machinist-belimang-airmata

# Cek logs aplikasi
kubectl logs -f deployment/belimang -n machinist-belimang-airmata
```

### Health Checks

```bash
# Test aplikasi health
kubectl get pods -n machinist-belimang-airmata

# Port forward untuk akses lokal
kubectl port-forward service/belimang 8080:80 -n machinist-belimang-airmata
```

### Scaling Operations

```bash
# Scale aplikasi
kubectl scale deployment belimang --replicas=4 -n machinist-belimang-airmata

# Update aplikasi
kubectl set image deployment/belimang belimang=arfiansr/belimang-app:v2 -n machinist-belimang-airmata

# Rollback jika diperlukan
kubectl rollout undo deployment/belimang -n machinist-belimang-airmata
```

### Database Management

```bash
# Cek tabel database
kubectl exec -n machinist-belimang-airmata belimang-postgresql-xxx -- psql -U postgres -d belimang -c "\dt"

# Cek logs PostgreSQL
kubectl logs -f deployment/belimang-postgresql -n machinist-belimang-airmata

# Cek logs Redis
kubectl logs -f deployment/belimang-redis -n machinist-belimang-airmata
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
