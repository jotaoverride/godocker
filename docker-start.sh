#!/bin/sh

haveDockerMachine=`hash docker-machine`
haveDocker=`hash docker-machine`

startDockerMachine () {
    docker-machine start default > /dev/null 2>&1
    eval $(docker-machine env default) > /dev/null 2>&1
}

haveImage () {
    docker images --no-trunc | grep $1 1>/dev/null
}

pullImage () {
    echo "Pulling docker image $1"; docker pull $1
}

# main >>

DockerMachineAvailable=false

if $haveDockerMachine ; then
    DockerMachineAvailable=true
    startDockerMachine
fi

if ! $haveDocker; then
    echo "Skipping test; 'docker' command not found"
    exit 1
fi

# if ! haveImage memcached && ! pullImage memcached; then exit 1; fi