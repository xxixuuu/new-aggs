#Dummy in following arguments "-" = use setting on json file sync
#[1] = SignerAddress
#[2] = SignerPort
#[3] = Mode:     0 = timeStamp, 1 = designated length randomBytes, 2 = designated length randomBytes with range    
#[4] = MessageInterval
#[5] = DataLength
#new [6] = MaxDataOutPut
#--non-local
go run ./cmd/dummy/DMain.go (ip) 3000 0 - - 600
#--local
go run DMain.go - 3003 0 - - 600

#Signer in following arguments "-" = use setting on json file
#[1] = AggregatorAddress
#[2] = AggregatorPort
#[3] = ListenPort
#[4] = Mode:     0 normal mode, 1 false mode, 2 manual timing false mode,  (only work on obsolete multi-thread mode) 3 normal -> false mode with falltime,
#[5] = TimeTillFall
#[6] = SID
#--non-local
go run ./cmd/signer/SMain.go (ip) 3000  -  -  - (sid)
#--local
go run SMain.go - 3002  3003  0 - (sid)

#Aggregator in following arguments "-" = use setting on json file
#[1] = ParentAddress
#[2] = ParentPort
#[3] = ListenPort
#[4] = Mode 0 -Timewindow aggregation, 1 -Count Threashold aggregation,  2 -Mixed Aggregation
#new [5] = Time window
#new [6] = Signer Handler
#new [7] = Tracking Method
#--non-local
go run ./cmd/aggregator/AMain.go (ip) 3000 -  0 10 8 0
#--local
go run AMain.go - 3001 3002 0 10 1 0

#Verifier in following arguments  "-" = use setting on json file
#[1] = RedisAddress
#[2] = ListenPort
#new [3] = Tracking Method
#--non-local
go run ./cmd/verifier/VMain.go - 3000 0
#--local
go run VMain.go - 3001 0

#Redis
redis-server #start redis backend
reids-cli    #start intractive cli