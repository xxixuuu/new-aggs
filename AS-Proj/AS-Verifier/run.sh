#!bin/bash
export PATH=$PATH:/usr/local/go/bin
go run cmd/verifier/VMain.go $@ #--race