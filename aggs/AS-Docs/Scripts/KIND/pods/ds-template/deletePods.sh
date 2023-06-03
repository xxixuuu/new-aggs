#!bin/bash
for i in {0..256}
do
    sudo kubectl delete pods signer-aggs-$i --grace-period=0 --force
done