package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"dummy.local/pkg/dummy"
	"github.com/xxixuuu/netcp"
	"github.com/xxixuuu/utils"
)

func main() {
	var err error
	var Dummy *dummy.DummyProperty = new(dummy.DummyProperty)
	if len(os.Args) < 5 {
		fmt.Println("Not Enough os Args to configure Dummy Data Source...")
		fmt.Println("Dummy Entity Terminated.")
		return
	}
	rand.Seed(time.Now().UnixNano())
	dummy.NewDummyProperty(Dummy, os.Args)
	if Dummy.Mode > 2 {
		fmt.Println("-Dummy: Invalid Dummy Data Source Mode:", Dummy.Mode)
		return
	}

	var remoteTCPConn *net.TCPConn
	initDial := true

	for i := 0; ; i++ {
		if initDial || err != nil {
			remoteTCPConn = netcp.CheckAndResolveDialAddress(Dummy.SignerAddr, Dummy.SignerPort)
			netcp.ServerLog("-Dummy: Connected to Signer : " + Dummy.SignerAddr + ":" + Dummy.SignerPort)
			netcp.ServerLog("-Dummy: Providing approximately: 2^" + fmt.Sprint(Dummy.DataLength) + "with Range" + fmt.Sprint(Dummy.DataLengthRange) + "\nbyte stream interval=" + fmt.Sprint(Dummy.Interval) + "us")

			initDial = false
		} else {
			//Dial Remote
			var message []byte
			var msglen uint32

			switch m := Dummy.Mode; m {
			case 0:
				message, msglen = utils.FormattedTimeRecord()
			case 1:
				message, msglen = dummy.RandStringbytes(float64(Dummy.DataLength))
			case 2:
				message, msglen = dummy.RandRangeStringbytes(float64(Dummy.DataLength), float64(Dummy.DataLengthRange))
			default:
				println("Dummy Unknown mode, termiate")
				return
			}
			err := dummy.DummySendMessage(msglen, message, remoteTCPConn, Dummy)
			if err != nil {
				fmt.Println("Error: dummy.SendMessage() Failed")
				fmt.Println(err)
				println("Signer terminated!!!")
				initDial = true
			}
		}

		time.Sleep(time.Duration(Dummy.Interval) * time.Millisecond)
		if i > Dummy.MaxMessage {
			println("All messages sent: Dummy Data Sensor Terminate......")
			break
		}
	}
}
