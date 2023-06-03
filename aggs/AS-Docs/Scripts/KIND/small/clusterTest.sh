#!/bin/bash

sudo kind create cluster -v10 --config ./singletest.yaml

sudo kind load docker-image redis:latest
sudo kind load docker-image aggs2:dummy1
sudo kind load docker-image aggs2:signer1
sudo kind load docker-image aggs2:aggregator1
sudo kind load docker-image aggs2:verifier1

cd ../pods/redis
sudo kubectl apply -k .
sudo kubectl get nodes --show-labels
sudo kubectl get pods -o wide
sudo kind export kubeconfig
sudo kube-universe serve