echo "\n\n[docker image vulnerabilitie scan]\n\n"

docker scan --accept-license aggs2:verifier$1
docker scan --accept-license aggs2:aggregator$1
docker scan --accept-license aggs2:signer$1
docker scan --accept-license aggs2:dummy$1