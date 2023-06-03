Mode Description -add 
mode: 0 -Timewindow aggregation,
mode: 1 -Count Threashold aggregation,
mode: 2 -Mixed Aggregation,

Aggregator Setting Template 
{
  "Mode":                       0,
  "AID":                        125,
  "Threads":                    1, 
  "ParentAddress":              "localhost",
  "ParentPort":                 "3000",
  "AggregatorListenPort":       "3001",
  "AggregatorBufferSize":       10,
  "ParentKeepAlive":            0,
  "TimeWindow":                 200,
  "CountThreashold":            0,
  "ChildEntityType":            1,
  "MaxMultiHandler":            1,
  "UplinkSpeed":                0,
  "DownlinkSpeed":              0,
  "FeedbackReadDeadline":       10,
  "AVFormatJSON": {
    "StateHeader":             "SPORE",
    "AIDSpace":                 1,
    "SIDSpace":                 1,
    "MessageHeaderLength":      2,
    "SignerOptionHeaderLength": 1,
    "SignerOption":             4,
    "SignCurveParameter":       5
  }
}


Aggregator OS Arguments Settings

os.Args[1] = VerifierAddr 
os.Args[2] = VerifierListenPort 
os.Args[3] = AggregatorListenPort 
os.Args[4] = Mode
os.Args[5] = FeedbackReadDeadline

if os.Args[index] = "-" then follow the json file settings