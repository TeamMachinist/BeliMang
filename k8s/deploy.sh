#!/bin/bash

# Belimang Kubernetes Deployment Script

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

# Function to check if kubectl is available
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    print_success "kubectl is available"
}

# Function to check if using registry images
check_registry_config() {
    print_status "Checking image configuration..."
    
    # Check if images in deployment files are using registry
    POSTGRES_IMAGE=$(grep "image:" postgres-deployment.yaml | awk '{print $2}')
    APP_IMAGE=$(grep "image:" app-deployment.yaml | awk '{print $2}')
    
    if [[ "$POSTGRES_IMAGE" == *"/"* ]] && [[ "$APP_IMAGE" == *"/"* ]]; then
        print_success "Using registry images: $POSTGRES_IMAGE, $APP_IMAGE"
        USING_REGISTRY=true
    else
        print_warning "Using local images. For production, use 'make k8s-deploy' with registry images."
        USING_REGISTRY=false
    fi
}

# Function to load local images to Kubernetes (for development)
load_local_images() {
    if [ "$USING_REGISTRY" = false ]; then
        print_status "Loading local images to Kubernetes..."
        
        # Detect Kubernetes environment
        if command -v minikube &> /dev/null && minikube status &> /dev/null; then
            print_status "Detected Minikube, loading images..."
            if docker image inspect belimang-postgres:latest &> /dev/null; then
                minikube image load belimang-postgres:latest
            fi
            if docker image inspect belimang-app:latest &> /dev/null; then
                minikube image load belimang-app:latest
            fi
            print_success "Images loaded to Minikube"
        elif command -v kind &> /dev/null; then
            # Check if kind cluster exists
            if kind get clusters 2>/dev/null | grep -q "kind"; then
                print_status "Detected Kind, loading images..."
                if docker image inspect belimang-postgres:latest &> /dev/null; then
                    kind load docker-image belimang-postgres:latest
                fi
                if docker image inspect belimang-app:latest &> /dev/null; then
                    kind load docker-image belimang-app:latest
                fi
                print_success "Images loaded to Kind"
            else
                print_warning "Kind command found but no cluster detected"
            fi
        elif command -v k3s &> /dev/null; then
            print_status "Detected K3s, importing images..."
            if docker image inspect belimang-postgres:latest &> /dev/null; then
                docker save belimang-postgres:latest | sudo k3s ctr images import -
            fi
            if docker image inspect belimang-app:latest &> /dev/null; then
                docker save belimang-app:latest | sudo k3s ctr images import -
            fi
            print_success "Images imported to K3s"
        else
            # Check if we're running on K3s by looking at cluster info
            CLUSTER_INFO=$(kubectl cluster-info 2>/dev/null | head -1 || echo "")
            if [[ "$CLUSTER_INFO" == *"127.0.0.1"* ]] || [[ "$CLUSTER_INFO" == *"localhost"* ]]; then
                print_warning "Detected possible K3s cluster. For K3s, use ./deploy-k3s.sh for better support."
                print_status "Attempting to import images to containerd..."
                if docker image inspect belimang-postgres:latest &> /dev/null; then
                    docker save belimang-postgres:latest | sudo ctr -n k8s.io images import - 2>/dev/null || true
                fi
                if docker image inspect belimang-app:latest &> /dev/null; then
                    docker save belimang-app:latest | sudo ctr -n k8s.io images import - 2>/dev/null || true
                fi
            else
                print_warning "Could not detect Minikube, Kind, or K3s. Make sure images are available in your cluster."
            fi
        fi
    else
        print_status "Using registry images, skipping local image loading"
    fi
}

# Function to update postgres deployment with current path
update_postgres_paths() {
    print_status "Updating PostgreSQL deployment paths..."
    
    CURRENT_DIR=$(pwd)
    PROJECT_ROOT=$(dirname "$CURRENT_DIR")
    
    # Create temporary file with updated paths
    sed "s|/path/to/your/project|$PROJECT_ROOT|g" postgres-deployment.yaml > postgres-deployment-temp.yaml
    
    print_success "Updated paths in postgres-deployment.yaml"
}

# Function to deploy to Kubernetes
deploy_k8s() {
    print_status "Deploying to Kubernetes..."
    
    # Deploy namespace first
    print_status "Creating namespace..."
    kubectl apply -f namespace.yaml
    
    # Deploy ConfigMap
    print_status "Deploying ConfigMap..."
    kubectl apply -f configmap.yaml
    
    # Deploy PVCs
    print_status "Deploying Persistent Volume Claims..."
    kubectl apply -f postgres-pvc.yaml
    kubectl apply -f redis-pvc.yaml
    
    # Deploy PostgreSQL
    print_status "Deploying PostgreSQL..."
    kubectl apply -f postgres-deployment-temp.yaml
    kubectl apply -f postgres-service.yaml
    
    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/postgres-deployment -n belimang
    
    # Deploy Redis
    print_status "Deploying Redis..."
    kubectl apply -f redis-deployment.yaml
    kubectl apply -f redis-service.yaml
    
    # Wait for Redis to be ready
    print_status "Waiting for Redis to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/redis-deployment -n belimang
    
    # Deploy Application
    print_status "Deploying Application..."
    kubectl apply -f app-deployment.yaml
    kubectl apply -f app-service.yaml
    
    # Wait for Application to be ready
    print_status "Waiting for Application to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/belimang-app-deployment -n belimang
    
    # Clean up temporary file
    rm -f postgres-deployment-temp.yaml
    
    print_success "Deployment completed successfully!"
}

# Function to show deployment status
show_status() {
    print_status "Deployment Status:"
    echo
    kubectl get pods -n belimang
    echo
    kubectl get services -n belimang
    echo
    
    # Get service URL
    SERVICE_TYPE=$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.spec.type}')
    
    if [ "$SERVICE_TYPE" = "LoadBalancer" ]; then
        EXTERNAL_IP=$(kubectl get service belimang-app-service -n belimang -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -n "$EXTERNAL_IP" ]; then
            print_success "Application available at: http://$EXTERNAL_IP"
        else
            print_warning "LoadBalancer IP pending. Use port-forward to access:"
            print_status "kubectl port-forward service/belimang-app-service 8080:80 -n belimang"
        fi
    else
        print_status "Use port-forward to access the application:"
        print_status "kubectl port-forward service/belimang-app-service 8080:80 -n belimang"
    fi
}

# Function to clean up deployment
cleanup() {
    print_status "Cleaning up deployment..."
    kubectl delete namespace belimang --ignore-not-found=true
    print_success "Cleanup completed"
}

# Main script
case "${1:-deploy}" in
    "deploy")
        print_status "Starting Belimang Kubernetes deployment..."
        check_kubectl
        check_registry_config
        load_local_images
        update_postgres_paths
        deploy_k8s
        show_status
        ;;
    "status")
        show_status
        ;;
    "cleanup")
        cleanup
        ;;
    "build")
        check_registry_config
        load_local_images
        ;;
    *)
        echo "Usage: $0 {deploy|status|cleanup|build}"
        echo "  deploy  - Deploy the application to Kubernetes (default)"
        echo "  status  - Show deployment status"
        echo "  cleanup - Remove the deployment"
        echo "  build   - Build and load Docker images"
        exit 1
        ;;
esac