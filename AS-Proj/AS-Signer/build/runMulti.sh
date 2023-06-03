#!bin/bash
for i in {3003..3034}
do
    go run SMain.go - 3002  $i  0 - $(($i-3002)) &
done
# go run SMain.go - 3002 3011 1 -  $((3011-3002))
