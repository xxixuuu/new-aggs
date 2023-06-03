#!/bin/bash
kubectl logs verifier-aggs > verifier.txt
kubectl logs aggregator-aggs > aggregator.txt
kubectl logs signer-aggs-1 signer > signer1.txt

mv *.txt /mnt/c/Users/kores/Desktop/Lab/NPro/AWS2/10minTest2/
