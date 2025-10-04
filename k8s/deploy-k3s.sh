#!/bin/bash

# Belimang K3s Deployment Script
# Optimized for K3s lightweight Kubernetes distribution

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables from .env file
load_env() {
    if [ -f ".env" ]; then
        print_status "Loading environment variables from .env file..."
        export $(grep -v '^#' .env | xargs)
        print_success "Environment variables loaded"
    else
        print_warning ".env file not found. Using default values from .env.example"
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_status "Created .env from .env.example. Please edit .env with your values."
            exit 1
        fi
    fi
}

# Function to check if kubectl is available and K3s is running
check_k3s() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        print_status "Install kubectl: curl -LO https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        exit 1
    fi
    
    # Check if K3s is running
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster. Is K3s running?"
        print_status "Start K3s: sudo systemctl start k3s"
        print_status "Or install K3s: curl -sfL https://get.k3s.io | sh -"
        exit 1
    fi
    
    # Check if we can access the cluster
    CLUSTER_INFO=$(kubectl cluster-info 2>/dev/null | head -1)
    if [[ "$CLUSTER_INFO" == *"k3s"* ]] || [[ "$CLUSTER_INFO" == *"127.0.0.1"* ]] || [[ "$CLUSTER_INFO" == *"localhost"* ]]; then
        print_success "K3s cluster detected and accessible"
    else
        print_warning "Cluster detected but may not be K3s: $CLUSTER_INFO"
    fi
}

# Function to check and import Docker images to K3s
import_images_k3s() {
    print_status "Checking Docker images for K3s..."
    
    # Check if images exist locally
    POSTGRES_IMAGE_EXISTS=$(docker image inspect belimang-postgres:latest &> /dev/null && echo "true" || echo "false")
    APP_IMAGE_EXISTS=$(docker image inspect belimang-app:latest &> /dev/null && echo "true" || echo "false")
    
    if [ "$POSTGRES_IMAGE_EXISTS" = "false" ] || [ "$APP_IMAGE_EXISTS" = "false" ]; then
        print_warning "Local images not found. Building images..."
        cd ..
        
        if [ "$POSTGRES_IMAGE_EXISTS" = "false" ]; then
            print_status "Building PostgreSQL image..."
            docker build -t belimang-postgres:latest --target postgres-h3 .
        fi
        
        if [ "$APP_IMAGE_EXISTS" = "false" ]; then
            print_status "Building application image..."
            docker build -t belimang-app:latest --target production .
        fi
        
        cd k8s
        print_success "Images built successfully"
    fi
    
    # Import images to K3s
    print_status "Importing images to K3s..."
    
    # K3s uses containerd, so we need to import images
    if command -v k3s &> /dev/null; then
        # Save images to tar and import to K3s
        print_status "Saving PostgreSQL image..."
        docker save belimang-postgres:latest | sudo k3s ctr images import -
        
        print_status "Saving application image..."
        docker save belimang-app:latest | sudo k3s ctr images import -
        
        print_success "Images imported to K3s successfully"
    else
        print_warning "k3s command not found. Images may not be available in K3s runtime."
        print_status "You can manually import with: docker save <image> | sudo k3s ctr images import -"
    fi
}

# Function to generate ConfigMap from .env
generate_configmap() {
    print_status "Generating ConfigMap from environment variables..."
    
    # Create configmap from .env file
    cat > configmap-generated.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: belimang-config
  namespace: belimang
data:
  ENV: "${ENV:-production}"
  GIN_MODE: "${GIN_MODE:-release}"
  HTTP_PORT: "${HTTP_PORT:-8080}"
  DB_HOST: "${DB_HOST:-postgres-service}"
  DB_PORT: "${DB_PORT:-5432}"
  DB_USER: "${DB_USER:-postgres}"
  DB_PASSWORD: "${DB_PASSWORD:-password}"
  DB_NAME: "${DB_NAME:-belimang}"
  DB_SSLMODE: "${DB_SSLMODE:-disable}"
  DATABASE_URL: "${DATABASE_URL:-postgres://postgres:password@postgres-service:5432/belimang?sslmode=disable}"
  CACHE_HOST: "${CACHE_HOST:-redis-service}"
  CACHE_PORT: "${CACHE_PORT:-6379}"
  CACHE_PASSWORD: "${CACHE_PASSWORD:-}"
  CACHE_DB: "${CACHE_DB:-0}"
  REDIS_ADDR: "${REDIS_ADDR:-redis-service:6379}"
  LOG_LEVEL: "${LOG_LEVEL:-info}"
  LOG_TYPE: "${LOG_TYPE:-simple}"
  JWT_SECRET_KEY: "${JWT_SECRET_KEY:-your-secret-key-change-in-production}"
  JWT_ISSUER: "${JWT_ISSUER:-belimang-app}"
  GOMAXPROCS: "${GOMAXPROCS:-4}"
  GOMEMLIMIT: "${GOMEMLIMIT:-1536MiB}"
  GODEBUG: "${GODEBUG:-asyncpreemptoff=1}"
  GOCACHE: "${GOCACHE:-/tmp/go-cache}"
  GOTMPDIR: "${GOTMPDIR:-/tmp/go-tmp}"
EOF
    
    print_success "ConfigMap generated from .env variables"
}

# Function to update PVC with K3s local-path storage class
update_pvc_k3s() {
    print_status "Updating PVC for K3s local-path storage..."
    
    # Update postgres PVC
    sed "s/storageClassName: standard/storageClassName: ${STORAGE_CLASS:-local-path}/g" postgres-pvc.yaml > postgres-pvc-k3s.yaml
    sed -i "s/storage: 10Gi/storage: ${POSTGRES_STORAGE_SIZE:-10Gi}/g" postgres-pvc-k3s.yaml
    
    # Update redis PVC
    sed "s/storageClassName: standard/storageClassName: ${STORAGE_CLASS:-local-path}/g" redis-pvc.yaml > redis-pvc-k3s.yaml
    sed -i "s/storage: 5Gi/storage: ${REDIS_STORAGE_SIZE:-5Gi}/g" redis-pvc-k3s.yaml
    
    print_success "PVC updated for K3s"
}

# Function to update deployments with resource limits from .env
update_deployments_k3s() {
    print_status "Updating deployments with K3s optimizations..."
    
    CURRENT_DIR=$(pwd)
    PROJECT_ROOT=$(dirname "$CURRENT_DIR")
    
    # Update postgres deployment
    sed "s|/path/to/your/project|$PROJECT_ROOT|g" postgres-deployment.yaml > postgres-deployment-k3s.yaml
    
    # Update resource limits if specified in .env
    if [ -n "$POSTGRES_CPU_LIMIT" ]; then
        sed -i "s/cpu: \"2\"/cpu: \"${POSTGRES_CPU_LIMIT}\"/g" postgres-deployment-k3s.yaml
    fi
    if [ -n "$POSTGRES_MEMORY_LIMIT" ]; then
        sed -i "s/memory: \"2Gi\"/memory: \"${POSTGRES_MEMORY_LIMIT}\"/g" postgres-deployment-k3s.yaml
    fi
    if [ -n "$POSTGRES_CPU_REQUEST" ]; then
        sed -i "s/cpu: \"1\"/cpu: \"${POSTGRES_CPU_REQUEST}\"/g" postgres-deployment-k3s.yaml
    fi
    if [ -n "$POSTGRES_MEMORY_REQUEST" ]; then
        sed -i "s/memory: \"512Mi\"/memory: \"${POSTGRES_MEMORY_REQUEST}\"/g" postgres-deployment-k3s.yaml
    fi
    
    # Update app deployment
    cp app-deployment.yaml app-deployment-k3s.yaml
    
    # Update app replicas
    if [ -n "$APP_REPLICAS" ]; then
        sed -i "s/replicas: 2/replicas: ${APP_REPLICAS}/g" app-deployment-k3s.yaml
    fi
    
    # Update app resource limits
    if [ -n "$APP_CPU_LIMIT" ]; then
        sed -i "s/cpu: \"4\"/cpu: \"${APP_CPU_LIMIT}\"/g" app-deployment-k3s.yaml
    fi
    if [ -n "$APP_MEMORY_LIMIT" ]; then
        sed -i "s/memory: \"2Gi\"/memory: \"${APP_MEMORY_LIMIT}\"/g" app-deployment-k3s.yaml
    fi
    if [ -n "$APP_CPU_REQUEST" ]; then
        sed -i "s/cpu: \"2\"/cpu: \"${APP_CPU_REQUEST}\"/g" app-deployment-k3s.yaml
    fi
    if [ -n "$APP_MEMORY_REQUEST" ]; then
        sed -i "s/memory: \"512Mi\"/memory: \"${APP_MEMORY_REQUEST}\"/g" app-deployment-k3s.yaml
    fi
    
    # Update redis deployment
    cp redis-deployment.yaml redis-deployment-k3s.yaml
    
    # Update redis resource limits
    if [ -n "$REDIS_CPU_LIMIT" ]; then
        sed -i "s/cpu: \"1\"/cpu: \"${REDIS_CPU_LIMIT}\"/g" redis-deployment-k3s.yaml
    fi
    if [ -n "$REDIS_MEMORY_LIMIT" ]; then
        sed -i "s/memory: \"768Mi\"/memory: \"${REDIS_MEMORY_LIMIT}\"/g" redis-deployment-k3s.yaml
    fi
    if [ -n "$REDIS_CPU_REQUEST" ]; then
        sed -i "s/cpu: \"500m\"/cpu: \"${REDIS_CPU_REQUEST}\"/g" redis-deployment-k3s.yaml
    fi
    if [ -n "$REDIS_MEMORY_REQUEST" ]; then
        sed -i "s/memory: \"256Mi\"/memory: \"${REDIS_MEMORY_REQUEST}\"/g" redis-deployment-k3s.yaml
    fi
    
    print_success "Deployments updated for K3s"
}

# Function to deploy to K3s
deploy_k3s() {
    print_status "Deploying to K3s..."
    
    # Deploy namespace first
    print_status "Creating namespace..."
    kubectl apply -f namespace.yaml
    
    # Deploy ConfigMap
    print_status "Deploying ConfigMap..."
    kubectl apply -f configmap-generated.yaml
    
    # Deploy PVCs
    print_status "Deploying Persistent Volume Claims..."
    kubectl apply -f postgres-pvc-k3s.yaml
    kubectl apply -f redis-pvc-k3s.yaml
    
    # Deploy PostgreSQL
    print_status "Deploying PostgreSQL..."
    kubectl apply -f postgres-deployment-k3s.yaml
    kubectl apply -f postgres-service.yaml
    
    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/postgres-deployment -n belimang
    
    # Deploy Redis
    print_status "Deploying Redis..."
    kubectl apply -f redis-deployment-k3s.yaml
    kubectl apply -f redis-service.yaml
    
    # Wait for Redis to be ready
    print_status "Waiting for Redis to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/redis-deployment -n belimang
    
    # Deploy Application
    print_status "Deploying Application..."
    kubectl apply -f app-deployment-k3s.yaml
    kubectl apply -f app-service.yaml
    
    # Wait for Application to be ready
    print_status "Waiting for Application to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/belimang-app-deployment -n belimang
    
    print_success "K3s deployment completed successfully!"
}

# Function to show deployment status
show_status() {
    print_status "K3s Deployment Status:"
    echo
    kubectl get nodes
    echo
    kubectl get pods -n belimang -o wide
    echo
    kubectl get services -n belimang
    echo
    kubectl get pvc -n belimang
    echo
    
    # Get service URL
    SERVICE_TYPE=$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.spec.type}')
    
    if [ "$SERVICE_TYPE" = "LoadBalancer" ]; then
        # K3s includes a built-in load balancer (servicelb)
        EXTERNAL_IP=$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -n "$EXTERNAL_IP" ] && [ "$EXTERNAL_IP" != "null" ]; then
            print_success "Application available at: http://$EXTERNAL_IP"
        else
            # Get node IP for K3s
            NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
            NODE_PORT=$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.spec.ports[0].nodePort}')
            if [ -n "$NODE_PORT" ]; then
                print_success "Application available at: http://$NODE_IP:$NODE_PORT"
            else
                print_warning "LoadBalancer IP pending. Use port-forward to access:"
                print_status "kubectl port-forward service/belimang-app-service 8080:80 -n belimang"
            fi
        fi
    else
        print_status "Use port-forward to access the application:"
        print_status "kubectl port-forward service/belimang-app-service 8080:80 -n belimang"
    fi
    
    print_status "To access from external network, you may need to configure K3s with --bind-address"
}

# Function to clean up deployment
cleanup() {
    print_status "Cleaning up K3s deployment..."
    kubectl delete namespace belimang --ignore-not-found=true
    
    # Clean up generated files
    rm -f configmap-generated.yaml
    rm -f postgres-pvc-k3s.yaml redis-pvc-k3s.yaml
    rm -f postgres-deployment-k3s.yaml app-deployment-k3s.yaml redis-deployment-k3s.yaml
    
    print_success "Cleanup completed"
}

# Function to show K3s specific help
show_help() {
    echo "Belimang K3s Deployment Script"
    echo
    echo "Usage: $0 {deploy|status|cleanup|help}"
    echo
    echo "Commands:"
    echo "  deploy  - Deploy the application to K3s (default)"
    echo "  status  - Show deployment status"
    echo "  cleanup - Remove the deployment and clean up files"
    echo "  help    - Show this help message"
    echo
    echo "Prerequisites:"
    echo "  - K3s installed and running"
    echo "  - kubectl configured"
    echo "  - Docker images built or available"
    echo "  - .env file configured (copy from .env.example)"
    echo
    echo "K3s Installation:"
    echo "  curl -sfL https://get.k3s.io | sh -"
    echo
    echo "Configuration:"
    echo "  Copy k8s/.env.example to k8s/.env and edit the values"
    echo
}

# Main script
case "${1:-deploy}" in
    "deploy")
        print_status "Starting Belimang K3s deployment..."
        load_env
        check_k3s
        import_images_k3s
        generate_configmap
        update_pvc_k3s
        update_deployments_k3s
        deploy_k3s
        show_status
        ;;
    "status")
        show_status
        ;;
    "cleanup")
        cleanup
        ;;
    "help")
        show_help
        ;;
    *)
        show_help
        exit 1
        ;;
esac