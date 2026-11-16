pipeline {
    agent any

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
                            if ! command -v docker &> /dev/null; then
                                echo 'Installing docker...'
                                sudo apt-get update && sudo apt-get install -y docker.io docker-compose-v2
                                sudo usermod -aG docker \$USER
                            fi

                            echo 'Launching DynamoDB Local, Go Backend, and React Frontend containers...'
                            sudo docker compose down --remove-orphans || true
                            sudo docker compose up -d --build --wait --wait-timeout 180

                            echo 'Verifying all 3 containers are running...'
                            sudo docker compose ps
                            sudo docker compose ps --status running --services | grep -qx backend
                            sudo docker compose ps --status running --services | grep -qx frontend
                            sudo docker compose ps --status running --services | grep -qx dynamodb-local

                            echo 'All 3 containers successfully launched!'
                        "
                    '''
                }
            }
        }

        stage('Verify Deployment Health') {
            steps {
                echo 'Checking backend & DynamoDB health...'
                sh """
                    set -e
                    for i in \$(seq 1 30); do
                        if curl -fsSL http://${GCP_VM_IP}:8080/health/; then
                            echo ""
                            echo "Backend healthy after \${i} attempt(s)."
                            exit 0
                        fi
                        echo "Backend not ready yet (\${i}/30), retrying in 2s..."
                        sleep 2
                    done
                    echo 'Backend health check failed after 60s'
                    exit 1
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
