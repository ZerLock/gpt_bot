#!/bin/sh
docker build . -t $1
docker run -d $1
