#!/bin/bash

cd $(dirname $0)

OS="`uname -s`"
LINUX="Linux"
DARWIN="Darwin"

if [ "${OS}" != "${LINUX}" ] && [ "${OS}" != "${DARWIN}" ];then
  echo "os unknown"
  exit 1
fi

DATA_PATH=""

if [ ${OS} = ${DARWIN} ];then
    if [ $(whoami) = "root" ];then
        DATA_PATH="/var/root/.pi-data"
    else
        DATA_PATH="/Users/$(whoami)/.pi-data"
    fi
else
    if [ $(whoami) = "root" ];then
        DATA_PATH="/root/.pi-data"
    else
        DATA_PATH="/home/$(whoami)/.pi-data"
    fi
fi

cp env.tmpl .env
sed -i '' "s!\${DATA_PATH}!${DATA_PATH}!g" .env
