package godocker

import (
	"github.com/mercadolibre/gomemcacheclient"
	"github.com/stretchr/testify/assert"
	"testing"
	"log"
)

func TestDeleteAndPullMemcache(t *testing.T)  {
	err := removeImage("memcached")
	assert.Nil(t, err)

	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEqual(t, "", string(container))

	err = container.KillRemove()
	assert.Nil(t, err)
}

func TestPullKillAndRemoveInvalidImageAndID(t *testing.T)  {
	container, err := StartContainer("JotaJotaRules")
	assert.NotNil(t, err)
	assert.Equal(t, "", string(container))

	err = ContainerID("JotaJotaRules").Kill()
	assert.NotNil(t, err)

	err = ContainerID("JotaJotaRules").Remove()
	assert.NotNil(t, err)
}

func TestStartAndUseMemcache(t *testing.T) {
	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotNil(t, container)

	ip, err := container.IP()
	assert.Nil(t, err)
	assert.NotNil(t, ip)

	port, err := container.GetPort("11211")
	assert.Nil(t, err)
	assert.NotNil(t, port)

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

func TestContainerIP(t *testing.T)  {
	container, err := StartContainer("memcached")
	assert.Nil(t, err)
	assert.NotEqual(t, "", string(container))

	ip, err := containerIP(string(container))
	log.Printf("This should be the container IP: %q", ip)
	assert.Nil(t, err)
	assert.NotEqual(t, "", ip)

	err = container.Kill()
	assert.Nil(t, err)

	err = container.Remove()
	assert.Nil(t, err)
}