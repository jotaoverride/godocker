package godocker

/*
Package godocker contains helper functions for setting up and tearing down docker containers to aid in testing.
*/

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	Debug                  bool // Debug, if set, prevents any container from being removed.
	DockerMachineAvailable bool
	pwd                    = os.Getenv("GOPATH") + "/src/github.com/mercadolibre/godocker/"
)

// dockerStart checks all the conditions for running docker and sets the environment variables.
func dockerStart() error {
	err := exec.Command(pwd + "docker-start.sh").Run()
	if err != nil {
		err = fmt.Errorf("Error starting docker-machine: %v", err)
		return err
	}
	setDockerEnv()
	return nil
}

// checkImage checks the image on the repository and pull it if necessary.
func checkImage(image string) (err error) {
	ok, err := haveImage(image)
	if !ok && err == nil {
		log.Printf("Pulling docker image '%s'... this can take a while but only happen the first time.", image)
		err = pull(image)
	}
	return err
}

// setDockerEnv sets the environment variables.
func setDockerEnv() {
	stdout, _ := exec.Command("docker-machine", "env", "default").Output()
	exports := regexp.MustCompile(`export (.+)="(.+)"`).FindAllStringSubmatch(string(stdout), -1)
	for _, export := range exports {
		os.Setenv(export[1], export[2])
	}
	return
}

// haveImage checks for the image in the repositories.
func haveImage(name string) (bool, error) {
	out, err := exec.Command("docker", "images", "--no-trunc").Output()
	if err != nil {
		return false, fmt.Errorf("Error running docker to check for image '%s': %v", name, err)
	}
	return bytes.Contains(out, []byte(name)), nil
}

// pull retrieves the docker image with 'docker pull'.
func pull(image string) error {
	err := exec.Command("docker", "pull", image).Run()
	if err != nil {
		return fmt.Errorf("Error pulling image '%s': %v", image, err)
	}
	return nil

}

// removeImage remove one image.
func removeImage(image string) error {
	err := exec.Command("docker", "rmi", image).Run()
	if err != nil {
		return fmt.Errorf("Error removing image '%s': %v", image, err)
	}
	return nil
}

// run a command in a new container.
func run(args ...string) (c ContainerID, err error) {
	stdout, err := exec.Command("docker", append([]string{"run", "-dP"}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("Error running %s: %v", strings.Join(args, " "), err)
	}
	return ContainerID(strings.TrimSpace(string(stdout))), nil
}

// killContainer kills a running container.
func killContainer(container string) error {
	err := exec.Command("docker", "kill", container).Run()
	if err != nil {
		return fmt.Errorf("Error killing %s: %v", container, err)
	}
	return nil
}

// DockerIP returns the IP address of docker-machine.
func DockerIP() (ip string, err error) {
	re := regexp.MustCompile(`tcp://(.+):`)
	if match := re.FindStringSubmatch(os.Getenv("DOCKER_HOST")); match != nil {
		ip = match[1]
	} else {
		err = fmt.Errorf("Error getting docker IP: Can't parse $DOCKER_HOST (" + os.Getenv("DOCKER_HOST") + ")")
	}
	return ip, err
}

// containerIP returns the IP address of the container. This is for Linux. TODO: Check it works
func containerIP(containerID string) (string, error) {
	out, err := exec.Command("docker", "inspect", "--format", "'{{ .NetworkSettings.IPAddress }}'", containerID).Output()
	if err != nil {
		return "", fmt.Errorf("Error getting '%s' IP: %v", containerID, err)
	}
	ip := strings.Trim(string(out), "'\n")
	if len(ip) < len("0.0.0.0") {
		return "", errors.New("No IP output from docker inspect")
	}
	return ip, nil
}
