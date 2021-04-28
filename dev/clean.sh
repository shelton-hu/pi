#!/bin/bash

cd $(dirname $0)

export $(cat .env)

rm -rf ${DATA_PATH}
