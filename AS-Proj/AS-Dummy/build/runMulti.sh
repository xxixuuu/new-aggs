#!bin/bash
for i in {3003..3034}
do 
    go run DMain.go - $i 0 - - 600 &
done