#!bin/bash
cd verifier &&
sudo kubectl apply -f . | echo "q" 
sleep 2
cd ../aggregator
sudo kubectl apply -f . | echo "q" 
cd ../../deployment
sleep 2
sudo kubectl apply -f . | echo "q"
