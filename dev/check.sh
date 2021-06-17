#! /bin/bash

if [ -z `which make` ];then
    echo "You need to install 'make' first."
    exit 1
fi

if [ -z `which docker` ];then
    echo "You need to install 'docker' first."
    exit 1
fi

if [ -z `which docker-compose` ];then
    echo "You need to install 'docker-compose' first."
    exit 1
fi

if [ -z `which jq` ];then
    echo "You need to install 'jq' first."
    exit 1
fi
