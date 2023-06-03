#!bin/bash
export PATH=$PATH:/usr/local/go/bin
go run cmd/signer/SMain.go $@ 
#args: 1AggregatorAddr, 2AggregatorPort, 3SignerListenPort, 4SignerBufferSize 5Mode, 6 KeepAlive(0 default) 7 ttf(timetillfall)[s], 8 threads(-1 = nolimit 1= single threads, when 1 then not able to use manual interaction)
#mode:0 normal mode, 1 false mode, 2 manual false mode, 3 normal -> false mode,