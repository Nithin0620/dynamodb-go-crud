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
                        SSH_OPTS="-o StrictHostKeyChecking=no -i ${SSH_KEY}"

                        # 1. Create app directory on the VM
                        ssh ${SSH_OPTS} ${GCP_VM_USER}@${GCP_VM_IP} "mkdir -p ${APP_DIR}"

                        # 2. Sync codebase to the VM (excluding unnecessary local node_modules)
                        rsync -avz --delete --exclude '.git' --exclude 'frontend/node_modules' -e "ssh ${SSH_OPTS}" ./ ${GCP_VM_USER}@${GCP_VM_IP}:${APP_DIR}/

                        # 3. Start the 3-container Docker Compose stack on the VM
                        ssh ${SSH_OPTS} ${GCP_VM_USER}@${GCP_VM_IP} "
                            cd ${APP_DIR}

                            # Ensure Docker & Docker Compose are installed
                            if ! command -v docker &> /dev/null; then
                                echo 'Installing docker...'
                                sudo apt-get update && sudo apt-get install -y docker.io docker-compose-v2
                                sudo usermod -aG docker \$USER
                            fi

                            echo 'Launching DynamoDB Local, Go Backend, and React Frontend containers...'
                            sudo docker compose down --remove-orphans || true
                            sudo docker compose up -d --build

                            echo 'All 3 containers successfully launched!'
                            sudo docker compose ps
                        "
                    '''
                }
            }
        }

        stage('Verify Deployment Health') {
            steps {
                echo 'Checking backend & DynamoDB health...'
                sh """
                    sleep 10
                    curl -sSf http://${GCP_VM_IP}:8080/health || (echo 'Backend health check failed' && exit 1)
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
