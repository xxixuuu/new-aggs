#!/bin/sh
sudo kubectl delete deployments signer-dummy-deployment
sleep 1
sudo kubectl delete pods --all --grace-period=0 --force
cd pods/redis
sudo kubectl apply -k .
sleep 10
cd ../verifier
sudo kubectl apply -f .
sleep 10
cd ../aggregator
sudo kubectl apply -f .
sleep 10
cd ../../deployment
sudo kubectl apply -f .
# sleep 20
# sudo kubectl logs verifier-aggs  > ./deployment/hp_Desktop/v_result_500_1.txt
# sudo kubectl logs aggregator-aggs  > ./deployment/hp_Desktop/a_result_500_1.txt
