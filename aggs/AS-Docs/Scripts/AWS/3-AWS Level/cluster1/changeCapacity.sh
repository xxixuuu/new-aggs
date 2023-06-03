#!/bin/bash
eksctl scale nodegroup --cluster=aggs-cl1 --nodes=$1 --name=SignerDummy --nodes-min=$1 --nodes-max=$1
