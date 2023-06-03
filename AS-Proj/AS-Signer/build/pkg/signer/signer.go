package signer

// #cgo LDFLAGS: -lcrypto
// #include <stdio.h>
// #include <openssl/sha.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/herumi/bls-go-binary/bls"
	"github.com/xxixuuu/netcp"
	"github.com/xxixuuu/utils"
)

type SAFormtaJson struct {
	SIDSpace                 int
	MessageHeaderLength      int
	SignerOptionHeaderLength int
	SignerOption             int
	SignCurveParameter       int
}

type SignerProperty struct {
	Mode                int
	SID                 int
	Threads             int
	AggregatorAddr      string
	AggregatorPort      string
	SignerListenPort    string
	SignerBufferSize    int
	TimeTillFall        int //MilliSec
	AggregatorKeepAlive int //MilliSec default = 0
	SIDGetAssigned      int
	SAFormat            SAFormtaJson `json:"SAFormatJSON"`
}

func NewSignerProperty(newSigner *SignerProperty, osArgs []string) {
	JSONFile, err := os.Open("./sconfig.json")
	if err != nil {
		fmt.Println("Error during open setting file sconfig.json:", err)
	}

	AllBytesFromJSONFile, err := ioutil.ReadAll(JSONFile)
	if err != nil {
		fmt.Println("Error read file sconfig.json:", err)
	}

	err = json.Unmarshal(AllBytesFromJSONFile, newSigner)
	if err != nil {
		fmt.Println("Error execute json unmarshal setting from file sconfig.json:", err)
	}

	//Then configure from os Arguments
	if os.Args[1] != "-" {
		newSigner.AggregatorAddr = osArgs[1]
	}
	if os.Args[2] != "-" {
		newSigner.AggregatorPort = osArgs[2]
	}
	if os.Args[3] != "-" {
		newSigner.SignerListenPort = osArgs[3]
	}
	if os.Args[4] != "-" {
		newSigner.Mode, err = strconv.Atoi(osArgs[4])
		if err != nil {
			fmt.Println("In NewSignerProperty, Error converting os.Args[4]", err)
		}
	}
	if os.Args[5] != "-" {
		newSigner.TimeTillFall, err = strconv.Atoi(osArgs[5])
		if err != nil {
			fmt.Println("In NewSignerProperty, Error converting os.Args[5]", err)
		}
	}
	if os.Args[6] != "-" {
		newSigner.SID, err = strconv.Atoi(osArgs[6])
		if err != nil {
			fmt.Println("In NewSignerProperty, Error converting os.Args[5]", err)
		}
	}

	fmt.Println("newSigner Configured:", newSigner)
}

func SendFormattedMessageSign(sign []byte, messageWithSignerOption []byte, op *net.TCPConn) error {
	//FORMAT [SIGN:48]-[HEADER:4]-[message<-HEADER] HEADER has to be littleEndian

	data := append(messageWithSignerOption[:], sign[:]...)
	// fmt.Println(data)
	readDataLegnth, err := op.Write(data)
	if err != nil {
		fmt.Println("Error: In SendFormattedMessageSign:")
		fmt.Println("Error: In netConn.Write: data")
		fmt.Println(err)
		fmt.Println("Read:", readDataLegnth)
		return err
	}
	return nil
}

func SendFormattedFalseMessageSign(sign []byte, messageWithSignerOption []byte, op *net.TCPConn) error {
	//FORMAT [SIGN:48]-[HEADER:4]-[false message<-HEADER] HEADER has to be littleEndian
	//change message is self here
	messageWithSignerOption[len(messageWithSignerOption)-2] += 3

	data := append(messageWithSignerOption[:], sign[:]...)
	readDataLegnth, err := op.Write(data)
	if err != nil {
		fmt.Println("Error: In SendFormattedFalseMessageSign:")
		fmt.Println("Error: In netConn.Write: data")
		fmt.Println(err)
		fmt.Println("Read:", readDataLegnth)
		return err
	}
	return nil
}

func DataSourceHandler(op *net.TCPConn, ch_databuffer chan []byte, newSigner *SignerProperty) string {
	clientinfo := strings.Split(op.RemoteAddr().String(), ":")
	netcp.ServerLog("Message came from IP: " + clientinfo[0])
	for {

		_, data, err := netcp.ReciveConstHeaderData(op, newSigner.SAFormat.MessageHeaderLength)
		if err != nil {
			fmt.Println("Error: In DataSourceHandler:")
			fmt.Println("Error: In ReciveFormattedDataFromDataSource: data")
			fmt.Println(err)
			return ""
		}
		ch_databuffer <- data
	}
}

func GenerateSignerOptions(message []byte, newSigner *SignerProperty) []byte {
	var SignerOption []byte
	switch so := newSigner.SAFormat.SignerOption; so {
	case 1: //Unix Epoc 4 byte
		SignerOption = utils.UnixTimeRecord()[:]
	case 2: //Unix Epoc nano 8 byte
		SignerOption = utils.UnixTimeRecordNano()[:]
	case 3: //16 byte rand
		SignerOption = utils.GenerateRandByteArray(16)[:]
	case 4: //32 byte rand
		SignerOption = utils.GenerateRandByteArray(32)[:]
		//Implement here
	}

	message = append(message, utils.CreateConstLengthHeader(len(SignerOption), newSigner.SAFormat.SignerOptionHeaderLength)...)
	return append(message, SignerOption...)
}

func SignerRegistration(op *net.TCPConn, SK *bls.SecretKey, SerializedPK []byte, newSigner *SignerProperty) error {

	fmt.Println("Registing self...SID", newSigner.SID)

	R := make([]byte, 4)
	binary.LittleEndian.PutUint32(R, uint32(len(SerializedPK)))
	R = R[:newSigner.SAFormat.MessageHeaderLength]                                              //[header]
	R = append(R, SerializedPK...)                                                              //[header][message(pk)]
	R = GenerateSignerOptions(R, newSigner)                                                     //[MessageHeader][message(pk)][SignerOptionHeader][SignerOption]
	R = append(utils.CreateConstLengthHeader(newSigner.SID, newSigner.SAFormat.SIDSpace), R...) //[SID][MessageHeader][message(pk)][SignerOptionHeader][SignerOption]

	RSign := SK.SignByte(R)
	RSign.Serialize()
	R = append(R, RSign.Serialize()...)
	// fmt.Println("R", R
	sentLength, err := op.Write(R)
	fmt.Println("-Signer: Signer send signer info to Aggregator ", sentLength, " byte")
	if err != nil {
		fmt.Println("In SignerRegistration: ", err)
	}
	return err
}

func HashAndSignHashedMessage(message []byte, sk bls.SecretKey) []byte {
	hash := make([]byte, C.SHA512_DIGEST_LENGTH)
	ctx := &C.struct_SHA512state_st{}
	C.SHA512_Init(ctx)
	_ = C.SHA512_Update(ctx, C.CBytes(message), C.size_t(len(message)))
	_ = C.SHA512_Final((*C.uchar)(&hash[0]), ctx)
	sign := *sk.SignHash(hash[:])
	return sign.Serialize()
}

func SignTransaction(message []byte, sk bls.SecretKey, newSigner *SignerProperty) (TransactionBody []byte, sign []byte) {
	//Append Message Header [MessageHeader]:[Message]

	//Append SignerOption Header and Option [MessageHeader]:[Message]:[SignerOptionHeader]:[SignerOption]
	TransactionBody = GenerateSignerOptions(message, newSigner)

	//Append Sid   [SID]:[MessageHeader]:[Message]:[SignerOptionHeader]:[SignerOption]
	if newSigner.SIDGetAssigned != 1 {
		TransactionBody = append(utils.CreateConstLengthHeader(newSigner.SID, newSigner.SAFormat.SIDSpace), TransactionBody...)
	}

	// fmt.Println("-TB:", TransactionBody)
	// serializedSign =  Sign([SID]:[MessageHeader]:[Message]:[SignerOptionHeader]:[SignerOption])
	sign = HashAndSignHashedMessage(TransactionBody, sk)
	if newSigner.SIDGetAssigned == 1 {
		TransactionBody = append(utils.CreateConstLengthHeader(newSigner.SID, newSigner.SAFormat.SIDSpace), TransactionBody...)
	}
	return TransactionBody, sign
}

func SingleThreadSigner(remoteConn *net.TCPConn, dataSourceConn *net.TCPConn, chTermination chan int, falltime time.Time, newSigner *SignerProperty) string {
	// Signer GenKeys
	var sk bls.SecretKey
	var pk bls.PublicKey
	sk.SetByCSPRNG()
	pk = *sk.GetPublicKey()
	netcp.ServerLog("-Signer: Genkey set done.")
	var err error
	for {
		//Send Signer Info to Aggregator with SA Format
		err = SignerRegistration(remoteConn, &sk, pk.Serialize(), newSigner)
		if err != nil {
			fmt.Println("In SingleThreadSigner, Error during Signer Registration:", err)
		} else {
			break
		}
	}

	//keyPressListnerLoop:
	f := false
	mode := newSigner.Mode

	//Regular loop
	for {
		//Dataformat [MessageHeader][Message]
		dataheader, data, err := netcp.ReciveConstHeaderData(dataSourceConn, newSigner.SAFormat.MessageHeaderLength)

		if err != nil {
			fmt.Println("Error: In DataSourceHandler:")
			fmt.Println("Error: In ReciveFormattedDataFromDataSource: data")
			fmt.Println(err)
			fmt.Println("-=Signer Terminated=-")
			return ""
		}
		signScope, sign := SignTransaction(append(dataheader, data...), sk, newSigner)

		var ttfNotPassedYet = false
		if mode == 0 {
			ttfNotPassedYet = true
		} else if mode == 1 {
			ttfNotPassedYet = false
		} else if mode == 3 {
			ttfNotPassedYet = time.Now().Before(falltime)
		}

		// println("ttfNotPassedYet", ttfNotPassedYet)
		if ttfNotPassedYet {
			err = SendFormattedMessageSign(sign, signScope, remoteConn)
		} else {
			if !f {
				err = SendFormattedFalseMessageSign(sign, signScope, remoteConn)
				// if !f {
				fmt.Println("----------------------------------------------")
				fmt.Println("-------        Intruder mode          --------")
				fmt.Println("----------------------------------------------")
				if mode == 3 {
					f = true
				}
				// }
			} else {
				err = SendFormattedMessageSign(sign, signScope, remoteConn)
			}
		}

		if err != nil {
			chTermination <- 1
			fmt.Println("Error: send error in Signer")
			fmt.Println("Error: send error in SendFormattedMessageSign")
			fmt.Println(err)
			return ""
		}

	}
}
