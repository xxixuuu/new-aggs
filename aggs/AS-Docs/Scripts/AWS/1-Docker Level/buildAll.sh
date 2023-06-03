#! /bin/bash
sudo echo start building? need password....:

cd ~/AS-Proj/AS-Verifier/build/ &&
bash build.sh $1 | echo "q" &

cd ~/AS-Proj/AS-Aggregator/build/ &&
bash build.sh $1 | echo "q" &

cd ~/AS-Proj/AS-Signer/build/ &&
bash build.sh $1 | echo "q" &

cd ~/AS-Proj/AS-Dummy/build/ &&
bash build.sh $1 | echo "q" &