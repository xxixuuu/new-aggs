package dummy

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
)

type DSFormtaJson struct {
	MessageHeaderLength int
}

type DummyProperty struct {
	Mode            int
	DataLength      int
	DataLengthRange int
	Interval        int
	MaxMessage      int
	SignerAddr      string
	SignerPort      string
	DSFormat        DSFormtaJson `json:"DSFormatJSON"`
}

func NewDummyProperty(newDummy *DummyProperty, osArgs []string) {
	//first configure with JSON File
	JSONFile, err := os.Open("./dconfig.json")
	if err != nil {
		fmt.Println("Error during open setting file dconfig.json:", err)
	}
	AllBytesFromJSONFile, err := ioutil.ReadAll(JSONFile)
	if err != nil {
		fmt.Println("Error read file dconfig.json:", err)
	}
	err = json.Unmarshal(AllBytesFromJSONFile, newDummy)
	if err != nil {
		fmt.Println("Error execute json unmarshal setting from file dconfig.json:", err)
	}

	//Then configure from os Arguments
	if os.Args[1] != "-" {
		newDummy.SignerAddr = osArgs[1]
	}
	if os.Args[2] != "-" {
		newDummy.SignerPort = osArgs[2]
	}
	if os.Args[3] != "-" {
		newDummy.Mode, err = strconv.Atoi(osArgs[3])
		if err != nil {
			fmt.Println("In NewDummyProperty, Error converting os.Args[3]", err)
		}
	}
	if os.Args[4] != "-" {
		newDummy.Interval, err = strconv.Atoi(osArgs[4])
		if err != nil {
			fmt.Println("In NewDummyProperty, Error converting os.Args[4]", err)
		}
	}
	if os.Args[5] != "-" {
		newDummy.DataLength, err = strconv.Atoi(osArgs[5])
		if err != nil {
			fmt.Println("In NewDummyProperty, Error converting os.Args[5]", err)
		}
	}

	if os.Args[6] != "-" {
		newDummy.MaxMessage, err = strconv.Atoi(osArgs[6])
		if err != nil {
			fmt.Println("In NewDummyProperty, Error converting os.Args[6]", err)
		}
	}

	fmt.Println("Dummy Configured:", newDummy)
}

//rune = int32
func RandStringbytes(index float64) ([]byte, uint32) {
	letterRunes := []rune("01234567789abcdefghijklmnopqrstuvwxyz!@$#*%)(_")

	b := make([]rune, int32(math.Pow(2, index)))
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	len := uint32(len(b))
	return []byte(string(b)), len
}

func RandRangeStringbytes(index float64, Range float64) ([]byte, uint32) {
	var inputRange float64 = float64(int(100*rand.Float64())%int(Range)) / 100.0
	if rand.Intn(2) == 0 {
		inputRange *= -1
	}
	var length = int(math.Pow(2, index) * (1.0 + inputRange))
	letterRunes := []rune("01234567789abcdefghijklmnopqrstuvwxyz!@$#*%)(_")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	len := uint32(len(b))
	return []byte(string(b)), len
}

func DummySendMessage(len uint32, m []byte, op *net.TCPConn, newDummy *DummyProperty) error { //could be async
	// fmt.Println(m)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, len)
	_, err := op.Write(append(buf[:newDummy.DSFormat.MessageHeaderLength], m[:]...))
	return err
}
