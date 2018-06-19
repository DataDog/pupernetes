#!/usr/bin/env bash

set -exo pipefail

cd $(dirname $0)
kubectl get hpa,svc,ep,ds,deploy,job,po --all-namespaces -o wide
kubectl apply -f pupernetes-validation.yaml

set +e
while true
do
    curl -fs "http://127.0.0.1:${PUPERNETES_API:-8989}/ready" -w '\n' || exit 0
    kubectl get hpa,svc,ep,ds,deploy,job,po --all-namespaces -o wide
    kubectl get po -n validation -o json | jq -re '.items[] | select(.status.phase=="Succeeded")' && exit 0
    sleep 3
done
