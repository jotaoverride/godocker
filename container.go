package godocker

import (
	"fmt"
	"os/exec"
	"strings"
)

type ContainerID string

// setupContainer sets up a container, using the start function to run the given image.
// It also looks up the IP address of the container, and tests this address with the given
// port and timeout. It returns the container ID and its IP address, or makes the test
// fail on error.
func StartContainer(image string) (c ContainerID, err error) {
	err = dockerStart()
	if err == nil {
		err = checkImage(image)
		if err == nil {
			containerID, err := run(image)
			if err == nil {
				return ContainerID(containerID), nil
			}
		}
	}
	return
}

func (c ContainerID) IP() (ip string, err error) {
	switch uname, _ := exec.Command("uname").Output(); strings.TrimSpace(string(uname)) {
	case "Darwin":
		ip, err = DockerIP()
	case "Linux":
		ip, err = containerIP(string(c))
	default:
		fmt.Errorf("uname not soported: %v", string(uname))
	}
	return ip, err
}

func (c ContainerID) GetPort(cPort string) (dockerPort string, err error) {
	out, err := exec.Command("docker", "inspect", "--format", `{{ (index (index .NetworkSettings.Ports "`+cPort+`/tcp") 0).HostPort }}`, string(c)).Output()
	if err != nil {
		err = fmt.Errorf("Error getting port mapping for %s: %v", cPort, err)
		return "", err
	}
	dockerPort = strings.TrimSpace(string(out))
	return dockerPort, nil
}

// KillRemove calls Kill on the container, and then Remove if there was
// no error. It logs any error to t.
func (c ContainerID) KillRemove() (err error) {
	err = c.Kill()
	if err == nil {
		err = c.Remove()
	}
	return
}

// Kill runs "docker kill" on the container
func (c ContainerID) Kill() error {
	return killContainer(string(c))
}

// Remove runs "docker rm" on the container
func (c ContainerID) Remove() (err error) {
	if !Debug {
		err = exec.Command("docker", "rm", "-v", string(c)).Run()
		if err != nil {
			err = fmt.Errorf("Error removing %s: %v", c, err)
		}
	}
	return
}
