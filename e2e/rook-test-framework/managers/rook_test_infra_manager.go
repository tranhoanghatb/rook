package managers

import (

	"github.com/dangula/rook/e2e/rook-test-framework/contracts"
	"github.com/dangula/rook/e2e/rook-test-framework/enums"
	"errors"
	"github.com/dangula/rook/e2e/rook-test-framework/transport"
	"github.com/dangula/rook/e2e/rook-test-framework/objects"
	"fmt"
	"bytes"
	"os/exec"

	"io/ioutil"
	"os"
	"sync"
)

type rookTestInfraManager struct {
	transportClient contracts.ITransportClient
	platformType enums.RookPlatformType
	dockerized bool
	dockerContext *objects.DockerContext
}


var (
	r *rookTestInfraManager
	once sync.Once
)


func GetRookTestInfraManager(platformType enums.RookPlatformType, isDockerized bool) (error, *rookTestInfraManager) {
	var transportClient contracts.ITransportClient
	var dockerContext objects.DockerContext
	var dockerized bool = isDockerized

	if r != nil {
		return nil, r
	}

	//this is needed when test development vs boot2docker
	//dockerEnv := []string {
	//	"DOCKER_TLS_VERIFY=1",
	//	"DOCKER_HOST=tcp://192.168.99.100:2376",
	//	"DOCKER_CERT_PATH=/Users/tyjohnson/.docker/machine/machines/default",
	//	"DOCKER_MACHINE_NAME=default"}

	switch {
	case platformType == enums.Kubernetes:
		transportClient = transport.CreateNewk8sTransportClient()
	case platformType == enums.StandAlone:
		return errors.New("Unsupported Rook Platform Type"), r
	default:
		return errors.New("Unsupported Rook Platform Type"), r
	}

	once.Do(func() {
		dockerEnv := []string {}

		if isDockerized {
			dockerContext = objects.SetDockerContext(transport.CreateDockerClient(dockerEnv))
		}

			r = &rookTestInfraManager{
				platformType: platformType,
				transportClient: transportClient,
				dockerized: dockerized,
				dockerContext: &dockerContext,
			}
	})

	return nil, r

}

func (r *rookTestInfraManager) ValidateAndPrepareEnvironment() error	{
	containerId := r.dockerContext.Get_ContainerId()
	if  containerId != "" && r.isContainerRunning(containerId) {
		return nil
	}

	//validate docker is available

	//verify docker container is not already running
	//execute command to init docker container
	_, dockerClient := r.dockerContext.Get_DockerClient()

	cmd := []string {
		"--rm", "-itd", "--net=host", "-e=\"container=docker\"", "--privileged", "--security-opt=seccomp:unconfined",
		"--cap-add=SYS_ADMIN", "-v", "/dev:/dev", "-v","/sys:/sys", "-v", "/sys/fs/cgroup:/sys/fs/cgroup", "-v", "/sbin/modprobe:/sbin/modprobe",
		"-v", "/lib/modules:/lib/modules:rw", "-v", "/var/run/docker.sock:/tmp/docker.sock", "-p", "8080", "-P", "quay.io/quantum/rook-test", "/sbin/init",
	}

	stdout, stderr, err := dockerClient.Run(cmd)

	if err != nil {
		return fmt.Errorf("%v --> %v --> %v", err, errors.New(stdout), errors.New(stderr))
	}

	//save containerId to struct --> TODO fix
	r.dockerContext.Set_ContainerId(stderr)
	containerId = stderr

	stdout, stderr, exitCode := dockerClient.Execute([]string{containerId, "sleep", "10"})



	stdout, stderr, exitCode = dockerClient.Execute([]string{containerId, "docker", "info"})

	stdout, stderr, exitCode = dockerClient.Execute([]string{containerId, "rm", "-rfv", "/var/run/docker.sock"})

	fmt.Print(exitCode)
	//STEP 1 --> Create symlink from /docker.sock to /var/run/docker.sock
	stdout, stderr, err = dockerClient.Execute([]string{containerId, "ln", "-s", "/tmp/docker.sock", "/var/run/docker.sock"})



	r.dockerContext.Set_ContainerId(stderr)


	//STEP 2 --> Bring up k8s cluster
	//download script to container
	stdout, stderr, err = dockerClient.Execute([]string{containerId, "curl", "-o", "dind-cluster-v1.5.sh",
		"https://raw.githubusercontent.com/Mirantis/kubeadm-dind-cluster/master/fixed/dind-cluster-v1.5.sh",
		})


	//chmod +x
	stdout, stderr, err = dockerClient.Execute([]string{containerId, "chmod", "+x", "dind-cluster-v1.5.sh"})


	//run script
	stdout, stderr, err = dockerClient.Execute([]string{containerId, "./dind-cluster-v1.5.sh", "up"})


	//stdout, stderr, err = dockerClient.Stop([]string{containerId})
	//STEP 3 --> Untaint master node
	k8sClient := transport.CreateNewk8sTransportClient()

	stdout, stderr, err = k8sClient.ExecuteCmd([]string{"taint", "nodes", "--all", "dedicated-"})

	//STEP 4 --> Drain node 2 --> TODO: fix script not to create 1st and 2nd node
	stdout, stderr, err = k8sClient.ExecuteCmd([]string{"drain", "kube-node-2", "--force", "--ignore-daemonsets"})
	// kubectl drain kube-node-2 --force --ignore-daemonsets

	//STEP 5 --> Delete 2nd unneeded node --> TODO: fix script not to create 1st and 2nd node
	stdout, stderr, err = k8sClient.ExecuteCmd([]string{"delete", "node", "kube-node-2", "--force"})
	// kubectl delete node kube-node-2 --force


	//STEP 6 --> Patch controller --> TODO: pre-patch image
	goPath := os.Getenv("GOPATH")
	bytes, err := ioutil.ReadFile(goPath + "/src/github.com/dangula/rook/e2e/pod-specs/kube-controller-manager.json")
	kubeController := string(bytes)

	stdout, stderr, err = dockerClient.Execute([]string{containerId, "bash", "-c", "echo '" + kubeController + "' > kube-controller-manager.json"})

	// yes | cp -rf kube-controller-manager.json $(find /var/lib/docker/aufs/mnt -type f -name kube-controller-manager.json)
	//stdout, stderr, err = dockerClient.Execute([]string{containerId, "bash", "-c", "yes", "|", "cp", "-rf", "kube-controller-manager.json", "$(find /var/lib/docker/aufs/mnt -type f -name kube-controller-manager.json)"})


	//STEP 7 --> Install Ceph --> TODO fix so images are already patched with ceph
	//curl --unix-socket /var/run/docker.sock http:/containers/json | jq -r '.[].Id' | xargs -i docker exec -i {} bash -c 'apt-get -y update && apt-get install -qqy ceph-common'


	return nil
}

func (r *rookTestInfraManager) InstallRook(tag string) (error, client contracts.Irook_client)	{
	//if k8
	//STEP 1 --> Create rook operator
	goPath := os.Getenv("GOPATH")
	rookOperatorPath := goPath + "/src/github.com/dangula/rook/e2e/pod-specs/rook-operator.yaml"

	raw, _ := ioutil.ReadFile(rookOperatorPath)

	rawUpdated := bytes.Replace(raw, []byte("#IMAGE_PATH#"), []byte(tag), 1)
	//rookOperator := string(rawUpdated)

	ioutil.WriteFile("temp_rook-operator.yaml", rawUpdated, 0644)

	stdOut, stdErr, exit := r.transportClient.Create([]string{ "temp_rook-operator.yaml"}, []string{})

	if exit != 0 {
		fmt.Println(stdOut + stdErr)
	}
	// create pod spec
	//wait for up

	//STEP 2 --> Create rook cluster
	rookCluster := goPath + "/src/github.com/dangula/rook/e2e/pod-specs/rook-cluster.yaml"

	stdOut, stdErr, exit = r.transportClient.Create([]string{rookCluster}, []string{})

	if exit != 0 {
		fmt.Println(stdOut + stdErr)
	}
	//create pod spec
	//wait for up

	//STEP 3 --> Create rook client
	rookClient := goPath + "/src/github.com/dangula/rook/e2e/pod-specs/rook-client.yaml"

	stdOut, stdErr, exit = r.transportClient.Create([]string{rookClient}, []string{})

	if exit != 0 {
		fmt.Println(stdOut + stdErr)
	}
	//create pod spec
	//wait for up


	return nil, nil
}

func (r *rookTestInfraManager) isContainerRunning(containerId string) bool {
	err, dockerClient := r.dockerContext.Get_DockerClient()

	if err != nil {
		return false
	}

	stdOut, stdErr, err := dockerClient.Execute([]string {"ps", "--filter", "\"status=running\"", "\"id=" + containerId +"\"", "--format \"{{.ID}}\""})

	fmt.Println(stdOut, stdErr)

	return true
}

func (r rookTestInfraManager) TearDownRook(client contracts.Irook_client) error	{

	return nil
}

func (r rookTestInfraManager) TearDownInfrastructureCreatedEnvironment() error {
	return nil
}

func (r rookTestInfraManager) isRookInstalled() bool {
	return false
}

func (r rookTestInfraManager) CanConnectToDocker() bool {
	return false
}

func (r rookTestInfraManager) CanConnectToK8s() bool {
	return false
}

func (r rookTestInfraManager) pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	// Collect the output from the command(s)
	var output bytes.Buffer
	var stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error

		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}