#!/bin/bash

sudo kind create cluster --config ./localSpeedTest.yaml
sudo kind load docker-image goben:host
sudo kind load docker-image goben:client
cd ./pods/goben/hosts
sudo kubectl apply -f .
cd ../../../pv\&pvc
sudo kubectl apply -f .
sudo kubectl get nodes --show-labels
sudo kubectl get pods -o wide
sudo kube-universe serve
