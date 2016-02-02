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
	if err != nil {
		return
	}
	err = checkImage(image)
	if err != nil {
		return
	}
	containerID, err := run(image)
	if err != nil {
		return
	}
	return ContainerID(containerID), err
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
	return
}

func (c ContainerID) GetPort(cPort string) (dockerPort string, err error) {
	out, err := exec.Command("docker", "inspect", "--format", `{{ (index (index .NetworkSettings.Ports "`+cPort+`/tcp") 0).HostPort }}`, string(c)).Output()
	if err == nil {
		dockerPort = strings.TrimSpace(string(out))
	}
	return
}

// KillRemove calls Kill on the container, and then Remove if there was
// no error. It logs any error to t.
func (c ContainerID) KillRemove() error {
	if err := c.Kill(); err != nil {
		return err
	}
	if err := c.Remove(); err != nil {
		return err
	}
	return nil
}

func (c ContainerID) Kill() error {
	return killContainer(string(c))
}

// Remove runs "docker rm" on the container
func (c ContainerID) Remove() error {
	if Debug {
		return nil
	}
	return exec.Command("docker", "rm", "-v", string(c)).Run()
}
