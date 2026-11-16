#!/bin/bash
set -euo pipefail

APP_DIR="${1:-/home/leonive_es_gmail_com/saltwake-ecommerce}"
cd "$APP_DIR"

echo "=== Deploying SaltWake Ecommerce Stack ==="

# 1. Recover Docker daemon if daemon.json was corrupted or stopped
if [ -f /etc/docker/daemon.json ]; then
    # Validate JSON syntax; if invalid, remove it
    if command -v python3 >/dev/null 2>&1; then
        if ! python3 -m json.tool /etc/docker/daemon.json >/dev/null 2>&1; then
            echo "Corrupted /etc/docker/daemon.json detected, resetting..."
            sudo rm -f /etc/docker/daemon.json
        fi
    elif ! sudo docker info >/dev/null 2>&1; then
        echo "Docker not responding, removing daemon.json to recover daemon..."
        sudo rm -f /etc/docker/daemon.json
    fi
fi

# Ensure Docker service is running
sudo systemctl restart docker || sudo systemctl start docker

# Ensure docker CLI works
if ! sudo docker info >/dev/null 2>&1; then
    echo "ERROR: Docker daemon is not running properly!"
    sudo systemctl status docker.service --no-pager || true
    exit 1
fi

# 2. Stop existing compose services
echo "Stopping existing compose containers..."
sudo docker compose down --remove-orphans -t 10 || true

# 3. Clean up named containers
sudo docker rm -f saltwake-backend saltwake-frontend dynamodb-local >/dev/null 2>&1 || true

# 4. Clean up any rogue container or processes on app ports (8000, 8081, 5173)
# Note: NEVER TOUCH PORT 8080 (where Jenkins runs!)
echo "Releasing app ports 8000, 8081, 5173..."
for port in 8000 8081 5173; do
    sudo docker ps -q --filter "publish=$port" | xargs -r sudo docker rm -f || true
    if command -v fuser >/dev/null 2>&1; then
        sudo fuser -k -9 "${port}/tcp" >/dev/null 2>&1 || true
    fi
    if command -v lsof >/dev/null 2>&1; then
        sudo lsof -t -i:"$port" | xargs -r sudo kill -9 >/dev/null 2>&1 || true
    fi
done
sleep 2

# 5. Launch containers with docker compose
echo "Starting containers via Docker Compose..."
sudo docker compose up -d --build --wait --wait-timeout 300

# 6. Verify containers are running
echo "Checking running containers..."
sudo docker compose ps
sudo docker compose ps --status running --services | grep -qx backend
sudo docker compose ps --status running --services | grep -qx frontend
sudo docker compose ps --status running --services | grep -qx dynamodb-local

echo "=== Deployment Successfully Completed ==="
sudo docker image prune -f >/dev/null || true
