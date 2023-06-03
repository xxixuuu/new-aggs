#!bin/bash
kubectl get "$(kubectl api-resources --verbs=list -o name | tr " >> ./podslog.txt
" "," | sed -e 's/,$//')" -o wide --all-namespaces >> ./podslog.txt
kubectl describe ns,no,ing,svc,endpoints,deploy,po,sa,pdb --all-namespaces >> ./podslog.txt
date && kubectl get events --sort-by=.metadata.creationTimestamp --all-namespaces >> ./podslog.txt
kubectl logs verifier-aggs >> ./verifier-aggs.txt
kubectl logs aggregator-aggs >> ./aggregator-aggs.txt
mv *.txt /mnt/c/Users/kores/Desktop/Lab/NPro/AWS2/10minTest2/