package aggregator

// #cgo LDFLAGS: -lcrypto
// #include <stdio.h>
// #include <openssl/sha.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"time"

	"aggregator.local/pkg/infra"
	"aggregator.local/pkg/tracer"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/xxixuuu/netcp"
	"github.com/xxixuuu/utils"
)

type AVFormatJson struct {
	StateHeader              string
	AIDSpace                 int
	SIDSpace                 int
	MessageHeaderLength      int
	SignerOptionHeaderLength int //Decides SOption Header Length
	SignerOption             int //Decides Signer Option type
	SignCurveParameter       int //Decides Signature length
}

type AggregatorProperty struct {
	Mode                 int //5 mode=0 = time window,mode =1 number threshold, mode=2 both
	Aid                  int //Aid assigned from verifier
	Threads              int
	ParentAddress        string //1 ParentAddress,
	ParentPort           string //2 ParentPort
	AggregatorListenPort string //3 AggregatorListenPort
	AggregatorBufferSize int    //4 buffer size per channel AggregatorBufferSize
	ParentKeepAlive      int    //6 if !=0 :true KeepAlive
	TimeWindow           int    //7 timewindow [ms]
	CountThreshold       int    //8 Count up threshold
	ChildEntityType      int    //9 Verifier =0, Aggregator =1
	MaxMultiHandler      int    //10 Maximum Multi-Handler goroutine
	UplinkSpeed          int64  //Auto callibrate Send void R
	DownlinkSpeed        int64  //Auto callibrate Recive Feedback twice
	MessageReadDeadline  int    //Deadline for read feedback from verifier
	FeedbackReadDeadline int    //Deadline for read feedback from verifier
	AIDAssign            int
	SIDAssign            int
	TrackingMethod       int //1 = feedback all, 2 = helper function
	RoundBufferSize      int
	AVFormat             AVFormatJson `json:"AVFormatJSON"`
}

type SignWithID struct {
	SID  int
	Sign bls.Sign
}

type RoundBuffer struct {
	VUID      int
	Signature []*SignWithID
}
type SortBy []*RoundBuffer

func (a *RoundBuffer) Len() int           { return len(a.Signature) }
func (a *RoundBuffer) Swap(i, j int)      { a.Signature[i], a.Signature[j] = a.Signature[j], a.Signature[i] }
func (a *RoundBuffer) Less(i, j int) bool { return a.Signature[i].SID < a.Signature[j].SID }

func (rb *RoundBuffer) Append(sid int, sig bls.Sign) {
	swi := new(SignWithID)
	swi.SID = sid
	swi.Sign = sig

	rb.Signature = append(rb.Signature, swi)
}

func (rb *RoundBuffer) Copy(op *RoundBuffer) { //ptr copy
	// rb = new(RoundBuffer)
	rb.VUID = op.VUID
	rb.Signature = op.Signature
	op = new(RoundBuffer)
}

func (rb *RoundBuffer) Print() { //ptr copy
	println("VUID", rb.VUID)
	for i := range rb.Signature {
		fmt.Print("i", rb.Signature[i].SID, " ")
	}
	println()
}

func RefreshRoundBuffer(rb []*RoundBuffer) {

	for i := len(rb) - 1; i > 0; i-- {
		rb[i].Copy(rb[i-1])
	}
	rb[0] = new(RoundBuffer)
}

func initRoundBuffer(rb []*RoundBuffer) {
	for i := len(rb) - 1; i >= 0; i-- {
		rb[i] = new(RoundBuffer)
	}
}

func SortAndAggregateIndependentSignWithID(rb *RoundBuffer) (res *RoundBuffer) {
	res = new(RoundBuffer)
	// println("len(rb.Signature):", len(rb.Signature))
	sort.Sort(rb)
	// rb.Print()
	// println("-------Right After Sort--------")
	var swi = new(SignWithID)
	swi.SID = 0

	for len(rb.Signature) > 0 {
		if rb.Signature[0].SID != rb.Signature[1].SID {
			if swi.SID == 0 {
				swi.Sign = rb.Signature[0].Sign
				// print("set+:")
			} else {
				swi.Sign.Add(&rb.Signature[0].Sign)
				// print("i+", rb.Signature[0].SID)
			}
			swi.SID = rb.Signature[0].SID
			res.Signature = append(res.Signature, swi)
			// println("->append:i", swi.SID)

			swi = new(SignWithID)
			swi.SID = 0
		} else {
			if swi.SID == 0 {
				swi.SID = rb.Signature[0].SID
				swi.Sign = rb.Signature[0].Sign
				// print("set+:")
			} else {
				swi.Sign.Add(&rb.Signature[0].Sign)
				swi.SID = rb.Signature[0].SID
				// print("j+", rb.Signature[0].SID)
			}
		}
		if len(rb.Signature) > 1 {
			rb.Signature = rb.Signature[1:]
			// println(len(rb.Signature))
		}

		if len(rb.Signature) == 1 {
			if swi.SID == rb.Signature[0].SID {
				swi.Sign.Add(&rb.Signature[0].Sign)
				// print("i+", rb.Signature[0].SID)
				res.Signature = append(res.Signature, swi)
				// println("->append:i", swi.SID)
				break
			} else {
				swi.SID = rb.Signature[0].SID
				swi.Sign = rb.Signature[0].Sign
				res.Signature = append(res.Signature, swi)
				// println("->append:i", swi.SID)
				break
			}
		}

	}

	res.VUID = rb.VUID
	// res.Print()
	return res
}

func hashToSHA512(message []byte) []byte {
	hash := make([]byte, C.SHA512_DIGEST_LENGTH)
	ctx := &C.struct_SHA512state_st{}
	C.SHA512_Init(ctx)
	_ = C.SHA512_Update(ctx, C.CBytes(message), C.size_t(len(message)))
	_ = C.SHA512_Final((*C.uchar)(&hash[0]), ctx)
	return hash
}

func NewAggregatorProperty(newAggregator *AggregatorProperty, osArgs []string) {
	newAggregator.Aid = 0 //change it later
	var err error
	//#args: 1ParentAddress, 2ParentPort, 3AggregatorListenPort, 4AggregatorBufferSize 5Mode, 6 KeepAlive(0 default) 7timewindow(ms) 8count up(mode1)9 ParentEntityType

	JSONFile, err := os.Open("./aconfig.json")
	if err != nil {
		fmt.Println("Error during open setting file sconfig.json:", err)
	}

	AllBytesFromJSONFile, err := ioutil.ReadAll(JSONFile)
	if err != nil {
		fmt.Println("Error read file aconfig.json:", err)
	}

	err = json.Unmarshal(AllBytesFromJSONFile, newAggregator)
	if err != nil {
		fmt.Println("Error execute json unmarshal setting from file aconfig.json:", err)
	}

	//Then configure from os Arguments
	if os.Args[1] != "-" {
		newAggregator.ParentAddress = osArgs[1]
	}
	if os.Args[2] != "-" {
		newAggregator.ParentPort = osArgs[2]
	}
	if os.Args[3] != "-" {
		newAggregator.AggregatorListenPort = osArgs[3]
	}
	if os.Args[4] != "-" {
		newAggregator.Mode, err = strconv.Atoi(osArgs[4])
		if err != nil {
			fmt.Println("In NewAggregatorProperty, Error converting os.Args[4]", err)
		}
	}
	if os.Args[5] != "-" {
		newAggregator.TimeWindow, err = strconv.Atoi(osArgs[5])
		if err != nil {
			fmt.Println("In NewSAggregatorProperty, Error converting os.Args[5]", err)
		}
	}
	if os.Args[6] != "-" {
		newAggregator.MaxMultiHandler, err = strconv.Atoi(osArgs[6])
		if err != nil {
			fmt.Println("In NewSAggregatorProperty, Error converting os.Args[6]", err)
		}
	}
	if os.Args[7] != "-" {
		newAggregator.TrackingMethod, err = strconv.Atoi(osArgs[7])
		if err != nil {
			fmt.Println("In NewSAggregatorProperty, Error converting os.Args[7]", err)
		}
	}

	fmt.Println("newAggregator Configed:", newAggregator)
}

func ReciveSAData(op *net.TCPConn, aggregatorProperty *AggregatorProperty) (*infra.MessageData, error) { //sign, message

	var err error
	var md infra.MessageData
	op.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(aggregatorProperty.MessageReadDeadline)))
	if aggregatorProperty.AVFormat.SIDSpace > 0 {
		md.SID, err = netcp.ReciveConstBytes(op, aggregatorProperty.AVFormat.SIDSpace)
		if err != nil {
			fmt.Println("In func ReciveS-AData Error:", err)
			fmt.Println("Error: In ReciveConstHeaderData: md.SID")
			fmt.Println(err)
			return nil, err
		}
	}

	op.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(aggregatorProperty.MessageReadDeadline)))
	md.MessageHeader, md.Message, err = netcp.ReciveConstHeaderData(op, aggregatorProperty.AVFormat.MessageHeaderLength)
	if err != nil {
		fmt.Println("In func ReciveS-AData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.Message")
		fmt.Println(err)
		return nil, err
	}
	op.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(aggregatorProperty.MessageReadDeadline)))
	if aggregatorProperty.AVFormat.SignerOptionHeaderLength > 0 {
		md.SignerOptionHeader, md.SignerOption, err = netcp.ReciveConstHeaderData(op, aggregatorProperty.AVFormat.SignerOptionHeaderLength)
		if err != nil {
			fmt.Println("In func ReciveSA-Data Error:", err)
			fmt.Println("Error: In ReciveConstHeaderData: md.SO")
			fmt.Println(err)
			return nil, err
		}
	} else {
		md.SignerOptionHeader = nil
		md.SignerOption = nil
	}

	var signlength int
	if aggregatorProperty.AVFormat.SignCurveParameter == 5 {
		signlength = 48
	} else {
		println("In Recive SAFormat Data: unknown pairing friendly curve:", aggregatorProperty.AVFormat.SignCurveParameter)
	}

	op.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(aggregatorProperty.MessageReadDeadline)))
	SerializedSign, err := netcp.ReciveConstBytes(op, signlength)
	if err != nil {
		fmt.Println("In func ReciveS-AData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.Sign")
		fmt.Println(err)
		return nil, err
	}

	md.Sign.Deserialize(SerializedSign)

	return &md, err
}

// Create R
func CreateR(AIDSID []byte, md *infra.MessageData, aggregatorProperty *AggregatorProperty) []byte {
	AID := AIDSID[:aggregatorProperty.AVFormat.AIDSpace]
	var SID []byte
	if aggregatorProperty.SIDAssign == 1 {
		SID = AIDSID[4 : 4+aggregatorProperty.AVFormat.SIDSpace]
		// fmt.Println("SID = ", SID)
	} else if aggregatorProperty.SIDAssign == 0 {
		SID = md.SID
		// fmt.Println("SID = ", SID)
	}
	R := append([]byte("R"), AID...)      //R+AID
	R = append(R[:], SID...)              //R+AID+SID
	R = append(R[:], md.MessageHeader...) //R+AID+SID+SignerOptionHeader+SignerOption+MessageHeader
	R = append(R[:], md.Message...)       //R+AID+SID+SignerOptionHeader+SignerOption+Message
	if md.SignerOptionHeader != nil {
		R = append(R[:], md.SignerOptionHeader...) //R+AID+SID+SignerOptionHeader
		R = append(R[:], md.SignerOption...)       //R+AID+SID+SignerOptionHeader+SignerOption
	}
	R = append(R[:], md.Sign.Serialize()...) //R+AIDSID+PK+PKSIGN

	return R

}

func CreateA(lbuf *RoundBuffer, AID []byte) (res []byte) {
	//Allfeedback(TM1, TM2)
	//Method1 send independet buffered signatures
	//Method2 send aggregated by SID buffered signatures
	res = make([]byte, 0)
	res = append(res, []byte("A")[0])
	res = append(res, AID...)
	res = append(res, utils.UintConvertToLittleEndianByteArray(uint32(len(lbuf.Signature)))...)
	for i := 0; i < len(lbuf.Signature); i++ {
		res = append(res, lbuf.Signature[i].Sign.Serialize()...)
	}
	return res
}

func CheckChildSelfSign(op *net.TCPConn, AidSid []byte, aggregatorProperty *AggregatorProperty) ([]byte, error) {
	// clientinfo := strings.Split(op.RemoteAddr().String(), ":")
	// netcp.ServerLog("Serving Entity IP: " + clientinfo[0] + ":" + clientinfo[1])
	var Sid uint32 = binary.LittleEndian.Uint32(AidSid[4:])
	fmt.Println("TCPRecvicer: Check-ChildSelfSign:  ReciveS-AData")
	op.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(aggregatorProperty.MessageReadDeadline)))
	md, err := ReciveSAData(op, aggregatorProperty)
	if err != nil {
		netcp.ServerLog("Error: In Signer handler Error") //escape here when signer terminated
		fmt.Println("SID:", Sid)
		fmt.Println(err)
		return nil, err
	}

	var pksign bls.Sign = md.Sign
	var pk bls.PublicKey

	pk.Deserialize(md.Message)
	signatureScope := append(append(append(append(md.SID, md.MessageHeader...), md.Message...), md.SignerOptionHeader...), md.SignerOption...)
	if pksign.VerifyByte(&pk, signatureScope) == true {
		fmt.Printf("Signer Self sign checked! Signer:%d no problem\n", Sid)
		return CreateR(AidSid, md, aggregatorProperty), nil //need foward R function
	} else {
		netcp.ServerLog("!!!Invalid Signer PK, PKSIGN. Intruder!!!")
		return nil, nil
	}
}

func ChildrenHandler(HandlerID int, chBufferHandlerTasks chan uint32, childMap *infra.ChildMap, handlerMap *infra.MultiHandlerMap, newAggregator *AggregatorProperty) string {
	//Regist to handlerMap
	// time.Sleep(time.Millisecond * 200 * time.Duration(HandlerID))
	handlerMap.Set(HandlerID, 0)

	//need correspond with designated channel structure
	var ChildQue []*infra.Child

	for {
		//check if any new
		if len(chBufferHandlerTasks) > 0 {
			// println("len chBufferHandlerTasks", len(chBufferHandlerTasks))
			// fmt.Println("[Handler:", HandlerID, "] Checking loads")
			if handlerMap.CheckLoadbalance(HandlerID) == true {
				if len(chBufferHandlerTasks) > 0 {
					println("HID:", HandlerID, " Trying handle new signer")
					ChildQue = append(ChildQue, childMap.Get(<-chBufferHandlerTasks))
					handlerMap.Inc(HandlerID)
					fmt.Println("[Handler:", HandlerID, "] handling Child SID:", ChildQue[len(ChildQue)-1].Sid)
				}
			}
		} else {
			if len(ChildQue) < 1 {
				// println("HID:", HandlerID, " Sleeping")
				time.Sleep(time.Millisecond * 200)
			}
		}

		//do routine
		if len(ChildQue) >= 1 {
			// println("Handler:", HandlerID, " -Starts Reading SA Message with :", len(ChildQue), "children")
			// println("HID:", HandlerID, " children:", len(ChildQue))
			for i := 0; i < len(ChildQue); i++ {
				if childMap.GetAlive(ChildQue[i].Sid) > 0 && ChildQue[i] != nil {
					// fmt.Println("Reciving from SID:", ChildQue[i].Sid, time.Now().String())

					messageSign, err := ReciveSAData(ChildQue[i].TcpConn, newAggregator)
					if err != nil {
						netcp.ServerLog(" Signer handler Error")
						fmt.Println(err)
						childMap.DeleteKey(ChildQue[i].Sid)

					} else {
						// fmt.Println("Que", ChildQue[i].Sid, " len:", len(*ChildQue[i].ChPtr))
						//Need to implement when Channel Overflow
						if newAggregator.SIDAssign == 1 {
							messageSign.SID = utils.UintConvertToLittleEndianByteArray(ChildQue[i].Sid)[:newAggregator.AVFormat.SIDSpace]
							// fmt.Println("!SID = ", messageSign.SID)
						}
						if len(*ChildQue[i].ChPtr) <= newAggregator.AggregatorBufferSize {
							*ChildQue[i].ChPtr <- messageSign
						} else {
							fmt.Println(ChildQue[i].Sid, "-Channel overflow, message discarded!")
						}

					}
					// fmt.Println("Recived from SID:", ChildQue[i].Sid, time.Now().String())

				} else {
					handlerMap.Dec(HandlerID)
					//kick out nil ptr Child from Que
					println("Kicked out Handler ID:", HandlerID)
					if len(ChildQue) == 1 {
						ChildQue = nil
					} else {
						ChildQue[i] = ChildQue[len(ChildQue)-1]
						ChildQue = ChildQue[:len(ChildQue)-1]
					}
					// fmt.Println("Dec", ChildQue)
				}
			}
		}
	}
}

func TCPRecvicer(ParentConn *net.TCPConn, localTCPListenerConn *net.TCPListener, aggregatorProperty *AggregatorProperty, chBufferHandlerTasks chan uint32, childMap *infra.ChildMap, handlerMap *infra.MultiHandlerMap) {
	var Sid uint32 = 1
	for {
		conn, err := localTCPListenerConn.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			return
		}

		//GenAIDSID
		AIDSID := utils.GenAIDSIDLittleEndian(uint32(aggregatorProperty.Aid), Sid)
		// fmt.Println("AIDSID", AIDSID)
		fmt.Println("TCPRecvicer: Checking Signer Self signature, SID:", AIDSID)
		//check signer self sign
		R, err := CheckChildSelfSign(conn, AIDSID, aggregatorProperty)
		if R == nil || err != nil {
			fmt.Println("In TCP Reciver CheckChildrenSelfSign failed...")
			fmt.Println(err)
			// return
		} else {
			fmt.Println("TCPRecvicer: sending -R-")
			_, err = ParentConn.Write(R) //
			if err != nil {
				fmt.Println(err)
			}
			//Regist NewSigner
			newChild := infra.NewChild(make(chan *infra.MessageData, aggregatorProperty.AggregatorBufferSize), conn, Sid)
			childMap.Set(Sid, newChild)
			fmt.Println("TCPRecvicer: New signer:", Sid, " Registed.")

			chBufferHandlerTasks <- Sid
			Sid++
		}

	}
}

func CreateAVFormatTransaction(StateHeader []byte, md *infra.MessageData, aggregator *AggregatorProperty, AGS []byte) []byte {
	Aid := make([]byte, 4)
	binary.LittleEndian.PutUint32(Aid, uint32(aggregator.Aid))
	T := append(StateHeader[0:1], Aid[:aggregator.AVFormat.AIDSpace]...)
	T = append(T, md.SID...)
	T = append(T, md.MessageHeader...)
	T = append(T, md.Message...)
	T = append(T, md.SignerOptionHeader...)
	T = append(T, md.SignerOption...)
	if StateHeader[0] == []byte("E")[0] || StateHeader[0] == []byte("O")[0] {
		T = append(T, AGS...)
	}
	return T
}

func ReciveVAFormatTransaction(verifierConn *net.TCPConn, verifierECDSAPublicKey ed25519.PublicKey) (SID int) {
	//
	aid, err := netcp.ReciveConstBytes(verifierConn, 4)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return 0
	}
	sid, err := netcp.ReciveConstBytes(verifierConn, 4)

	timeRecord, err := netcp.ReciveConstBytes(verifierConn, 8)

	VSIGN, err := netcp.ReciveConstBytes(verifierConn, 64)

	signatureScope := append(aid, sid...)
	signatureScope = append(signatureScope, timeRecord...)

	if ed25519.Verify(verifierECDSAPublicKey, signatureScope, VSIGN) == true {
		//utils.ServerLog("Vrifier verify succeed!!")
		return utils.ByteArrayConvertToUint(sid)
	} else {
		netcp.ServerLog("Vrifier verify failed!! Warning !! Somebody is Pretending itself is a Verifier!!!")
		return -1
	}
}

// singleton
func verifierHandler(verifierConn *net.TCPConn, aggregatorProperty *AggregatorProperty, childMap *infra.ChildMap, ch_trackingSource <-chan *RoundBuffer, ch_trackingAlart chan<- int, verifierECDSAPublicKey ed25519.PublicKey) {
	for {
		//Tracking with buffer
		FeedbackInstruction := ReciveVerifierInstruction(verifierConn, verifierECDSAPublicKey)
		targetVUID := utils.ByteArrayConvertToUint(FeedbackInstruction.Instruction[4:8])
		println("Target VUID:", targetVUID)
		ch_trackingAlart <- targetVUID

		if utils.ByteArrayConvertToUint(FeedbackInstruction.Instruction[0:4]) == 0 {
			sent, err := verifierConn.Write(CreateA(<-ch_trackingSource, utils.UintConvertToLittleEndianByteArray(uint32(aggregatorProperty.Aid))))
			println("A format sent: ", sent, "bytes")
			if err != nil {
				fmt.Println("In verifierHandler, Error:", err)
			}
		} else {
			println("case :TargetSID != 0, not implemented yet.")
		}
	}
}

func CheckBufferDumpRequirements(VerifierConn *net.TCPConn, ch_trackingAlart chan int, rb []*RoundBuffer, ch_trackingSource chan<- *RoundBuffer) {
	//check Tracking Alart
	if len(ch_trackingAlart) > 0 {
		recivedVUID := <-ch_trackingAlart
		for i := range rb {
			println("BufferedVUID", rb[i].VUID)
			if rb[i].VUID == recivedVUID {
				ch_trackingSource <- rb[i]
				println("Found Buffered Signs, VUID:", rb[i].VUID)
				break
			}
		}
	}
}

// singleton
func SyncAggregator(VerifierConn *net.TCPConn, aggregatorProperty *AggregatorProperty, childMap *infra.ChildMap) {
	bls.Init(aggregatorProperty.AVFormat.SignCurveParameter)

	status := 0
	signCounts := 0
	AGGStart := time.Now()

	netcp.ServerLog("Establishing connection with verifier ...")
	Aid, VPK := AggregatorRegistration(VerifierConn, aggregatorProperty)
	fmt.Println("AID:", Aid)

	var ags bls.Sign
	var md *infra.MessageData
	// var fb *tracer.Feedback
	var VUstart time.Time
	for {
		alives := childMap.CheckAlivesAhead(0)
		// println("SyncAggregator: Alive connected signer:", alives)
		if alives > 0 {
			println("SyncAggregator: --Start--")
			break
		} else {
			time.Sleep(time.Duration(5) * time.Millisecond)
		}
	}
	ch_trackingSource := make(chan *RoundBuffer, 1)
	ch_trackingAlart := make(chan int, 1)
	go verifierHandler(VerifierConn, aggregatorProperty, childMap, ch_trackingSource, ch_trackingAlart, VPK)

	AggregatorRoundBuffer := make([]*RoundBuffer, aggregatorProperty.RoundBufferSize)
	initRoundBuffer(AggregatorRoundBuffer)
	VerificationUnitCount := 2

	TemporaryTimeWindow := aggregatorProperty.TimeWindow

	for { //Main loop
		CheckBufferDumpRequirements(VerifierConn, ch_trackingAlart, AggregatorRoundBuffer, ch_trackingSource)

		switch status {
		case 0: //start
			VUstart = time.Now()

			// start := time.Now()
			md = childMap.GetRandomMessage()
			AggregatorRoundBuffer[0].Append(utils.ByteArrayConvertToUint(md.SID), md.Sign)
			S := CreateAVFormatTransaction([]byte("S"), md, aggregatorProperty, nil)
			// fmt.Println("S", S)
			VerifierConn.Write(S)
			signCounts = 1
			status = 1
			AGGStart = time.Now()
			// ags = md.Sign
			ags = md.Sign
			// fmt.Println("-S-took,", time.Since(start))

		case 1: //normaly send data
			// fmt.Println("-P-")
			// fmt.Println("P")
			// start := time.Now()
			md = childMap.GetRandomMessage()
			// fmt.Println("-P-took:before append,", time.Since(start))

			AggregatorRoundBuffer[0].Append(utils.ByteArrayConvertToUint(md.SID), md.Sign)

			signCounts++
			P := CreateAVFormatTransaction([]byte("P"), md, aggregatorProperty, nil)
			// fmt.Println("P", P)
			// fmt.Println("-P-took:before write,", time.Since(start))
			VerifierConn.Write(P)
			//check Feedback

			now := time.Since(AGGStart)
			ags.Add(&md.Sign)
			// fmt.Println("-P-took,", time.Since(start))
			if aggregatorProperty.Mode == 1 || aggregatorProperty.Mode == 2 {
				if signCounts-1 >= aggregatorProperty.CountThreshold {
					status = 2
					break
				}
			}
			if aggregatorProperty.Mode == 0 || aggregatorProperty.Mode == 2 {
				if now.Milliseconds() >= int64(TemporaryTimeWindow) {
					status = 2
					break
				}
			}
			if aggregatorProperty.Mode == 2 && (signCounts-1 >= aggregatorProperty.CountThreshold && time.Now().Before(AGGStart)) {
				status = 3
				break
			}

		case 2: // end normaly because time up
			// fmt.Println("-E-")

			// fmt.Println("E")
			md = childMap.GetRandomMessage()
			AggregatorRoundBuffer[0].Append(utils.ByteArrayConvertToUint(md.SID), md.Sign)
			AggregatorRoundBuffer[0].VUID = VerificationUnitCount
			println("VUID", AggregatorRoundBuffer[0].VUID)
			// println("----Before Refresh----")
			// for i := range AggregatorRoundBuffer {
			// 	AggregatorRoundBuffer[i].Print()
			// }
			if aggregatorProperty.TrackingMethod == 1 {
				AggregatorRoundBuffer[0] = SortAndAggregateIndependentSignWithID(AggregatorRoundBuffer[0])
			}
			RefreshRoundBuffer(AggregatorRoundBuffer)
			// println("----After Refresh----")
			// for i := range AggregatorRoundBuffer {
			// 	AggregatorRoundBuffer[i].Print()
			// }
			ags.Add(&md.Sign)
			E := CreateAVFormatTransaction([]byte("E"), md, aggregatorProperty, ags.Serialize())
			// fmt.Println("E", E)

			VerifierConn.Write(E)
			//recive Feedback here
			status = 0
			VerificationUnitCount++
			// if VerificationUnitCount%100 == 0 && signCounts > 0 {
			// 	TemporaryTimeWindow *= 2
			// 	println("---------------------------")
			// 	println("VU:", VerificationUnitCount)
			// 	println("TW update to ", TemporaryTimeWindow, "ms")
			// 	println("---------------------------")
			// }
			fmt.Println("E-Elapsed:", time.Since(VUstart))
			break
		case 3: //data overflow during time
			// fmt.Println("-O-")
			md = childMap.GetRandomMessage()
			AggregatorRoundBuffer[0].Append(utils.ByteArrayConvertToUint(md.SID), md.Sign)
			AggregatorRoundBuffer[0].VUID = VerificationUnitCount
			if aggregatorProperty.TrackingMethod == 1 {
				AggregatorRoundBuffer[0] = SortAndAggregateIndependentSignWithID(AggregatorRoundBuffer[0])
			}
			RefreshRoundBuffer(AggregatorRoundBuffer)

			ags.Add(&md.Sign)
			O := CreateAVFormatTransaction([]byte("O"), md, aggregatorProperty, ags.Serialize())
			// fmt.Println("O", O)
			VerifierConn.Write(O)
			status = 0
			VerificationUnitCount++
			// if VerificationUnitCount%100 == 0 && signCounts > 0 {
			// 	TemporaryTimeWindow *= 2
			// 	println("---------------------------")
			// 	println("VU:", VerificationUnitCount)
			// 	println("TW update to ", TemporaryTimeWindow, "ms")
			// 	println("---------------------------")
			// }
			fmt.Println("O-Elapsed:", time.Since(VUstart))
			break
		}
	}
}

//**********************************************************************
//****************VerifierHandler internal functions********************
//**********************************************************************

func ReciveVerifierECDSAPK(verifierConn *net.TCPConn) (fb *tracer.Feedback, VPK []byte) {
	var err error
	fb = new(tracer.Feedback)

	fb.AID, err = netcp.ReciveConstBytes(verifierConn, 4)
	if err != nil {
		println("In ReciveVerifierECDSAPK: Recive fb.AID")
		fmt.Println("Error:", err)
	}
	fb.Instruction, err = netcp.ReciveConstBytes(verifierConn, 40)
	if err != nil {
		println("In ReciveVerifierECDSAPK: Recive initial fb.instruction")
		fmt.Println("Error:", err)
	}
	fmt.Println("VPK from Verifier time record:", time.Unix(0, int64(utils.ByteArrayConvertToUint(fb.Instruction[32:]))))
	fb.VSIGN, err = netcp.ReciveConstBytes(verifierConn, 64)
	if err != nil {
		println("In ReciveVerifierECDSAPK: Recive initial fb.VECSDAPK")
		fmt.Println("Error:", err)
	}

	if !ed25519.Verify(fb.Instruction[0:32], hashToSHA512(append(fb.AID, fb.Instruction...)), fb.VSIGN) {
		fmt.Println("Verifier Self ECDSA Signature checked, false false false!!!!! A imposter!!!!!")
	} else {
		fmt.Println("Verifier Self ECDSA Signature checked, true. VPK preserved")
	}

	return fb, fb.Instruction[0:32]
}

func ReciveVerifierInstruction(verifierConn *net.TCPConn, VPK []byte) (fb *tracer.Feedback) {
	var err error
	fb = new(tracer.Feedback)

	fb.AID, err = netcp.ReciveConstBytes(verifierConn, 4)
	if err != nil {
		println("In ReciveVerifierECDSAPK: Recive fb.AID")
		fmt.Println("Error:", err)
	}
	fb.Instruction, err = netcp.ReciveConstBytes(verifierConn, 16)
	if err != nil {
		println("In ReciveVerifierECDSAPK: Recive initial fb.instruction")
		fmt.Println("Error:", err)
	}
	fmt.Println("VPK from Verifier time record:", time.Unix(0, int64(utils.ByteArrayConvertToUint(fb.Instruction[8:]))))
	fb.VSIGN, err = netcp.ReciveConstBytes(verifierConn, 64)
	if err != nil {
		println("In ReciveVerifierECDSAPK: Recive initial fb.instruction")
		fmt.Println("Error:", err)
	}

	if !ed25519.Verify(VPK, hashToSHA512(append(fb.AID, fb.Instruction...)), fb.VSIGN) {
		fmt.Println("Verifier Self ECDSA Signature checked, false false false!!!!! A imposter!!!!!")
	} else {
		fmt.Println("Verifier Self ECDSA Signature checked, true. Executing Instruction...")
	}

	return fb
}

// AutoFeedback wait time calibration
func AggregatorRegistration(VerifierConn *net.TCPConn, newAggregator *AggregatorProperty) (uint32, []byte) { //AID, Verifier EdDSA key
	//Send void MD

	var voidmd infra.MessageData
	voidmd.SID = utils.CreateConstLengthHeader(0, newAggregator.AVFormat.SIDSpace)
	for i := 0; i < 25; i++ {
		voidmd.Message = append(voidmd.Message, []byte{0, 0, 0, 0}...)
	}
	voidmd.MessageHeader = utils.CreateConstLengthHeader(len(voidmd.Message), newAggregator.AVFormat.MessageHeaderLength)
	for i := 0; i < 25; i++ {
		voidmd.SignerOption = append(voidmd.SignerOption, []byte{0, 0, 0, 0}...)
	}
	voidmd.SignerOptionHeader = utils.CreateConstLengthHeader(len(voidmd.SignerOption), newAggregator.AVFormat.SignerOptionHeaderLength)
	var voidSign []byte
	for i := 0; i < 12; i++ {
		voidSign = append(voidSign, []byte{0, 0, 0, 0}...)
	}

	R := CreateAVFormatTransaction([]byte("R"), &voidmd, newAggregator, nil)

	//Recive void Feedback with Assigned AID and public key
	_, err := VerifierConn.Write(R)
	if err != nil {
		fmt.Println("In AggregatorRegistration, ERROR:", err)
	}

	var fb, VPK = ReciveVerifierECDSAPK(VerifierConn)

	//check verifier self sign

	return uint32(utils.ByteArrayConvertToUint(fb.AID)), VPK
}
