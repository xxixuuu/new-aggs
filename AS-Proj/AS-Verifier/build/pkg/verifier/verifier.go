package verifier

// #cgo LDFLAGS: -lcrypto
// #include <stdio.h>
// #include <openssl/sha.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"context"
	ed25519 "crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/xxixuuu/netcp"
	"github.com/xxixuuu/utils"
	"verifier.local/pkg/infra"
	"verifier.local/pkg/kvs"
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

type VAFormatJson struct {
	AIDSpace                 int
	InstructionSpace         int    //byte
	InstructionContentsSpace int    //byte
	ECDSAParams              string //Decides Signature length(Default EdDSA)
}

type VerifierProperty struct {
	Mode                     int //5 mode=0 = all Aggregator handler dose pairing,mode =1 single pairing thread with multiAggregator Handler
	RedisAddress             string
	VerifierListenPort       string //3 AggregatorListenPort
	VerifierBufferSize       int    //4 buffer size per channel AggregatorBufferSize
	ChildKeepAlive           int    //6 if !=0 :true KeepAlive
	MaxMultiHandler          int    //10 Maximum Multi-Handler goroutine //0 = unlimit
	AIDAssign                int
	SIDAssign                int
	SignCurveParameter       int
	TrackingMethod           int
	HashedMessageRoundBuffer int
	AVFormat                 AVFormatJson `json:"AVFormatJSON"`
	VAFormat                 VAFormatJson `json:"VAFormatJSON"`
}

func NewVerifierProperty(newVerifier *VerifierProperty, osArgs []string) {
	var err error

	JSONFile, err := os.Open("./vconfig.json")
	if err != nil {
		fmt.Println("Error during open setting file sconfig.json:", err)
	}

	AllBytesFromJSONFile, err := ioutil.ReadAll(JSONFile)
	if err != nil {
		fmt.Println("Error read file aconfig.json:", err)
	}

	err = json.Unmarshal(AllBytesFromJSONFile, newVerifier)
	if err != nil {
		fmt.Println("Error execute json unmarshal setting from file aconfig.json:", err)
	}

	//Then configure from os Arguments
	if os.Args[1] != "-" {
		newVerifier.RedisAddress = osArgs[1]
	}
	if os.Args[2] != "-" {
		newVerifier.VerifierListenPort = osArgs[2]
	}
	if os.Args[3] != "-" {
		newVerifier.TrackingMethod, err = strconv.Atoi(osArgs[3])
		if err != nil {
			fmt.Println("In NewVerifierProperty, Error converting os.Args[4]", err)
		}
	}

	fmt.Println("newVerifier Configed:", newVerifier)
}

//--------------Crypto--------------

func hashToSHA512(message []byte) []byte {
	hash := make([]byte, C.SHA512_DIGEST_LENGTH)
	ctx := &C.struct_SHA512state_st{}
	C.SHA512_Init(ctx)
	_ = C.SHA512_Update(ctx, C.CBytes(message), C.size_t(len(message)))
	_ = C.SHA512_Final((*C.uchar)(&hash[0]), ctx)
	return hash
}

type ECDSAKeyset struct {
	pk ed25519.PublicKey //32byte
	sk ed25519.PrivateKey
}

func NewEdDSAKeyset() *ECDSAKeyset {
	key := new(ECDSAKeyset)
	var err error
	key.pk, key.sk, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		netcp.ServerLog("Failed to generate new EdSAKeyset Error")
		println(err)
	}
	return key
}

//--------------Crypto--------------

//--------------Network functions--------------

func Send(m string, op *net.TCPConn) { //could be async
	op.Write([]byte(string(m + "\n")))
}

func SendBytes(m []byte, op *net.TCPConn) { //for ECDSA

	mlen := len(m)
	//fmt.Printf("Send Data %v", m)
	//fmt.Println("Send bytes mlen=" + strconv.Itoa(mlen))
	if mlen < 256 {
		op.Write(m)
	} else {
		println("Function SendBytes(): mlen >= 256")
	}
}

func CheckSignersIfDistinct(KSidArray []string) bool {
	tagSid := ""
	nowSid := ""
	for i := range KSidArray {
		tagSid = KSidArray[i]
		for j := i + 1; j < len(KSidArray); j++ {
			nowSid = KSidArray[j]
			if tagSid == nowSid {
				return false
			}
		}
	}
	return true
}

//--------------Network functions--------------

func ReciveAVFormatTransaction(op *net.TCPConn, mid int, verifierProperty *VerifierProperty) (*kvs.MessageData, error) {
	// println("\nReciving......\n")
	var err error = nil
	var md *kvs.MessageData = new(kvs.MessageData)
	md.MID = uint64(mid)
	var StateHeaderLength int = 1
	md.STATEHEADER, err = netcp.ReciveConstBytes(op, StateHeaderLength)
	if err != nil {
		fmt.Println("In func ReciveSiAData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.AID")
		fmt.Println(err)
		return nil, err
	}

	if md.STATEHEADER[0] == []byte("A")[0] {
		fmt.Println("In func ReciveSiAData Header == A:")
		fmt.Println("HandOver to Function ReciveA")
		return md, nil
	}

	// fmt.Println("StateHeader", md.STATEHEADER)
	md.AID, err = netcp.ReciveConstBytes(op, verifierProperty.AVFormat.AIDSpace)
	if err != nil {
		fmt.Println("In func ReciveSiAData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.AID")
		fmt.Println(err)
		return nil, err
	}
	// fmt.Println("AID", md.AID)

	md.SID, err = netcp.ReciveConstBytes(op, verifierProperty.AVFormat.SIDSpace)
	if err != nil {
		fmt.Println("In func ReciveSiAData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.SID")
		fmt.Println(err)
		return nil, err
	}

	// fmt.Println("SID", md.SID)
	md.MESSAGEHEAHDER, md.MESSAGE, err = netcp.ReciveConstHeaderData(op, verifierProperty.AVFormat.MessageHeaderLength)
	if err != nil {
		fmt.Println("In func ReciveSiAData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.Message")
		fmt.Println(err)
		return nil, err
	}
	// fmt.Println("MH", md.MESSAGEHEAHDER)
	// fmt.Println("M", md.MESSAGE)
	md.SIGNEROPTIONHEADER, md.SIGNEROPTION, err = netcp.ReciveConstHeaderData(op, verifierProperty.AVFormat.SignerOptionHeaderLength)
	if err != nil {
		fmt.Println("In func ReciveSiAData Error:", err)
		fmt.Println("Error: In ReciveConstHeaderData: md.SO")
		fmt.Println(err)
		return nil, err
	}
	// fmt.Println("SOH", md.SIGNEROPTIONHEADER)
	// fmt.Println("SO", md.SIGNEROPTION)

	if md.STATEHEADER[0] == []byte("E")[0] || md.STATEHEADER[0] == []byte("O")[0] || (md.STATEHEADER[0] == []byte("R")[0] && md.MESSAGEHEAHDER[0] != 100) {

		var signlength int = 48
		md.SIGN, err = netcp.ReciveConstBytes(op, signlength)
		if err != nil {
			fmt.Println("In func ReciveSiAData Error:", err)
			fmt.Println("Error: In ReciveConstHeaderData: md.Sign")
			fmt.Println(err)
			return nil, err
		}
	}
	md.REGDATE = utils.UintConvertToLittleEndianByteArray(uint64(time.Now().UnixNano()))
	// fmt.Println("REGDATE", md.REGDATE)
	return md, nil
}

func SignECDSA(message []byte, VSK []byte) (sign []byte) {
	hashedMessage := hashToSHA512(message)
	return ed25519.Sign(VSK, hashedMessage)

}

func SendVerifierECDSAPK(aggregatorConn *net.TCPConn, AID []byte, keyset *ECDSAKeyset) {
	// fmt.Println("ECDSA len(pk)", len(keyset.pk))
	message := make([]byte, 0)
	message = append(message, AID...)
	message = append(message, keyset.pk...)
	message = append(message, utils.UnixTimeRecordNano()...)

	sign := SignECDSA(message, keyset.sk)

	message = append(message, sign...)
	_, err := aggregatorConn.Write(message)
	if err != nil {
		fmt.Println("-In SendVerifierECDSAPK, write tcpconn error")
		fmt.Println("Error,", err)
	}
}

func SendVerifierInstruction(aggregatorConn *net.TCPConn, AID []byte, SID []byte, VUID []byte, keyset *ECDSAKeyset, ctx context.Context, redisConn *redis.Client) {
	fb := new(kvs.Feedback)
	fb.AID = AID
	fb.SID = SID
	fb.VUID = VUID

	message := make([]byte, 0)
	message = append(message, AID...)
	message = append(message, SID...)
	message = append(message, VUID...)
	ts := utils.UnixTimeRecordNano()
	message = append(message, ts...)
	fb.REGDATE = ts
	sign := SignECDSA(message, keyset.sk)
	fb.VSIGN = sign
	message = append(message, sign...)
	_, err := aggregatorConn.Write(message)
	if err != nil {
		fmt.Println("-In SendVerifierInstruction, write tcpconn error")
		fmt.Println("Error,", err)
	}
	go fb.Set(ctx, redisConn)
}

func AggregatorRegistration(AggregatorConn *net.TCPConn, AID uint32, PK []byte, keyset ECDSAKeyset, ctx context.Context, RedisClient *redis.Client, verifierProperty *VerifierProperty) (*kvs.MessageData, *kvs.Aggregator) {
	clientinfo := strings.Split(AggregatorConn.RemoteAddr().String(), ":")

	md, err := ReciveAVFormatTransaction(AggregatorConn, 0, verifierProperty)
	if err != nil {
		fmt.Println("In AggregatorRegistration, ReciveAVFormatTransaction.\n ERROR:")
		fmt.Println(err)
		return nil, nil
	}
	fmt.Println("New Aggreagtor, AID:", md.AID)
	newAggregator := new(kvs.Aggregator)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, AID)
	newAggregator.SetFromMessageData(md, buf[:], verifierProperty.AIDAssign, clientinfo[0], clientinfo[1])
	go newAggregator.Set(ctx, RedisClient)
	//send signer AID
	SendVerifierECDSAPK(AggregatorConn, utils.UintConvertToLittleEndianByteArray(AID), &keyset)
	return md, newAggregator

}

// --------------Entityts-----------------------
type TrackingSource struct {
	SerializedSignArray [][]byte
	HashedMDArray       *kvs.HashedMessageArraywithID //
}

func VerifyIndependentSignArray(AID []byte, ts *TrackingSource, verifierProperty *VerifierProperty, pkmap *infra.PKMap, ctx context.Context, RedisClient *redis.Client) {
	var sign bls.Sign
	println("Start Independent Verify")
	fb := new(kvs.Feedback)
	fb.AID = AID
	if verifierProperty.TrackingMethod == 0 {
		println("Trackinge Method:", 0)
		fmt.Println("Independent Signature Counts:", len(ts.SerializedSignArray))
		fmt.Println("Independent HashedMessageArray:", len(ts.HashedMDArray.HashedMessageArray))

		for i := range ts.SerializedSignArray {
			// fmt.Println("hashed message", fmt.Sprint(ts.HashedMDArray.HashedMessageArray[i].HashedMessage))
			err := sign.Deserialize(ts.SerializedSignArray[i])
			if err != nil {
				fmt.Println("-In VerifyIndependentSignArray")
				fmt.Println("Error:", err)
			}
			if !sign.VerifyHash(pkmap.Get(fmt.Sprint(ts.HashedMDArray.HashedMessageArray[i].SID)), ts.HashedMDArray.HashedMessageArray[i].HashedMessage) {
				println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				fmt.Println("|      Defective Sender Tracked!!! SID:", ts.HashedMDArray.HashedMessageArray[i].SID, "     |")
				trackedDate := time.Now()
				fmt.Println("|      Date:", trackedDate, "|")
				println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				fb.VUID = utils.UintConvertToLittleEndianByteArray(uint64(ts.HashedMDArray.VUID))
				fb.TRACKEDDATE = utils.UintConvertToLittleEndianByteArray(uint64(trackedDate.UnixNano()))
				// break
			}
		}
	} else if verifierProperty.TrackingMethod == 1 {
		//Sort MDArray First\
		println("Trackinge Method:", 1)
		fmt.Print("Independent Signature by Signer Counts:", len(ts.SerializedSignArray))
		println(", len(ts.HashedMDArray.HashedMessageArray):", len(ts.HashedMDArray.HashedMessageArray))

		sort.Sort(ts.HashedMDArray)
		var sign bls.Sign
		for i := range ts.SerializedSignArray {
			err := sign.Deserialize(ts.SerializedSignArray[i])
			if err != nil {
				fmt.Println("-In VerifyIndependentSignArray")
				fmt.Println("Error:", err)
			}
			TargetSID := utils.ByteArrayConvertToUint(ts.HashedMDArray.HashedMessageArray[0].SID)
			SPK := pkmap.Get(fmt.Sprint(ts.HashedMDArray.HashedMessageArray[0].SID))
			signerHashedMessages := ts.HashedMDArray.TrimSingleSignerMessageArrayScope()
			// println("\nlen(signerHashedMessages):", len(signerHashedMessages))

			SPKArray := make([]bls.PublicKey, len(signerHashedMessages))
			for i := range signerHashedMessages {
				SPKArray[i] = *SPK
			}
			if !sign.VerifyAggregateHashes(SPKArray, signerHashedMessages) {
				println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				fmt.Println("|      Defective Sender Tracked!!! SID:", TargetSID, "              |")
				trackedDate := time.Now()
				fmt.Println("|      Date:", trackedDate, "|")
				println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				fb.VUID = utils.UintConvertToLittleEndianByteArray(uint64(ts.HashedMDArray.VUID))
				fb.TRACKEDDATE = utils.UintConvertToLittleEndianByteArray(uint64(trackedDate.UnixNano()))
				// go fb.SetTrackedData(ctx, RedisClient)

			}
		}
	} else {
		println("UnKnown Tracking Method, Tracking terminate")
	}
	go fb.SetTrackedData(ctx, RedisClient)

}

func Verifier(AggregatorConn *net.TCPConn, AID []byte, verifierProperty *VerifierProperty, SPKMap *infra.PKMap, ch_TrackingSource <-chan *TrackingSource, ctx context.Context, RedisClient *redis.Client) {

	for {
		if len(ch_TrackingSource) > 0 {
			ts := <-ch_TrackingSource
			VerifyIndependentSignArray(AID, ts, verifierProperty, SPKMap, ctx, RedisClient) //I bet this is not gonna work.....
		}
	}
}

func ReciveIndependentSignArray(op *net.TCPConn) (signArray [][]byte) { //When its A

	AID, err := netcp.ReciveConstBytes(op, 4)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	} else {
		fmt.Println("Recived Independent Tracking Resource form AID:", utils.ByteArrayConvertToUint(AID))
	}
	SignNumber, err := netcp.ReciveConstBytes(op, 4)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	SignCounts := utils.ByteArrayConvertToUint(SignNumber)
	// println("SignCounts:", SignCounts)
	signArray = make([][]byte, 0)
	for i := 0; i < SignCounts; i++ {
		serializedSign, err := netcp.ReciveConstBytes(op, 48)
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		signArray = append(signArray, serializedSign)
	}
	return signArray
}

func AggregatorHandler(AggregatorConn *net.TCPConn, AID uint32, verifierProperty *VerifierProperty, keyset *ECDSAKeyset, ctx context.Context, RedisClient *redis.Client) string {
	//Recive
	bls.Init(verifierProperty.SignCurveParameter)
	clientinfo := strings.Split(AggregatorConn.RemoteAddr().String(), ":")
	netcp.ServerLog("Serving Entity IP: " + clientinfo[0] + clientinfo[1])

	AIDbuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(AIDbuf, AID)
	println(AID)

	netcp.ServerLog("Establishing connection with Aggregator ..." + fmt.Sprint(AID))
	vpk := make([]byte, 32) //EdDSA PK
	copy(vpk[:], keyset.pk)
	_, aggregator := AggregatorRegistration(AggregatorConn, uint32(AID), vpk, *keyset, ctx, RedisClient, verifierProperty)
	fmt.Println("New Aggregator Connected", aggregator)

	//signerPKMap
	SPKMAP := infra.NewPKMap()

	//Initialize bls-go-binary
	bls.Init(bls.BLS12_381)

	//spwan verifier
	var ch_TrackingSource = make(chan *TrackingSource, 1)
	go Verifier(AggregatorConn, AIDbuf, verifierProperty, SPKMAP, ch_TrackingSource, ctx, RedisClient)
	SID := 0
	MID := 0
	VUID := 1
	var vd *kvs.VerifyData
	var SignerKeyArray []string
	var start time.Time
	// var lastFalseVID = 0
	var CHTrackingVUID = make(chan int, verifierProperty.HashedMessageRoundBuffer)
	RoundHashBufferArray := make([]*kvs.HashedMessageArraywithID, verifierProperty.HashedMessageRoundBuffer)

	kvs.InitRoundBufferArray(&RoundHashBufferArray)

	for {
		vd = new(kvs.VerifyData)
		vd.AID = utils.UintConvertToLittleEndianByteArray(AID)
		SignerKeyArray = nil
		for {
			MID++
			md, err := ReciveAVFormatTransaction(AggregatorConn, MID, verifierProperty)
			if err != nil {
				fmt.Println("In Aggregator Handler during ReciveAVFormatTransaction.\n ERROR:", err)
				fmt.Println("Aggregator", AID, " Handler Terminated...")
				return ""
			}

			if md.STATEHEADER[0] == byte('R') {
				println("New Signer Recived")
				//New Signer
				SID++
				buf := make([]byte, 8)
				binary.LittleEndian.PutUint64(buf, uint64(SID))
				newSigner := new(kvs.Signer)
				newSigner.SetFromMessageData(md, buf, AIDbuf, verifierProperty.SIDAssign)

				//Regist to Redis
				go newSigner.Set(ctx, RedisClient)
				spk := new(bls.PublicKey)
				err = spk.Deserialize(newSigner.SPK)
				if err != nil {
					println("In Recive R: Signer Deserialize failed")
					fmt.Println(err)
				}
				//Add to Map
				SPKMAP.Set(fmt.Sprint(newSigner.SID), spk)
				//Aggregator Append
				aggregator.Append(ctx, RedisClient, newSigner)

			} else if md.STATEHEADER[0] == byte('A') {
				println("A")
				TrackingSource := new(TrackingSource)
				TrackingSource.SerializedSignArray = ReciveIndependentSignArray(AggregatorConn)
				println("Recived Sign number:", len(TrackingSource.SerializedSignArray))
				//blocking
				lastFalseVID := <-CHTrackingVUID
				for i := range RoundHashBufferArray {
					if lastFalseVID == RoundHashBufferArray[i].VUID {
						TrackingSource.HashedMDArray = new(kvs.HashedMessageArraywithID)
						*&TrackingSource.HashedMDArray.VUID = *&RoundHashBufferArray[i].VUID
						*TrackingSource.HashedMDArray = *RoundHashBufferArray[i]
					}
				}
				println("-A:VUID:", TrackingSource.HashedMDArray.VUID)
				println("HashedVec len:", len(TrackingSource.HashedMDArray.HashedMessageArray))
				ch_TrackingSource <- TrackingSource
			} else {
				if md.STATEHEADER[0] == byte('S') {
					start = time.Now()
					RoundHashBufferArray[0] = new(kvs.HashedMessageArraywithID)
					RoundHashBufferArray[0].VUID = VUID
				}
				go md.Set(ctx, RedisClient)
				// fmt.Println("-TB", md.TrimSignatureScope())
				hashedMessage := hashToSHA512(md.TrimSignatureScope(verifierProperty.SIDAssign))
				vd.HASHVEC = append(vd.HASHVEC, hashedMessage) //depends on mode
				RoundHashBufferArray[0].AppendNew(md.SID, hashedMessage)

				SignerKeyArray = append(SignerKeyArray, fmt.Sprint(md.SID))
				//append Public ket
			}

			if md.STATEHEADER[0] == byte('E') || md.STATEHEADER[0] == byte('O') {
				// fmt.Println(string(md.STATEHEADER[0]))
				//verify
				getSPKVecStart := time.Now()
				SPKArray := SPKMAP.GetSPKVec(SignerKeyArray)
				getSPKVecElapsed := time.Since(getSPKVecStart)
				fmt.Println("getSPKVecElapsed:", getSPKVecElapsed)
				// println("-A:VUID:", RoundHashBufferArray[0].VUID)
				// println("HashedVec len:", len(RoundHashBufferArray[0].HashedMessageArray))
				kvs.UpdateRoundBufferArray(&RoundHashBufferArray)

				val := func(vd kvs.VerifyData) {
					vd.AGS = md.SIGN //depends on mode

					vd.VDID = uint64(VUID)
					var ags bls.Sign
					err := ags.Deserialize(vd.AGS)
					if err != nil {
						fmt.Println("In verifier Failed to Deserilize AGS, ABORT!")
						fmt.Println(err)
					}

					vd.AGGDATE = utils.UnixTimeRecordNano()
					HASHVEC := vd.HASHVEC
					// fmt.Println("SPKArray Length:", len(SPKArray), ", vd.HASHVEC Length:", len(HASHVEC))
					if ags.VerifyAggregateHashes(SPKArray, HASHVEC) == true {
						println("-TRUE- msg num:", len(HASHVEC))
						vd.RES = true
						//vd.PrintVerifyData()
					} else {
						println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						println("-------------------------------------------------------------")
						println("FALSE ! FALSE ! FALSE ! AGS verified false!!!! An Itruder !!!")
						println("-------------------------------------------------------------")
						println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						// lastFalseVID = VUID
						println("-TRUE- msg num:", len(HASHVEC))
						CHTrackingVUID <- VUID - 1
						vd.RES = false
						println("target VUID", VUID)
						//Send Verifier Feedback
						SendVerifierInstruction(AggregatorConn, aggregator.AID, utils.UintConvertToLittleEndianByteArray(uint32(0)), utils.UintConvertToLittleEndianByteArray(uint32(VUID)), keyset, ctx, RedisClient)
					}
					vd.REGDATE = utils.UnixTimeRecordNano()
					if !vd.RES {
						println(time.Now().String())
					}
					fmt.Print("VUID:", VUID, "  Elapsed:", time.Since(start), "+msg num:", len(HASHVEC), "\n\n")

					go vd.Set(ctx, RedisClient)
				}

				go val(*vd)

				vd = nil
				VUID++
				break

			}

		}
	}

}
