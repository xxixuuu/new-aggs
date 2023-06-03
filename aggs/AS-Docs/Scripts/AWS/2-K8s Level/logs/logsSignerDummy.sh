#!bin/bash
for i in {0..8}
do
    kubectl logs signer-aggs-$i signer
    kubectl logs signer-aggs-$i dummy
done