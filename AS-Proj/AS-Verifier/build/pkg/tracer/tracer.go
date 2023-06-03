package tracer

// //------------obsolete--------------

// //-------------Feedback datastruct-----------------
// type Feedback struct {
// 	Aid  uint32 //key
// 	FbID uint32 //key
// 	Vts  string //verify time stamp
// 	//[4]byte  //Uint32
// 	Result uint8
// 	//[1]byte  //1byte
// 	CondCode uint32
// 	//[4]byte  //if nil => optional == nil //Uint32
// 	Optional  [32]byte //32byte
// 	TimeStamp [32]byte //32byte
// 	VSIGN     [64]byte //ECDSA
// }

// func NewFeedback(FbID int, aggnum uint32, result uint8, concode uint32, opt []byte) Feedback {
// 	var fb Feedback
// 	fb.Aid = aggnum
// 	fb.FbID = uint32(FbID)
// 	fb.Result = result
// 	fb.CondCode = concode
// 	fb.Vts = time.Now().String()
// 	if concode != 0 && opt != nil {
// 		///fb.Optional = opt[0:32]
// 		copy(fb.Optional[:], opt[0:32])
// 	}
// 	copy(fb.TimeStamp[:], utils.FormattedTimeRecordConst()[0:32])
// 	//dosent sign here
// 	return fb
// }

// func (feedback *Feedback) SetFeedback(ctx context.Context, db *redis.Client) {

// 	err := db.HSet(ctx, fmt.Sprint(feedback.Aid)+"fb:"+fmt.Sprint(feedback.FbID), "Vts", feedback.Vts, "Res", fmt.Sprint(feedback.Result), "CC", fmt.Sprint(feedback.CondCode), "Opt", fmt.Sprint(feedback.Optional), "TS", fmt.Sprint(feedback.TimeStamp), "VSIG", fmt.Sprint(feedback.VSIGN)).Err()
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func PrintFeedback(fb Feedback) {
// 	println("Feedback{")
// 	println("Aggnum:", fb.Aid)
// 	println("Result:", fb.Result)
// 	println("CondCode:", fb.CondCode)
// 	fmt.Printf("--%s,%d,%d\n", fb.TimeStamp, len(fb.Optional), len(fb.VSIGN))
// }

// //PKMap -- use it multithread safe lock & unlock
// type SignerMap struct {
// 	SMapID      int
// 	Aid         uint32
// 	Header      uint32 //signerNumber
// 	Sid         []uint32
// 	IsAlive     []bool
// 	Compromized []bool
// 	SpllitWay   int
// 	SearchIndex int
// 	CreatedTime [32]byte
// }

// func (signerM *SignerMap) NewSignerMap(Aid uint32, smID int) {
// 	signerM.SMapID = smID
// 	signerM.Aid = Aid
// 	signerM.Header = 0
// 	signerM.Sid = make([]uint32, 4)
// 	signerM.SpllitWay = 2
// 	signerM.SearchIndex = 0
// 	copy(signerM.CreatedTime[:], utils.FormattedTimeRecordConst()[0:32])
// }

// func (signerM *SignerMap) ReciveSignerMap(header []byte, op *net.TCPConn, ctx context.Context, db *redis.Client) error {

// 	signerM.Header = binary.LittleEndian.Uint32(header) //necessary
// 	signerM.Sid = make([]uint32, signerM.Header)
// 	signerM.IsAlive = make([]bool, signerM.Header)
// 	signerM.Compromized = make([]bool, signerM.Header)

// 	signerM.SpllitWay = 2
// 	signerM.SearchIndex = 0
// 	for i := 0; i < int(signerM.Header); i++ {
// 		var buf []byte = make([]byte, 5)
// 		_, err := op.Read(buf)
// 		if err != nil {
// 			println(err)
// 			return err
// 		}
// 		signerM.Sid[i] = binary.LittleEndian.Uint32(buf[0:4])
// 		signerM.Compromized[i] = false
// 		if buf[4] == uint8(0) {
// 			signerM.IsAlive[i] = false
// 		} else {
// 			signerM.IsAlive[i] = true
// 		}
// 	}
// 	signerM.SetSignerMap(ctx, db)
// 	return nil
// }

// func (signerMap *SignerMap) SetSignerMap(ctx context.Context, db *redis.Client) {
// 	SidArray := utils.GobEncordeOutString(signerMap.Sid)
// 	IsAliveArray := utils.GobEncordeOutString(signerMap.IsAlive)
// 	CompromizedArray := utils.GobEncordeOutString(signerMap.Compromized)

// 	err := db.HSet(ctx, fmt.Sprint(signerMap.Aid)+"sm:"+fmt.Sprint(signerMap.SMapID), "SMID", fmt.Sprint(signerMap.SMapID), "Aid", fmt.Sprint(signerMap.Aid), "Hdr", fmt.Sprint(signerMap.Header), "Sid", SidArray, "IsAl", IsAliveArray, "Czed", CompromizedArray, "Sw", fmt.Sprint(signerMap.SpllitWay), "Si", fmt.Sprint(signerMap.SearchIndex), "Ctm", fmt.Sprint(signerMap.CreatedTime)).Err()
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func (signerM *SignerMap) PrintSignerMap(AID uint32) {
// 	println("---------SignerMap-------------")
// 	println("AID:", AID)
// 	println("Signer Num:", signerM.Header)
// 	for i := range signerM.Compromized {
// 		println("--SID", signerM.Sid[i])
// 		print("--isAlive;", signerM.IsAlive[i])
// 		if signerM.IsAlive[i] == true {
// 			println(" T")
// 		} else {
// 			println(" F")
// 		}
// 		print("--isCompromized;", signerM.Compromized[i])
// 		if signerM.Compromized[i] == true {
// 			println(" T")
// 		} else {
// 			println(" F")
// 		}
// 		println("---------------")
// 	}
// 	println("---------SignerMap-------------")
// }

// func (signerM *SignerMap) CountAlive() int {
// 	res := 0
// 	for i := range signerM.IsAlive {
// 		if signerM.IsAlive[i] == true {
// 			res++
// 		}
// 	}
// 	return res
// }

// //-------------Feedback datastruct-----------------
