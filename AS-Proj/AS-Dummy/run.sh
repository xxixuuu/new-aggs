#!bin/bash
export PATH=$PATH:/usr/local/go/bin
echo "plz provide signer port number:"
go run ./cmd/dummy/DMain.go - $@ 0 - - 
