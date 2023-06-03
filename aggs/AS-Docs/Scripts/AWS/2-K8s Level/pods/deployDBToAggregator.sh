#!bin/bash
cd redis
kubectl apply -k .
cd ../verifier
kubectl apply -f .
sleep 2
cd ../aggregator
kubectl apply -f .
sleep 2
cd ../../deployment
sleep 2
kubectl apply -f .
