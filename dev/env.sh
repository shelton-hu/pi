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
        DATA_PATH="/var/root/.dev-data"
    else
        DATA_PATH="/Users/$(whoami)/.dev-data"
    fi
else
    if [ $(whoami) = "root" ];then
        DATA_PATH="/root/.dev-data"
    else
        DATA_PATH="/home/$(whoami)/.dev-data"
    fi
fi

cp env.tmpl .env
sed -i '' "s!\${DATA_PATH}!${DATA_PATH}!g" .env
