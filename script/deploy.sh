#!/bin/bash
set -ex
docker pull kubegames/kubegames-api:latest
docker pull kubegames/kubegames-proxy:latest
docker pull kubegames/kubegames-scheduler:latest
docker pull kubegames/web:latest
docker pull kubegames/hall:latest

docker pull kubegames/bqtp:latest
docker pull kubegames/csd:latest
docker pull kubegames/jinpingmei:latest
docker pull kubegames/jinyumantang:latest
docker pull kubegames/jsxs:latest
docker pull kubegames/jsyc:latest
docker pull kubegames/jzbf:latest
docker pull kubegames/sg777:latest
docker pull kubegames/sgxml:latest
docker pull kubegames/shz:latest
docker pull kubegames/wflm:latest
docker pull kubegames/wlzb:latest
docker pull kubegames/wszs:latest

set +ex
kubectl delete -f ./games.yaml
kubectl delete -f ./kubegames.yaml

kubectl apply -f ./games.yaml
kubectl apply -f ./kubegames.yaml