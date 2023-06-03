#!bin/bash
kubectl delete pods verifier-aggs
kubectl delete pods aggregator-aggs
kubectl delete deployment signer-dummy-deployment