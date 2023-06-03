Mode Description -add mode:
0 Multi thread Pairing mode,
1 Single thread Pairing mode, //havent implemented yet


{
  "Mode":                       0,
  "RedisAddress":               "localhost",
  "VerifierListenPort":         "3001",
  "VerifierBufferSize":         10,
  "MaxMultiHandler":            1,
  "TCPConnAcceptDeadline":      5,
  "VAFormatJSON": {
    "AIDSpace":                 1,
    "ConditionalCodeSpace":     4,
    "InstructionContentsSpace": 32,
    "ECDSAParams":              "EdDSA"
  }
}


Verifier OS Arguments Settings

os.Args[1] = RedisAddr 
os.Args[2] = VerifierListenPort 
os.Args[3] = Mode

if os.Args[index] = "-" then follow the json file settings