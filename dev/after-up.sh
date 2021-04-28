#! /bin/bash

cd $(dirname $0)

# 等待kafka-manager起来
success_code="200"
http_status_code=`curl -I -m 10 -o /dev/null -s -w %{http_code} http://127.0.0.1:9000`
while ( [ ${http_status_code} -ne ${success_code} ] )
do
    echo "waiting for kafka-manager running..."
    sleep 3s
    http_status_code=`curl -I -m 10 -o /dev/null -s -w %{http_code} http://127.0.0.1:9000`
done

# 将kafka的信息配置到kafka管理中心
curl -XPOST -d "name=pi-zookeeper&zkHosts=pi-zookeeper%3A2181&kafkaVersion=0.9.0.1&jmxUser=&jmxPass=&tuning.brokerViewUpdatePeriodSeconds=30&tuning.clusterManagerThreadPoolSize=2&tuning.clusterManagerThreadPoolQueueSize=100&tuning.kafkaCommandThreadPoolSize=2&tuning.kafkaCommandThreadPoolQueueSize=100&tuning.logkafkaCommandThreadPoolSize=2&tuning.logkafkaCommandThreadPoolQueueSize=100&tuning.logkafkaUpdatePeriodSeconds=30&tuning.partitionOffsetCacheTimeoutSecs=5&tuning.brokerViewThreadPoolSize=2&tuning.brokerViewThreadPoolQueueSize=1000&tuning.offsetCacheThreadPoolSize=2&tuning.offsetCacheThreadPoolQueueSize=1000&tuning.kafkaAdminClientThreadPoolSize=2&tuning.kafkaAdminClientThreadPoolQueueSize=1000" "http://127.0.0.1:9000/clusters"

echo "all finished"
