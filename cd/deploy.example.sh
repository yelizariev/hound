#!/bin/bash
set -e
# don't do anything on a wrong request
test "$1" = "/your-secret"

DOCKER=docker.pkg.github.com/itpp-labs/hound/production
NAME=hound
DATA=$(pwd)
docker pull $DOCKER
docker stop $NAME
docker rm $NAME
docker run -d -p 6080:6080 --name $NAME -v $DATA:/data $DOCKER
