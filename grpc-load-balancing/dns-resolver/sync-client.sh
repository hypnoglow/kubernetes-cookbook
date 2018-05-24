#!/bin/bash
# Hacky stuff. Better use skaffold.

set -euo pipefail
set -x # For debugging purposes.

timestamp=$(date +%s)
image=grpc-client:${timestamp}

docker build -t ${image} -f ./client/Dockerfile .

docker save ${image} | (eval $(minikube docker-env) && docker load)

cat ./client/deploy.yaml | sed "s#image: \"grpc-client\"#image: \"${image}\"#" | kubectl apply -f -
