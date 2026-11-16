pipeline {
    agent any

    options {
        disableConcurrentBuilds()
        timeout(time: 30, unit: 'MINUTES')
    }

    environment {
        // GCP VM Connection Details
        GCP_VM_IP   = '34.135.165.205'
        GCP_VM_USER = 'leonive_es_gmail_com'
        APP_DIR     = '/home/leonive_es_gmail_com/saltwake-ecommerce'
    }

    stages {
        stage('Checkout Code') {
            steps {
                echo 'Pulling latest code from Git repository...'
                checkout scm
            }
        }

        stage('Deploy All-in-One Stack (Backend, Frontend, DynamoDB)') {
            steps {
                // Only ONE credential needed: SSH key to your GCP VM
                withCredentials([
                    sshUserPrivateKey(credentialsId: 'gcp-vm-ssh-key', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')
                ]) {
                    echo "Connecting to GCP VM at ${GCP_VM_IP} and launching 3-container stack..."

                    sh '''
                        set -e
                        SSH_OPTS="-o StrictHostKeyChecking=no -i ${SSH_KEY}"

                        # 1. Create app directory on the VM
                        ssh ${SSH_OPTS} ${GCP_VM_USER}@${GCP_VM_IP} "mkdir -p ${APP_DIR}"

                        # 2. Sync codebase using rsync if available, otherwise fallback to tar over ssh
                        if command -v rsync >/dev/null 2>&1; then
                            rsync -avz --delete --exclude '.git' --exclude 'frontend/node_modules' -e "ssh ${SSH_OPTS}" ./ ${GCP_VM_USER}@${GCP_VM_IP}:${APP_DIR}/
                        else
                            echo "rsync not found on Jenkins host; streaming tar archive over ssh..."
                            tar --exclude='./.git' --exclude='./frontend/node_modules' -czf - . | ssh ${SSH_OPTS} ${GCP_VM_USER}@${GCP_VM_IP} "tar -xzf - -C ${APP_DIR}"
                        fi

                        # 3. Start the 3-container Docker Compose stack on the VM
                        ssh ${SSH_OPTS} ${GCP_VM_USER}@${GCP_VM_IP} "
                            set -e
                            cd ${APP_DIR}

                            # Ensure Docker & Docker Compose are installed
                            if ! command -v docker >/dev/null 2>&1; then
                                echo 'Installing docker...'
                                sudo apt-get update && sudo apt-get install -y docker.io docker-compose-v2
                                sudo usermod -aG docker \$USER
                            fi

                            # 1. Stop and remove existing compose project first
                            echo 'Stopping existing compose services...'
                            sudo docker compose down --remove-orphans -t 10 || true

                            # 2. Force remove named containers if still around
                            sudo docker rm -f saltwake-backend saltwake-frontend dynamodb-local >/dev/null 2>&1 || true

                            # 3. Kill any rogue docker container bound to our ports (including other compose projects / standalone runs)
                            for port in 8000 8080 5173; do
                                sudo docker ps -q --filter publish=\$port | xargs -r sudo docker rm -f || true
                            done

                            # 4. Kill any host-level or zombie processes bound to our ports
                            echo 'Releasing ports 8000, 8080, 5173...'
                            for port in 8000 8080 5173; do
                                if command -v fuser >/dev/null 2>&1; then
                                    sudo fuser -k -9 \${port}/tcp >/dev/null 2>&1 || true
                                fi
                                if command -v lsof >/dev/null 2>&1; then
                                    sudo lsof -t -i:\$port | xargs -r sudo kill -9 >/dev/null 2>&1 || true
                                fi
                                if command -v ss >/dev/null 2>&1; then
                                    sudo ss -lptn "sport = :\$port" || true
                                fi
                            done
                            sleep 2

                            echo 'Launching DynamoDB Local, Go Backend, and React Frontend containers...'
                            sudo docker compose up -d --build --wait --wait-timeout 300

                            echo 'Verifying all 3 containers are running...'
                            sudo docker compose ps
                            sudo docker compose ps --status running --services | grep -qx backend
                            sudo docker compose ps --status running --services | grep -qx frontend
                            sudo docker compose ps --status running --services | grep -qx dynamodb-local

                            echo 'All 3 containers successfully launched!'

                            # Keep the VM from filling up with dangling images over repeated deploys
                            sudo docker image prune -f >/dev/null || true
                        "
                    '''
                }
            }
        }

        stage('Verify Deployment Health') {
            steps {
                echo 'Checking frontend, backend & DynamoDB health...'
                sh """
                    set -e
                    check_url() {
                        url=\$1
                        for i in \$(seq 1 30); do
                            if curl -fsS -o /dev/null "\$url"; then
                                echo "\$url is healthy after \${i} attempt(s)."
                                return 0
                            fi
                            echo "\$url not ready yet (\${i}/30), retrying in 2s..."
                            sleep 2
                        done
                        echo "\$url failed health check after 60s"
                        return 1
                    }
                    check_url http://${GCP_VM_IP}:8080/health/
                    check_url http://${GCP_VM_IP}:5173/
                """
            }
        }
    }

    post {
        success {
            echo "=========================================================="
            echo "DEPLOYMENT SUCCEEDED!"
            echo "Frontend UI:  http://${GCP_VM_IP}:5173"
            echo "Backend API:  http://${GCP_VM_IP}:8080"
            echo "DynamoDB:     http://${GCP_VM_IP}:8000"
            echo "=========================================================="
        }
        failure {
            echo "Deployment failed! Check the Jenkins console logs above."
        }
    }
}
