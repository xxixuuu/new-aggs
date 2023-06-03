Mode Description
-add mode : int 

0 = timeStamp, 

1 = designated length randomBytes, 

2 = designated length randomBytes with range

Dummy Setting Template
{
    "Mode":             1,
    "DataLength":       7,
    "DataLengthRange":  10,
    "Interval":        1000,
    "SignerAddr":      "localhost",
    "SignerPort":       "3002",
    "DSFormatJSON": {
      "MessageHeaderLength": 2
    }
}

Dummy OS Arguments Settings
 
os.Args[1] = SignerAddr

os.Args[2] = SignerPort

os.Args[3] = Mode

os.Args[4] = Interval

os.Args[5] = DataLength


if os.Args[index] = "-" then follow the json file settings
