package godocker

import (
	"github.com/mercadolibre/gomemcacheclient"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestDeleteAndPullMemcache(t *testing.T) {
	err := removeImage("memcached")
	assert.Nil(t, err)

	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEmpty(t, string(container))

	err = container.KillRemove()
	assert.Nil(t, err)
}

func TestPullKillAndRemoveInvalidImageAndID(t *testing.T) {
	container, err := StartContainer("jota-rules")
	assert.Equal(t, "Error pulling image 'jota-rules': exit status 1", err.Error())
	assert.Empty(t, string(container))

	err = ContainerID("jota-rules").Kill()
	assert.Equal(t, "Error killing jota-rules: exit status 1", err.Error())

	err = ContainerID("jota-rules").Remove()
	assert.Equal(t, "Error removing jota-rules: exit status 1", err.Error())
}

func TestStartAndUseMemcache(t *testing.T) {
	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEmpty(t, string(container))

	ip, err := container.IP()
	assert.Nil(t, err)
	assert.True(t, len(ip) >= len("0.0.0.0"))

	port, err := container.GetPort("11211")
	assert.Nil(t, err)
	assert.NotEmpty(t, port)

	server := ip + ":" + port
	memcacheClient := new(gomemcacheclient.MemcacheClient)
	err = memcacheClient.ConnectClient([]string{server})
	assert.Nil(t, err)

	var bar string = "bar"
	err = memcacheClient.Set("foo", &bar)
	assert.Nil(t, err)

	bar = ""
	err = memcacheClient.Get("foo", &bar)
	assert.Nil(t, err)
	assert.Equal(t, bar, "bar")

	err = container.KillRemove()
	assert.Nil(t, err)
}

func TestContainerIP(t *testing.T) {
	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEmpty(t, string(container))

	ip, err := containerIP(string(container))
	log.Printf("This should be the container IP: %q", ip)
	assert.Nil(t, err)
	assert.NotEqual(t, "", ip)

	err = container.Kill()
	assert.Nil(t, err)

	err = container.Remove()
	assert.Nil(t, err)
}

func TestGetInvalidPort(t *testing.T) {
	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEmpty(t, string(container))

	port, err := container.GetPort("666")
	log.Printf("This should be nothing: %q", port)
	assert.Equal(t, "Error getting port mapping for 666: exit status 1", err.Error())
	assert.Empty(t, port)

	err = container.KillRemove()
	assert.Nil(t, err)
}

func TestRemoveImageInUse(t *testing.T) {
	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEmpty(t, string(container))

	err = removeImage("memcached")
	assert.Equal(t, "Error removing image 'memcached': exit status 1", err.Error())

	err = container.KillRemove()
	assert.Nil(t, err)
}

func TestRunInvalidImage(t *testing.T) {
	container, err := run("jota-rules")
	assert.Equal(t, "Error running jota-rules: exit status 1", err.Error())
	assert.Empty(t, string(container))
}

func TestParseInvalidDockerHost(t *testing.T) {
	realDockerHost := os.Getenv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "")

	ip, err := DockerIP()
	assert.Equal(t, "Error getting docker IP: Can't parse $DOCKER_HOST ()", err.Error())
	assert.Empty(t, ip)

	os.Setenv("DOCKER_HOST", realDockerHost)
}

func TestGetIPForInvalidContainer(t *testing.T) {
	ip, err := containerIP("jota-rules")
	assert.Equal(t, "Error getting 'jota-rules' IP: exit status 1", err.Error())
	assert.Empty(t, ip)
}
