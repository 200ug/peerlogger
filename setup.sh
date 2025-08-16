#!/bin/bash

# all-in-one podman + ufw setup helper for debian vps deployment(s)

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root. Run with sudo when needed."
        exit 1
    fi
}

check_dependencies() {
    log_info "Checking dependencies for Debian VPS..."
    
    sudo apt update
    sudo apt install -y ufw podman podman-compose curl wget git htop
    
    log_info "Dependencies installation completed"
}

configure_firewall() {
    log_info "Configuring UFW firewall..."
    
    # reset
    sudo ufw --force reset
    
    # defaults
    sudo ufw default deny incoming
    sudo ufw default allow outgoing
    
    # ssh to non-default port
    sudo ufw allow 10980/tcp comment 'SSH access'
    
    # crawler udp traffic to ethereum's default port
    sudo ufw allow 30303/udp comment 'Ethereum peer discovery'
    
    # loopback traffic
    sudo ufw allow in on lo
    sudo ufw allow out on lo
    
    # enable
    sudo ufw --force enable
    
    log_info "UFW firewall configured successfully"
}

configure_podman_networking() {
    log_info "Configuring Podman networking for Debian VPS..."
    
    # enable and start podman socket service for better systemd integration
    systemctl --user enable podman.socket || log_warn "Could not enable podman.socket (may require re-login)"
    
    # network config directory
    mkdir -p ~/.config/containers/
    
    if [[ ! -f ~/.config/containers/containers.conf ]]; then
        cat > ~/.config/containers/containers.conf << 'EOF'
[containers]
# use systemd for cgroup management (default on debian)
cgroup_manager = "systemd"

# optimization for vps resource constraints
default_ulimits = [
  "nofile=65536:65536",
  "nproc=4096:4096"
]

[network]
network_backend = "netavark"

[engine]
# optimization for vps resource constraints
runtime = "crun"
EOF
        log_info "Created Podman containers.conf optimized for Debian VPS"
    fi
    
    # ensure proper permissions for rootless podman
    if [[ ! -f /etc/subuid ]] || ! grep -q "^$(whoami):" /etc/subuid; then
        log_warn "Setting up subuid/subgid for rootless Podman..."
        echo "$(whoami):100000:65536" | sudo tee -a /etc/subuid
        echo "$(whoami):100000:65536" | sudo tee -a /etc/subgid
    fi
    
    log_info "Podman networking configured for UFW compatibility"
}

verify_prod_compose() {
    log_info "Checking for production compose file..."
    
    if [[ ! -f "docker-compose.prod.yml" ]]; then
        log_error "docker-compose.prod.yml not found!"
        log_error "This file should be included in your project."
        exit 1
    fi
    
    log_info "Production compose file found"
}

configure_podman_ufw_rules() {
    log_info "Adding Podman-specific UFW rules..."
    
    # allow traffic on podman's default bridge network
    sudo ufw allow in on podman0
    sudo ufw allow out on podman0
    
    # allow traffic on docker/podman bridge interfaces
    # (covers any bridge networks podman might create)
    sudo ufw allow in on br-+
    sudo ufw allow out on br-+
    
    log_info "Podman UFW rules added"
}

show_configuration() {
    log_info "Firewall (UFW) config:"
    echo
    sudo ufw status verbose
    echo
    log_info "Podman config:"
    echo -e "\tNetwork mode: Host networking for crawler service"
    echo -e "\tDatabase: Bridge network, localhost-only binding"
    echo
    log_warn "Important Notes:"
    echo -e "\t- The crawler will use host networking (no Docker network isolation)"
    echo -e "\t- Database remains on bridge network for security"
    echo -e "\t- Only UDP port 30303 and TCP port 10980 are exposed"
    echo
    log_info "Setup completed successfully!"
}

verify_setup() {
    log_info "Verifying firewall setup..."
    
    # check if ufw is active
    if ! sudo ufw status | grep -q "Status: active"; then
        log_error "UFW is not active!"
        exit 1
    fi
    
    # check if required ports are allowed
    if ! sudo ufw status | grep -q "10980/tcp"; then
        log_error "SSH port 10980 not found in UFW rules!"
        exit 1
    fi
    
    if ! sudo ufw status | grep -q "30303/udp"; then
        log_error "Crawler UDP port 30303 not found in UFW rules!"
        exit 1
    fi
    
    log_info "Firewall verification passed"
}

main() {
    log_info "Starting setup helper..."
    
    check_root
    check_dependencies
    verify_prod_compose
    configure_firewall
    configure_podman_networking
    configure_podman_ufw_rules
    verify_setup
    show_configuration
    
    echo
    log_info "You can now run your crawler with:"
    echo "  podman-compose -f docker-compose.prod.yml up -d"
    echo
    log_warn "Next steps for Debian VPS:"
    echo "1. Copy your .env file to the project directory"
    echo "2. Make sure GeoIP databases are in ./geoip/ directory"
    echo "3. Run './launch.sh up' to start"
    echo "4. Monitor logs with 'podman-compose -f docker-compose.prod.yml logs -f'"
}

main "$@"
