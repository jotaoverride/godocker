package godocker

/*
Package godocker contains helper functions for setting up and tearing down docker containers to aid in testing.
*/

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Debug, if set, prevents any container from being removed.
var Debug bool

// runLongTest checks all the conditions for running a docker container
// based on image.
func runLongTest(image string) error {

	if !haveDocker() {
		return errors.New("skipping test; 'docker' command not found")
	}

	if ok, err := haveImage(image); !ok || err != nil {

		if err != nil {
			return errors.New(fmt.Sprintln("Error running docker to check for %s: %v", image, err))
		}

		log.Printf("Pulling docker image %s ...", image)
		if err := Pull(image); err != nil {
			return errors.New(fmt.Sprintln("Error pulling %s: %v", image, err))
		}
	}

	return nil
}

// haveDocker returns whether the "docker" command was found.
func haveDocker() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func haveImage(name string) (ok bool, err error) {
	out, err := exec.Command("docker", "images", "--no-trunc").Output()

	if err != nil {
		return
	}

	return bytes.Contains(out, []byte(name)), nil
}

func run(args ...string) (containerID string, err error) {

	cmd := exec.Command("docker", append([]string{"run"}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err = cmd.Run(); err != nil {

		err = fmt.Errorf("%v%v", stderr.String(), err)
		return
	}

	containerID = strings.TrimSpace(stdout.String())

	if containerID == "" {
		return "", errors.New("unexpected empty output from `docker run`")
	}

	return
}

func KillContainer(container string) error {
	return exec.Command("docker", "kill", container).Run()
}

// Pull retrieves the docker image with 'docker pull'.
func Pull(image string) error {
	out, err := exec.Command("docker", "pull", image).CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%v: %s", err, out)
	}

	return err
}

type ContainerID string

func (c ContainerID) Kill() error {
	return KillContainer(string(c))
}

// Remove runs "docker rm" on the container
func (c ContainerID) Remove() error {
	if Debug {
		return nil
	}
	return exec.Command("docker", "rm", "-v", string(c)).Run()
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

// lookup retrieves the ip address of the container, and tries to reach
// before timeout the tcp address at this ip and given port.
func (c ContainerID) lookup(port int, timeout time.Duration) (ip string, err error) {
	ip, err = "192.168.99.100", nil //c.IP()
	if err != nil {
		err = fmt.Errorf("error getting IP: %v", err)
		return
	}
	addr := fmt.Sprintf("%s:%d", ip, port)
	fmt.Println(addr)
	err = awaitReachable(addr, timeout)
	return
}

// AwaitReachable tries to make a TCP connection to addr regularly.
// It returns an error if it's unable to make a connection before maxWait.
func awaitReachable(addr string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := net.Dial("tcp", addr)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("%v unreachable for %v", addr, maxWait)
}

const (
	memcached = "memcached"
)

// setupContainer sets up a container, using the start function to run the given image.
// It also looks up the IP address of the container, and tests this address with the given
// port and timeout. It returns the container ID and its IP address, or makes the test
// fail on error.
func SetupContainer(image string, getPort func(ContainerID) (int, error), timeout time.Duration, start func() (string, error)) (c ContainerID, ip string, port int, err error) {

	err = runLongTest(image)

	containerID, err := start()
	if err != nil {
		return
	}

	c = ContainerID(containerID)

	port, err = getPort(c)
	if err != nil {
		return
	}

	ip, err = c.lookup(port, timeout)
	if err != nil {
		c.KillRemove()
		return c, "", 0, errors.New(fmt.Sprintln("Skipping test for container %s: %s", c, err))
	}

	return
}

func startMemcache() (string, error) {
	return run("-d", "-p", ":11211", memcached)
}

func SetupMemcachedContainer() (c ContainerID, ip string, port int, err error) {
	setDockerEnv()
	return SetupContainer(memcached, getMemcachedPort, 2*time.Second, startMemcache)
}

func setDockerEnv() {
	cmd := exec.Command("docker-machine", "env", "default")

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("%v%v", stderr.String(), err)
		return
	}

	re := regexp.MustCompile(`export (.+)="(.+)"`)
	exports := re.FindAllStringSubmatch(stdout.String(), -1)

	for _, export := range exports {
		os.Setenv(export[1], export[2])
	}
}

func getMemcachedPort(c ContainerID) (port int, err error) {

	cmd := exec.Command("docker", "port", string(c))

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err = cmd.Run(); err != nil {

		err = fmt.Errorf("%v%v", stderr.String(), err)
		return
	}

	re := regexp.MustCompile(":[0-9]+")
	port, _ = strconv.Atoi(re.FindString(stdout.String())[1:])

	return
}
