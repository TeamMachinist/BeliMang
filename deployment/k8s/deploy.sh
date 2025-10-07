#!/bin/bash

# Kubernetes Deployment Script for Belimang Application
# This script handles deployment to production Kubernetes cluster

set -e

# Configuration
NAMESPACE="belimang"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables
load_env() {
    if [ -f "$SCRIPT_DIR/.env" ]; then
        log_info "Loading environment variables from .env file..."
        set -a
        source "$SCRIPT_DIR/.env"
        set +a
        log_success "Environment variables loaded"
    else
        log_error ".env file not found in k8s directory"
        exit 1
    fi
}

# Check if kubectl is available and cluster is accessible
check_k8s_cluster() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    log_success "Kubernetes cluster detected and accessible"
}

# Generate ConfigMap from environment variables
generate_configmap() {
    log_info "Generating ConfigMap from environment variables..."
    
    # Create configmap with non-sensitive environment variables
    kubectl create configmap belimang-config \
        --namespace="$NAMESPACE" \
        --from-literal=ENV="$ENV" \
        --from-literal=GIN_MODE="$GIN_MODE" \
        --from-literal=HTTP_PORT="$HTTP_PORT" \
        --from-literal=REDIS_ADDR="$REDIS_ADDR" \
        --from-literal=JWT_ISSUER="$JWT_ISSUER" \
        --from-literal=LOG_LEVEL="$LOG_LEVEL" \
        --from-literal=LOG_TYPE="$LOG_TYPE" \
        --from-literal=GOMAXPROCS="$GOMAXPROCS" \
        --from-literal=GOMEMLIMIT="$GOMEMLIMIT" \
        --from-literal=GODEBUG="$GODEBUG" \
        --from-literal=GOCACHE="$GOCACHE" \
        --from-literal=GOTMPDIR="$GOTMPDIR" \
        --dry-run=client -o yaml > "$SCRIPT_DIR/configmap-generated.yaml"
    
    log_success "ConfigMap generated from .env variables"
}

# Deploy to Kubernetes
deploy() {
    log_info "Starting Belimang Kubernetes deployment..."
    
    load_env
    check_k8s_cluster
    generate_configmap
    
    log_info "Deploying to Kubernetes..."
    
    # Create namespace
    log_info "Creating namespace..."
    kubectl apply -f "$SCRIPT_DIR/namespace.yaml"
    
    # Deploy Secret using environment variables
    log_info "Deploying Secret using environment variables..."
    cd "$SCRIPT_DIR" && ./apply-secrets.sh
    
    # Deploy ConfigMap
    log_info "Deploying ConfigMap..."
    kubectl apply -f "$SCRIPT_DIR/configmap-generated.yaml"
    
    # Deploy Persistent Volume Claims
    log_info "Deploying Persistent Volume Claims..."
    kubectl apply -f "$SCRIPT_DIR/postgres-pvc.yaml"
    kubectl apply -f "$SCRIPT_DIR/redis-pvc.yaml"
    
    # Deploy PostgreSQL
    log_info "Deploying PostgreSQL..."
    kubectl apply -f "$SCRIPT_DIR/postgres-deployment.yaml"
    kubectl apply -f "$SCRIPT_DIR/postgres-service.yaml"
    
    # Wait for PostgreSQL to be ready
    log_info "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/postgres-deployment -n "$NAMESPACE"
    
    # Deploy Redis
    log_info "Deploying Redis..."
    kubectl apply -f "$SCRIPT_DIR/redis-deployment.yaml"
    kubectl apply -f "$SCRIPT_DIR/redis-service.yaml"
    
    # Wait for Redis to be ready
    log_info "Waiting for Redis to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/redis-deployment -n "$NAMESPACE"
    
    # Deploy Application
    log_info "Deploying Application..."
    kubectl apply -f "$SCRIPT_DIR/app-deployment.yaml"
    kubectl apply -f "$SCRIPT_DIR/app-service.yaml"
    
    # Wait for Application to be ready
    log_info "Waiting for Application to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/belimang-app-deployment -n "$NAMESPACE"
    
    log_success "Belimang application deployed successfully!"
    
    # Show deployment status
    log_info "Deployment Status:"
    kubectl get all -n "$NAMESPACE"
    
    # Show external access information
    log_info "External Access:"
    kubectl get services -n "$NAMESPACE" -o wide
}

# Cleanup deployment
cleanup() {
    log_info "Cleaning up Kubernetes deployment..."
    
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        kubectl delete namespace "$NAMESPACE"
        log_success "Cleanup completed"
    else
        log_warning "Namespace $NAMESPACE does not exist"
    fi
}

# Show help
show_help() {
    echo "Usage: $0 {deploy|cleanup|help}"
    echo ""
    echo "Commands:"
    echo "  deploy   - Deploy Belimang application to Kubernetes"
    echo "  cleanup  - Remove Belimang application from Kubernetes"
    echo "  help     - Show this help message"
    echo ""
    echo "Environment:"
    echo "  Requires .env file in k8s directory with configuration"
    echo "  Requires kubectl access to Kubernetes cluster"
}

# Main script logic
case "${1:-}" in
    deploy)
        deploy
        ;;
    cleanup)
        cleanup
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Invalid command: ${1:-}"
        show_help
        exit 1
        ;;
esac