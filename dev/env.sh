#!/bin/bash

cd $(dirname $0)
dir=$(pwd)

OS="`uname -s`"
DATA_PATH="${dir}/.dev-data"

cp env.tmpl .env

if [ "${OS}" = "Darwin" ];then
    sed -i '' "s!\${DATA_PATH}!${DATA_PATH}!g" .env
    sed -i '' "s!\${OS}!${OS}!g" .env
elif [ "${OS}" = "Linux" ];then
    sed -i "s!\${DATA_PATH}!${DATA_PATH}!g" .env
    sed -i "s!\${OS}!${OS}!g" .env
else
    echo "Your OS [${OS}] is not supported yet."
    exit 1
fi