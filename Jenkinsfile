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

                        # 3. Execute dedicated deployment script on the VM
                        ssh ${SSH_OPTS} ${GCP_VM_USER}@${GCP_VM_IP} "bash ${APP_DIR}/deploy/deploy.sh ${APP_DIR}"
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
                    check_url http://${GCP_VM_IP}:8081/health/
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
            echo "Backend API:  http://${GCP_VM_IP}:8081"
            echo "DynamoDB:     http://${GCP_VM_IP}:8000"
            echo "=========================================================="
        }
        failure {
            echo "Deployment failed! Check the Jenkins console logs above."
        }
    }
}
