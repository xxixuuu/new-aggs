#!bin/bash
sudo kubectl delete deployment signer-dummy-deployment | echo "q" 
sudo kubectl delete pods aggregator-aggs | echo "q" 
sudo kubectl delete pods verifier-aggs | echo "q" 