pipeline {
    agent any

    environment {
        DOCKER_IMAGE = 'golang:1.21' 
    }

    stages {
        stage('Checkout') {
            steps {
                // Clonar el repositorio
                git 'https://github.com/v3ctr4x/SentryGo.git' 
            }
        }

        stage('Build Linux AMD64') {
            steps {
                script {
                    // Ejecutar la compilación en un contenedor Docker
                    docker.image(DOCKER_IMAGE).inside {
                        sh 'GOOS=linux GOARCH=amd64 go build -o sentrygo-server-linux cmd/server/main.go'
                        sh 'GOOS=linux GOARCH=amd64 go build -o sentrygo-agent-linux cmd/agent/main.go'
                    }
                }
            }
        }

        stage('Build Windows AMD64') {
            steps {
                script {
                    // Ejecutar la compilación en un contenedor Docker
                    docker.image(DOCKER_IMAGE).inside {
                        sh 'GOOS=windows GOARCH=amd64 go build -o sentrygo-server.exe cmd/server/main.go'
                        sh 'GOOS=windows GOARCH=amd64 go build -o sentrygo-agent.exe cmd/agent/main.go'
                    }
                }
            }
        }

        stage('Build macOS AMD64') {
            steps {
                script {
                    // Ejecutar la compilación en un contenedor Docker
                    docker.image(DOCKER_IMAGE).inside {
                        sh 'GOOS=darwin GOARCH=amd64 go build -o sentrygo-server-darwin cmd/server/main.go'
                        sh 'GOOS=darwin GOARCH=amd64 go build -o sentrygo-agent-darwin cmd/agent/main.go'
                    }
                }
            }
        }

        stage('Archive Build Artifacts') {
            steps {
                script {
                    // Archivar los archivos binarios generados
                    archiveArtifacts artifacts: '**/sentrygo-*', allowEmptyArchive: true
                }
            }
        }

        stage('Cleanup') {
            steps {
                script {
                    // Eliminar contenedores Docker que se hayan usado
                    sh 'docker system prune -af || true'
                }
            }
        }
    }

    post {
        always {
            // Limpiar contenedores y volúmenes en caso de cualquier error
            sh 'docker system prune -af || true'
        }
    }
}