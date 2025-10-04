# Belimang Application

Aplikasi Belimang adalah sebuah aplikasi Go yang menggunakan PostgreSQL dengan ekstensi H3 dan Redis sebagai cache. Aplikasi ini dapat dijalankan dalam tiga cara berbeda: development mode, production mode dengan Docker Compose, dan deployment di Kubernetes dengan Docker registry.

## Prasyarat

- Docker dan Docker Compose
- Kubernetes cluster (untuk deployment K8s)
- kubectl (untuk deployment K8s)
- Docker Hub account atau registry lainnya (untuk K8s deployment)
- Go 1.25+ (untuk development lokal)

## Struktur Aplikasi

```
├── cmd/                    # Entry point aplikasi
├── internal/              # Kode aplikasi internal
│   ├── app/              # Domain logic (items, merchant, purchase, user)
│   ├── config/           # Konfigurasi aplikasi
│   ├── infrastructure/   # Database, cache, dan infrastruktur
│   ├── middleware/       # HTTP middleware
│   └── pkg/              # Utilities dan packages
├── k8s/                  # Kubernetes manifests
├── migrations/           # Database migrations
├── seeds/               # Database seed data
├── compose.yaml         # Production Docker Compose
├── compose.dev.yaml     # Development Docker Compose
├── Dockerfile           # Production Dockerfile
└── Makefile            # Build automation
```

## Konfigurasi Environment

Salin file `.env.example` ke `.env` dan sesuaikan konfigurasi:

```bash
cp .env.example .env
```

Edit file `.env` sesuai kebutuhan:

```env
# Environment Configuration
ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5000
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=belimang
DB_SSLMODE=disable
DATABASE_URL=postgres://postgres:password@postgres:5432/belimang?sslmode=disable

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# Cache Configuration
CACHE_HOST=localhost
CACHE_PORT=6000
CACHE_PASSWORD=
CACHE_DB=0
REDIS_ADDR=redis:6379

# Logger Configuration
LOG_LEVEL=info
LOG_TYPE=simple

# JWT Configuration
JWT_SECRET_KEY=your-secret-key-change-in-production
JWT_ISSUER=belimang-app
```

## Quick Start dengan Makefile

Gunakan Makefile untuk operasi yang lebih mudah:

```bash
# Lihat semua perintah yang tersedia
make help

# Development
make up-dev-build    # Build dan start development environment
make down-dev        # Stop development environment

# Production
make up-prod-build   # Build dan start production environment
make down-prod       # Stop production environment

# Docker Registry & Kubernetes
make deploy-images REGISTRY_USER=your-username  # Build, push ke registry
make k8s-deploy REGISTRY_USER=your-username     # Deploy ke Kubernetes

# K3s (Lightweight Kubernetes)
make k3s-deploy      # Deploy ke K3s menggunakan local images
make k3s-cleanup     # Cleanup K3s deployment
```

## 1. Development Mode (compose.dev.yaml)

Mode development menggunakan hot reload dan volume mounting untuk development yang lebih cepat.

### Menjalankan Development Mode

```bash
# Build dan jalankan semua services
docker-compose -f compose.dev.yaml up --build

# Atau jalankan di background
docker-compose -f compose.dev.yaml up -d --build

# Melihat logs
docker-compose -f compose.dev.yaml logs -f

# Stop services
docker-compose -f compose.dev.yaml down

# Stop dan hapus volumes
docker-compose -f compose.dev.yaml down -v
```

### Fitur Development Mode

- **Hot Reload**: Perubahan kode akan otomatis ter-reload
- **Volume Mounting**: Source code di-mount ke container
- **Debug Mode**: Logging lebih verbose
- **Fast Build**: Build time lebih cepat untuk development

### Akses Aplikasi Development

- **Aplikasi**: http://localhost:8080
- **PostgreSQL**: localhost:5000
- **Redis**: localhost:6000

## 2. Production Mode (compose.yaml)

Mode production dengan optimasi performa dan keamanan.

### Menjalankan Production Mode

```bash
# Build dan jalankan semua services
docker-compose -f compose.yaml up --build

# Atau jalankan di background
docker-compose -f compose.yaml up -d --build

# Melihat logs
docker-compose -f compose.yaml logs -f

# Stop services
docker-compose -f compose.yaml down

# Stop dan hapus volumes
docker-compose -f compose.yaml down -v
```

### Fitur Production Mode

- **Optimized Build**: Binary yang dioptimasi untuk production
- **Security**: Non-root user, security constraints
- **Performance Tuning**: PostgreSQL dan Redis dengan konfigurasi optimal
- **Health Checks**: Health checks untuk semua services
- **Resource Limits**: CPU dan memory limits
- **Restart Policies**: Automatic restart on failure

### Akses Aplikasi Production

- **Aplikasi**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## 3. Kubernetes Deployment dengan Docker Registry

Deployment di Kubernetes cluster menggunakan Docker registry untuk production-ready deployment dengan high availability dan scalability.

### Workflow Docker Registry

Kubernetes deployment menggunakan images dari Docker registry (Docker Hub, AWS ECR, Google GCR, dll.) untuk memastikan konsistensi dan kemudahan deployment.

#### 1. Setup Docker Registry

```bash
# Login ke Docker Hub (atau registry lainnya)
make docker-login

# Atau manual
docker login docker.io
# Masukkan username dan password Docker Hub Anda
```

#### 2. Build dan Push Images ke Registry

```bash
# Build dan push dengan Makefile (recommended)
make deploy-images REGISTRY_USER=your-dockerhub-username

# Atau manual
make build-images REGISTRY_USER=your-dockerhub-username
make push-images REGISTRY_USER=your-dockerhub-username
```

Perintah ini akan:

- Build image PostgreSQL dengan H3 extension
- Build image aplikasi Go
- Push kedua images ke registry dengan tag `latest`

#### 3. Deploy ke Kubernetes

```bash
# Deploy otomatis dengan registry images
make k8s-deploy REGISTRY_USER=your-dockerhub-username

# Atau manual
cd k8s
./deploy.sh deploy
```

### Konfigurasi Environment Variables

Kubernetes menggunakan ConfigMap untuk environment variables (bukan Secrets). Edit `k8s/configmap.yaml` untuk menyesuaikan konfigurasi:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: belimang-config
  namespace: belimang
data:
  ENV: "production"
  DB_USER: "postgres"
  DB_PASSWORD: "your-secure-password" # Ganti dengan password yang aman
  JWT_SECRET_KEY: "your-jwt-secret" # Ganti dengan JWT secret yang aman
  # ... konfigurasi lainnya
```

### Deployment Manual (Step by Step)

Jika ingin deploy manual tanpa Makefile:

1. **Update Path di postgres-deployment.yaml**

Edit file `k8s/postgres-deployment.yaml` dan ganti path berikut dengan path absolut ke project Anda:

```yaml
volumes:
  - name: migrations
    hostPath:
      path: /path/to/your/project/migrations # Ganti dengan path absolut
      type: Directory
  - name: seeds
    hostPath:
      path: /path/to/your/project/seeds # Ganti dengan path absolut
      type: Directory
```

2. **Update Images di Deployment Files**

Edit `k8s/postgres-deployment.yaml` dan `k8s/app-deployment.yaml`:

```yaml
# postgres-deployment.yaml
spec:
  template:
    spec:
      containers:
      - name: postgres
        image: your-dockerhub-username/belimang-postgres:latest
        imagePullPolicy: Always

# app-deployment.yaml
spec:
  template:
    spec:
      containers:
      - name: belimang-app
        image: your-dockerhub-username/belimang-app:latest
        imagePullPolicy: Always
```

3. **Deploy Semua Komponen**

```bash
# Deploy namespace
kubectl apply -f k8s/namespace.yaml

# Deploy ConfigMap
kubectl apply -f k8s/configmap.yaml

# Deploy PersistentVolumeClaims
kubectl apply -f k8s/postgres-pvc.yaml
kubectl apply -f k8s/redis-pvc.yaml

# Deploy PostgreSQL
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/postgres-service.yaml

# Deploy Redis
kubectl apply -f k8s/redis-deployment.yaml
kubectl apply -f k8s/redis-service.yaml

# Deploy Aplikasi
kubectl apply -f k8s/app-deployment.yaml
kubectl apply -f k8s/app-service.yaml
```

### Development vs Production Deployment

#### Development (Local Images)

Untuk development lokal dengan minikube/kind:

```bash
# Build images lokal
docker build -t belimang-postgres:latest --target postgres-h3 .
docker build -t belimang-app:latest --target production .

# Load ke minikube/kind
minikube image load belimang-postgres:latest belimang-app:latest
# atau
kind load docker-image belimang-postgres:latest belimang-app:latest

# Deploy dengan images lokal
cd k8s && ./deploy.sh deploy
```

#### Production (Registry Images)

Untuk production dengan registry:

```bash
# Build dan push ke registry
make deploy-images REGISTRY_USER=your-username

# Deploy dengan registry images
make k8s-deploy REGISTRY_USER=your-username
```

## 4. K3s Deployment (Lightweight Kubernetes)

K3s adalah distribusi Kubernetes yang ringan, sempurna untuk development, edge computing, dan resource-constrained environments. K3s deployment menggunakan local images dan konfigurasi yang dioptimasi untuk single-node atau small cluster.

### Prasyarat K3s

```bash
# Install K3s
curl -sfL https://get.k3s.io | sh -

# Atau dengan konfigurasi khusus
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--bind-address=0.0.0.0 --disable=traefik" sh -

# Cek status K3s
sudo systemctl status k3s

# Setup kubectl untuk user biasa
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $USER:$USER ~/.kube/config
```

### Konfigurasi Environment K3s

1. **Setup Environment Variables**

```bash
# Copy dan edit file environment untuk K3s
cd k8s
cp .env.example .env
```

Edit file `k8s/.env` sesuai kebutuhan:

```env
# Application Environment
ENV=production
GIN_MODE=release
HTTP_PORT=8080

# Database Configuration
DB_HOST=postgres-service
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=change-this-secure-password
DB_NAME=belimang
DATABASE_URL=postgres://postgres:change-this-secure-password@postgres-service:5432/belimang?sslmode=disable

# JWT Configuration
JWT_SECRET_KEY=change-this-jwt-secret-key-in-production

# K3s Specific Configuration
K3S_NODE_IP=192.168.1.100
STORAGE_CLASS=local-path
POSTGRES_STORAGE_SIZE=10Gi
REDIS_STORAGE_SIZE=5Gi

# Resource Configuration
APP_REPLICAS=2
POSTGRES_CPU_LIMIT=2
POSTGRES_MEMORY_LIMIT=2Gi
APP_CPU_LIMIT=4
APP_MEMORY_LIMIT=2Gi
```

### Deploy ke K3s

#### Quick Deploy dengan Makefile

```bash
# Deploy ke K3s (otomatis build images jika diperlukan)
make k3s-deploy

# Cleanup deployment
make k3s-cleanup
```

#### Manual Deploy dengan Script

```bash
# Masuk ke directory k8s
cd k8s

# Deploy dengan script K3s
./deploy-k3s.sh deploy

# Cek status deployment
./deploy-k3s.sh status

# Cleanup deployment
./deploy-k3s.sh cleanup
```

### Fitur K3s Deployment

#### Automatic Image Import

Script K3s otomatis:

- Build Docker images jika belum ada
- Import images ke K3s containerd runtime
- Generate ConfigMap dari file `.env`
- Update resource limits sesuai konfigurasi

#### Local Path Storage

K3s menggunakan `local-path` storage class yang:

- Menyimpan data di node lokal
- Tidak memerlukan external storage provider
- Cocok untuk development dan testing

#### Optimized Resource Usage

- Resource limits disesuaikan dengan environment
- Support untuk single-node deployment
- Minimal overhead dibanding full Kubernetes

### Akses Aplikasi K3s

```bash
# Cek status deployment
kubectl get pods -n belimang
kubectl get services -n belimang

# Port forward untuk akses lokal
kubectl port-forward service/belimang-app-service 8080:80 -n belimang

# Akses aplikasi
curl http://localhost:8080

# Jika menggunakan LoadBalancer (K3s servicelb)
kubectl get service belimang-app-service -n belimang
# Akses via node IP dan port
```

### K3s vs Standard Kubernetes

| Feature            | K3s                    | Standard K8s            |
| ------------------ | ---------------------- | ----------------------- |
| **Installation**   | Single binary, 5 menit | Complex, 30+ menit      |
| **Memory Usage**   | ~512MB                 | ~2GB+                   |
| **Storage**        | Built-in local-path    | Perlu external provider |
| **Load Balancer**  | Built-in servicelb     | Perlu external LB       |
| **Image Registry** | Local containerd       | Perlu registry setup    |
| **Use Case**       | Dev, Edge, IoT         | Production, Enterprise  |

### Troubleshooting K3s

#### 1. K3s Service Issues

```bash
# Cek status K3s
sudo systemctl status k3s

# Restart K3s
sudo systemctl restart k3s

# Cek logs K3s
sudo journalctl -u k3s -f
```

#### 2. Image Import Issues

```bash
# Manual import images
docker save belimang-postgres:latest | sudo k3s ctr images import -
docker save belimang-app:latest | sudo k3s ctr images import -

# List images di K3s
sudo k3s ctr images list | grep belimang
```

#### 3. Storage Issues

```bash
# Cek storage class
kubectl get storageclass

# Cek PVC status
kubectl get pvc -n belimang

# Cek local-path provisioner
kubectl get pods -n local-path-storage
```

#### 4. Network Access Issues

```bash
# Cek node IP
kubectl get nodes -o wide

# Cek service endpoints
kubectl get endpoints -n belimang

# Test internal connectivity
kubectl run test-pod --image=busybox -it --rm -- nslookup belimang-app-service.belimang.svc.cluster.local
```

### K3s Production Considerations

#### Security Hardening

```bash
# Install dengan security options
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--protect-kernel-defaults --secrets-encryption" sh -

# Setup firewall rules
sudo ufw allow 6443/tcp  # K3s API server
sudo ufw allow 10250/tcp # Kubelet
```

#### High Availability K3s

```bash
# Setup K3s cluster dengan external database
curl -sfL https://get.k3s.io | K3S_DATASTORE_ENDPOINT="mysql://user:pass@tcp(host:3306)/k3s" sh -s - server

# Join additional nodes
curl -sfL https://get.k3s.io | K3S_URL=https://master-ip:6443 K3S_TOKEN=node-token sh -
```

#### Backup dan Recovery

```bash
# Backup K3s data
sudo cp -r /var/lib/rancher/k3s/server/db /backup/k3s-$(date +%Y%m%d)

# Backup aplikasi data
kubectl get pvc -n belimang -o yaml > belimang-pvc-backup.yaml
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
