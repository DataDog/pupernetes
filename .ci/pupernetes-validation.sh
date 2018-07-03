#!/usr/bin/env bash

set -exo pipefail

cd $(dirname $0)
kubectl get no -o wide
kubectl get svc,ep,ds,deploy,job,po --all-namespaces -o wide

curl -fv "http://127.0.0.1:${PUPERNETES_API:-8989}/metrics"
curl -fv "http://127.0.0.1:${PUPERNETES_API:-8989}/ready" -w '\n'

kubectl apply -f pupernetes-validation.yaml

set +e
while true
do
    kubectl get no
    kubectl get svc,ep,ds,deploy,job,po --all-namespaces -o wide

    # Job is completed and succeeded
    kubectl get po -n validation -o json | jq -re '.items[] | select(.status.phase=="Succeeded")' && break

    # Job is completed and succeeded but we didn't have time to collect its stats.phase because p8s already cleaned up
    curl -fs "http://127.0.0.1:${PUPERNETES_API:-8989}/ready" -w '\n' || {
        echo "pupernetes is already stopped"
        break
    }
    sleep 1
done

exit 0