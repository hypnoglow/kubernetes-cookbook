#!/bin/bash
# Hacky stuff. Better use skaffold.

set -euo pipefail
set -x # For debugging purposes.

timestamp=$(date +%s)
image=grpc-server:${timestamp}

docker build -t ${image} -f ./server/Dockerfile .

docker save ${image} | (eval $(minikube docker-env) && docker load)

cat ./server/deploy.yaml | sed "s#image: \"grpc-server\"#image: \"${image}\"#" | kubectl apply -f -
