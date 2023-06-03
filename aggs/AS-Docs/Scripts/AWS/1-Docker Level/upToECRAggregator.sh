#!/bin/bash
cd ../../../../AS-Aggregator/build/
bash build.sh $1
docker tag aggs2:aggregator$1 057607951461.dkr.ecr.ap-northeast-1.amazonaws.com/aggs2:aggregator$1
docker push 057607951461.dkr.ecr.ap-northeast-1.amazonaws.com/aggs2:aggregator$1
