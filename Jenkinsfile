// Rook build for Jenkins Pipelines

try {
    node("ec2-stateful") {

        def DOWNLOADDIR='~/.download'

        stage('Checkout') {
            echo 'faking a check-out'
            //checkout scm
        }

        stage('Validation') {
            echo 'faking validation'
        }

        stage('Build') {
            echo 'Simulating a build by doing a pull'
            //sh "sudo mkdir -p /to-host"
            //sh "sudo docker pull quay.io/rook/rook-client"
            //sh "sudo docker pull quay.io/rook/rook-operator"
            //sh "sudo docker pull quay.io/rook/rookd"
        }

        stage('Tests') {
            //sh "sudo apt-get install -qy golang-go"
            //sh "sudo mkdir -p ~/go/src"
            //sh "sudo mkdir -p ~/go/bin"
            sh "export $GOPATH=~/go/"
            sh "export GOROOT=/usr/local/go/"
            sh "export PATH=$GOPATH/bin:$GOROOT/bin:$PATH"

            sh "go get -u github.com/jstemmer/go-junit-report"
            sh "go get -u github.com/dangula/rook"

            //sh "cd $GOPATH/src/github.com/rook/e2e/tests/integration/smokeTest"

            //sh "go test -run TestFileStorage_SmokeTest -v | go-junit-report > file-test-report.xml"

            step([$class: 'JUnitResultArchiver', testResults: '**/target/surefire-reports/*.xml'])

        }

        stage('Cleanup') {


            deleteDir()
        }
    }
}
catch (Exception e) {
    echo 'Failure encountered'

    node("ec2-stateful") {
        echo "Cleaning docker"

        echo 'Cleaning up workspace'
        deleteDir()
    }

    exit 1
}