pipeline {
    agent {
        docker { image 'golang:1.19' }
    }
    
    parameters {
        string(name: 'GOOS', defaultValue: 'linux', description: 'Target OS')
        string(name: 'GOARCH', defaultValue: 'amd64', description: 'Target Architecture')
        string(name: 'PLATFORM_VERSION', defaultValue: '1.0.0', description: 'Platform Version')
        string(name: 'PLATFORM_REVISION', defaultValue: 'abc123', description: 'Platform Revision')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Build') {
            steps {
                script {
                    sh '''
                        GOOS=${GOOS} GOARCH=${GOARCH} go build \
                        -ldflags="-X 'main.Version=${PLATFORM_VERSION}' -X 'main.Revision=${PLATFORM_REVISION}'" \
                        -o '${WORKSPACE}/platform'
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
