node("slave") {

    stage('prepare') {
        checkout scm
    }

    def containersDown = true

    try {
        stage('start containers') {
            sh 'make docker-up'
            containersDown = false
        }

        stage('go vet') {
            sh "make vet-in-docker"
        }

        stage('go build') {
            sh "make build-in-docker"
        }

        stage('go test') {
            sh 'make test-in-docker'
        }

        stage('stop containers') {
            sh 'make docker-down'
            containersDown = true
        }

        stage('docker build') {
            sh 'make docker-build'
        }

        if ("${env.BRANCH_NAME}".equals("master")) {
            stage('docker push') {
                sh 'make docker-push'
            }
        }

        stage('docker rmi') {
            sh 'make docker-rmi'
        }
    } finally {
        if (!containersDown) {
            stage ('stop containers') {
                sh 'make docker-down'
            }
        }
    }
}