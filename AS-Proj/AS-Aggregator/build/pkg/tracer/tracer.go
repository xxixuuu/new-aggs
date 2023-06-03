package tracer

//
type RoundBuffer struct {
	Sign [][]byte
}

func UpdateBuffer(lbuf *RoundBuffer, cbuf *RoundBuffer) {
	copy(lbuf.Sign, cbuf.Sign)
}

type Feedback struct {
	AID         []byte
	Instruction []byte
	VSIGN       []byte
}

//--obsolete from here --//

type TrackingStateParams struct {
	IsTracking      int    //0,1,2
	LastSPOREHeader int    //status
	Splitting       uint32 // default = 2
	LastVerify      bool
	TargetArray     []uint32
	AccumedIndex    uint32
}

func NewTrackingStateParams() *TrackingStateParams {
	res := new(TrackingStateParams)
	res.IsTracking = 0
	res.LastSPOREHeader = 0
	res.Splitting = 2
	res.LastVerify = false
	res.TargetArray = nil
	res.AccumedIndex = 0
	return res
}

// func CheckIfStartTracking(tp *TrackingStateParams, chFeedbackBuffer <-chan *Feedback, childMap *infra.ChildMap) {
// 	//if tracking started
// 	if len(chFeedbackBuffer) > 0 && childMap.CountAlive() > 0 {
// 		fmt.Println("-T:Feedback Recived")
// 		if tp.IsTracking == 0 {
// 			tp.IsTracking = 2
// 			fmt.Println("-Waiting for last round finish")
// 		}
// 		//Recive instruction
// 		fb := <-chFeedbackBuffer
// 		switch cc := fb.CondCode; cc {
// 		case 2: //change splitting
// 			recSpl := binary.LittleEndian.Uint32(fb.Optional[0:4])
// 			tp.Splitting = recSpl
// 		case 3: //Revoke
// 			sid := binary.LittleEndian.Uint32(fb.Optional[0:4])
// 			childMap.Set(sid, nil)
// 		}
// 	}

// 	//Start tracking
// 	if tp.IsTracking == 2 && tp.LastSPOREHeader >= 2 {
// 		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
// 		tp.IsTracking = 1
// 		fmt.Println("-Start tracking:")
// 		fmt.Println("-Initializing....")

// 		//prepare target array
// 		tp.TargetArray, tp.AccumedIndex = CreateTargetSignerArray(tp.LastVerify, 0, 0, int(tp.Splitting), childMap)
// 		fmt.Println("-T:Target Map created!")
// 		for i := range tp.TargetArray {
// 			fmt.Println(i, "[", tp.TargetArray[i], "]")
// 		}
// 	}
// }

// func ReactToFeedback(tp *TrackingStateParams, fb Feedback, childMap *infra.ChildMap) {
// 	if fb.Result == 1 {
// 		tp.LastVerify = true
// 	} else {
// 		tp.LastVerify = false
// 	}
// 	switch cc := fb.CondCode; cc {
// 	case 2: //change splitting
// 		fmt.Println("Change splliting")
// 		recSpl := binary.LittleEndian.Uint32(fb.Optional[0:4])
// 		tp.Splitting = recSpl
// 		break
// 	case 3: //Revoke
// 		sid := binary.LittleEndian.Uint32(fb.Optional[0:4])
// 		fmt.Println("Revoke!")
// 		childMap.Set(sid, nil)
// 		fmt.Println("-T: Signer:", sid, " revoked!. Back to normal mode now")
// 		tp.IsTracking = 0
// 		//initialize variables for next tracking
// 		tp.Splitting = 2 // default = 2
// 		// lastVerify = false
// 		tp.TargetArray = nil
// 		tp.AccumedIndex = 0
// 		break
// 	default:
// 		fmt.Println("-T:Feedback normal, result", fb.Result)
// 		break
// 	}
// 	if (fb.CondCode != 3) || childMap.CountAlive() > 1 { //|
// 		tp.TargetArray, tp.AccumedIndex = CreateTargetSignerArray(tp.LastVerify, len(tp.TargetArray), tp.AccumedIndex, int(tp.Splitting), childMap)
// 	} else {
// 		fmt.Println("Agggregator: All signers are dead, waiting for new connection...")
// 	}
// }

// //
// func PrintFeedback(fb Feedback) {
// 	fmt.Println("Feedback{")
// 	fmt.Println("Aggnum:", fb.AggNum)
// 	fmt.Println("Result:", fb.Result)
// 	fmt.Println("CondCode:", fb.CondCode)
// 	fmt.Printf("--OP:%v\nTS:%s\nSIGN:%v\n", fb.Optional, string(fb.TimeStamp[:]), fb.VSIGN)
// }

// func (fb *Feedback) ReciveFeedback(op *net.TCPConn, pk ...ed25519.PublicKey) (err error) {
// 	//fmt.Println("-T:Reciving Feedback...")
// 	HeaderBytes := 9
// 	for {
// 		//TCP Recive
// 		HeaderData := make([]byte, HeaderBytes)

// 		readbyte, err := op.Read(HeaderData) //here
// 		if err != nil && err != io.EOF {
// 			// fmt.Println("Reading err")
// 			// fmt.Println(err)
// 			return err
// 		}

// 		if len(HeaderData) == 0 {
// 			fmt.Println("No Responce from Verifier")
// 			return nil
// 		}
// 		//fmt.Printf("Header bytes:%v", HeaderData)
// 		if readbyte == HeaderBytes {
// 			if err != nil && err != io.EOF {
// 				fmt.Println(err)
// 				return err
// 			}
// 			fb.AggNum = binary.LittleEndian.Uint32(HeaderData[0:4])
// 			fb.Result = HeaderData[4]
// 			fb.CondCode = binary.LittleEndian.Uint32(HeaderData[5:])
// 			var BodyBytes int
// 			if fb.CondCode == 0 { //!= 0 && !=1	//None optional
// 				BodyBytes = 96
// 			} else {
// 				BodyBytes = 128
// 			}
// 			netData := make([]byte, BodyBytes)
// 			readBodyByte, err := op.Read(netData) //here
// 			if err != nil && err != io.EOF || len(netData) == 0 || readBodyByte != BodyBytes {

// 				fmt.Println("Reading err ")
// 				fmt.Println(err)
// 				return err
// 			}

// 			if fb.CondCode == 0 { //!= 0 && !=1	//None optional
// 				copy(fb.TimeStamp[:], netData[:32])
// 				copy(fb.VSIGN[:], netData[32:])
// 				if ed25519.Verify(pk[0], append(HeaderData, netData[:32]...), fb.VSIGN[:]) == true {
// 					//utils.ServerLog("Vrifier verify succeed!!")
// 				} else {
// 					netcp.ServerLog("Vrifier verify failed!! Warning !! Somebody is Pretending itself is a Verifier!!!")
// 				}
// 			} else { //==0	//other case
// 				copy(fb.Optional[:], netData[0:32])
// 				copy(fb.TimeStamp[:], netData[32:64])
// 				copy(fb.VSIGN[:], netData[64:])
// 				switch CC := fb.CondCode; CC {
// 				case 1:
// 					if ed25519.Verify(fb.Optional[:], append(HeaderData, netData[:64]...), fb.VSIGN[:]) == true {
// 						netcp.ServerLog("Vrifier verify succeed!!")
// 					} else {
// 						netcp.ServerLog("Vrifier verify failed!! Warning !! Somebody is Pretending itself is a Verifier!!!")
// 					}
// 				default:
// 					if ed25519.Verify(pk[0], append(HeaderData, netData[:64]...), fb.VSIGN[:]) == true {
// 						netcp.ServerLog("Vrifier verify succeed!!")
// 					} else {
// 						netcp.ServerLog("Vrifier verify failed!! Warning !! Somebody is Pretending itself is a Verifier!!!")
// 					}
// 				}
// 			}
// 			return err
// 		} else if readbyte > 0 {
// 			fmt.Println("In ReciveECDSA: Didint ReadExact. Wrong Format")
// 			return nil
// 		}
// 	}
// }

// func CreateTargetSignerArray(lastVerify bool, TargetRange int, lastAccumedIndex uint32, splitWay int, childMap *infra.ChildMap) ([]uint32, uint32) { //signerArray,lastAccumedIndex

// 	//Decide range
// 	//Count all alives
// 	fmt.Println(lastAccumedIndex)
// 	if TargetRange == 0 { //initial state
// 		Alives := childMap.CountAlive()
// 		if Alives == 1 {
// 			TargetRange = 1
// 		} else {
// 			TargetRange = int(math.Floor(float64(Alives) / float64(splitWay)))
// 		}

// 	} else {
// 		if !lastVerify {
// 			lastAccumedIndex = uint32(childMap.ReverseWalk(int(lastAccumedIndex), TargetRange))
// 			TargetRange = int(math.Floor(float64(TargetRange) / float64(splitWay)))
// 			fmt.Println(lastAccumedIndex, TargetRange)
// 		}
// 		if childMap.CheckAlivesAhead(lastAccumedIndex)-1 == TargetRange {
// 			TargetRange++
// 		}
// 	}
// 	fmt.Println("range:", TargetRange)
// 	//get signerMap
// 	targetArray := childMap.GetAlivesAhead(lastAccumedIndex, TargetRange)
// 	return targetArray, targetArray[len(targetArray)-1]
// }
