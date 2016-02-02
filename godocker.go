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

// runLongTest checks all the conditions for running a docker container
// based on image.
func dockerStart() (err error) {
	_, err = exec.Command(pwd+"docker-start.sh").Output()
	setDockerEnv()
	return
}

func checkImage(image string) (err error)  {
	if ok, err := haveImage(image); !ok || err != nil {
		if err != nil {
			err = errors.New(fmt.Sprintf("Error running docker to check for %s: %v", image, err))
			return err
		}
		log.Printf("Pulling docker image %s ... this can take a while but only happen the first time.", image)
		if err := pull(image); err != nil {
			err = errors.New(fmt.Sprintf("Error pulling %s: %v", image, err))
			return err
		}
	}
	return
}

func setDockerEnv() {
	stdout, _ := exec.Command("docker-machine", "env", "default").Output()
	exports := regexp.MustCompile(`export (.+)="(.+)"`).FindAllStringSubmatch(string(stdout), -1)
	for _, export := range exports {
		os.Setenv(export[1], export[2])
	}
	return
}

func haveImage(name string) (ok bool, err error) {
	out, err := exec.Command("docker", "images", "--no-trunc").Output()
	if err != nil {
		return
	}
	return bytes.Contains(out, []byte(name)), nil
}

// Pull retrieves the docker image with 'docker pull'.
func pull(image string) error {
	out, err := exec.Command("docker", "pull", image).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, out)
	}
	return err
}

// Remove runs "docker rmi" on the container
func removeImage(image string) (err error) {
	out, err := exec.Command("docker", "rmi", image).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, out)
	}
	return err
}

func run(args ...string) (c ContainerID, err error) {
	stdout, err := exec.Command("docker", append([]string{"run", "-dP"}, args...)...).Output()
	c = ContainerID(strings.TrimSpace(string(stdout)))
	return
}

func killContainer(container string) error {
	_, err := exec.Command("docker", "kill", container).CombinedOutput()
	return err
}

// DockerIP returns the IP address of docker-machine.
func DockerIP() (ip string, err error) {
	re := regexp.MustCompile(`tcp://(.+):`)
	if match := re.FindStringSubmatch(os.Getenv("DOCKER_HOST")); match != nil {
		ip = match[1]
	} else {
		err = fmt.Errorf("error getting IP: Can't parse $DOCKER_HOST ( " + os.Getenv("DOCKER_HOST") + " )")
	}
	return
}

// containerIP returns the IP address of the container.
// This is for Linux. TODO: Check it works
func containerIP(containerID string) (string, error) {
	out, err := exec.Command("docker", "inspect", "--format", "'{{ .NetworkSettings.IPAddress }}'", containerID).Output()
	if err != nil {
		return "", err
	}
	ip := strings.Trim(string(out), "'\n")
	if len(ip) < len("0.0.0.0") {
		return "", errors.New("No IP output from docker inspect")
	}
	return ip, nil
}
