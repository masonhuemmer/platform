pipeline {
    agent any
    
    parameters {
        string(name: 'GOOS', defaultValue: 'linux', description: 'Target OS')
        string(name: 'GOARCH', defaultValue: 'amd64', description: 'Target Architecture')
        string(name: 'PLATFORM_VERSION', defaultValue: '1.0.0', description: 'Platform Version')
        string(name: 'PLATFORM_REVISION', defaultValue: 'abc123', description: 'Platform Revision')
    }
    
    environment {
        BUILD_BINARIES_DIR = "${WORKSPACE}/build"
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Setup Environment') {
            steps {
                script {
                    sh 'mkdir -p ${BUILD_BINARIES_DIR}'
                }
            }
        }
        
        stage('Build') {
            steps {
                script {
                    sh '''
                        GOOS=${GOOS} GOARCH=${GOARCH} go build \
                        -ldflags="-X 'main.Version=${PLATFORM_VERSION}' -X 'main.Revision=${PLATFORM_REVISION}'" \
                        -o ${BUILD_BINARIES_DIR}/platform
                    '''
                }
            }
        }
        
        stage('Archive Artifact') {
            steps {
                archiveArtifacts artifacts: 'build/platform', fingerprint: true
            }
        }
    }
}
