package godocker

import (
	"log"
	"os"
	"testing"
	"errors"
)

func TestMain(m *testing.M) {
	log.Println("Init test")
	err := setup()
	var code int
	if err != nil {
		code = 1
	} else {
		code = m.Run()
	}
	teardown()
	os.Exit(code)
}

func setup() error {
	err := dockerStart()
	if err != nil {
		return errors.New("Docker is not installed. Skip tests")
	}
	setDockerEnv()
	return nil
}

func teardown() {
	log.Println("Test finished")
}
