#!/bin/bash
PUSH=$1
DEPLOY=$2

set -ex
docker build -t kubegames/web:latest .
set +ex

if [ $PUSH == "true" ];then
  set -ex
  docker push kubegames/web:latest
  set +ex
fi

if [ $DEPLOY == "true" ];then
  kubectl delete -f ./deploy.yaml
  kubectl apply -f ./deploy.yaml
fi