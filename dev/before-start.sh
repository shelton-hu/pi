#! /bin/bash

cd $(dirname $0)
dir=$(pwd)

export $(cat .env)

mkdir -p ${DATA_PATH}/config
cp -r redis.conf etcdkeeper-index.html ${DATA_PATH}/config/
