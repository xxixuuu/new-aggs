Mode Description
-add mode: 
0 normal mode,
1 false mode,  
2 manual timing false mode,  (only work on pbsolete multi-thread mode)
3 normal -> false mode with falltime,

Signer Setting Template
{
  "Mode":                1,
  "SID":                 125,
  "Threads":             1, 
  "AggregatorAddr":      "localhost",
  "AggregatorPort":      "3001",
  "SignerListenPort":    "3002",
  "SignerBufferSize":    10,
  "TimeTillFall":        0,     //Millisecond
  "AggregatorKeepAlive": 0,
  "SAFormatJSON": {
			"SIDSpace":                 1,
			"MessageHeaderLength":      2,
      "SignerOptionHeaderLength": 1,
			"SignerOption":             4,
			"SignCurveParameter":       5
  }
}

Signer OS Arguments Settings

os.Args[1] = AggregatorAddr
os.Args[2] = AggreagtorListenPort
os.Args[3] = SignerListenPort
os.Args[4] = Mode
os.Args[5] = TimeUntillFall
os.Args[6] = SID

if os.Args[index] = "-" then follow the json file settings