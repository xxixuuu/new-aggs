#!bin/bash
for i in {0..256}
do
    kubectl delete pods signer-aggs-$i &
done