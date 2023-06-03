package main

import (
	"fmt"
	"os"
	"time"

	"github.com/herumi/bls-go-binary/bls"
	"github.com/xxixuuu/netcp"
	"github.com/xxixuuu/utils"
	"signer.local/pkg/signer"
)

func main() {
	//Signer Params set
	if len(os.Args) < 6 {
		fmt.Println("Not Enough os Args to configure Signer Data Source...")
		fmt.Println("Dummy Entity Terminated.")
		return
	}

	var signerProperty *signer.SignerProperty = new(signer.SignerProperty)
	signer.NewSignerProperty(signerProperty, os.Args)
	fmt.Println("-Signer: Signer Mode set to ", signerProperty.Mode)
	bls.Init(signerProperty.SAFormat.SignCurveParameter) //init for bls381, maybe not MT safe

	//channel set
	chTermination := make(chan int, 2)
	chMode := make(chan int, 2)
	messageBufferChannel := make(chan []byte, signerProperty.SignerBufferSize)

	//Dial to Aggregator and run Signer()
	remoteTCPConn := netcp.CheckAndResolveDialAddress(signerProperty.AggregatorAddr, signerProperty.AggregatorPort)
	defer remoteTCPConn.Close()
	fmt.Println("-Signer: Connected to aggregator : " + signerProperty.AggregatorAddr + ":" + " signerProperty.AggregatorPort")
	falltime := time.Now().Add(time.Duration(signerProperty.TimeTillFall) * time.Millisecond)

	//Listen on port
	localTCPListenerConn := netcp.CheckAndListeningOnPort("0.0.0.0", ":"+signerProperty.SignerListenPort)
	defer localTCPListenerConn.Close()
	fmt.Printf("["+time.Now().String()+"]\n"+"-Signer: Using maxmum cpu core up to :%d\n", signerProperty.Threads)
	fmt.Println("-Signer: Signer tcp4 handler listning on " + signerProperty.SignerListenPort + "\n[I'm all ears]")

	tcpConn, err := localTCPListenerConn.AcceptTCP()
	if err != nil {
		fmt.Println("Error: In main:")
		fmt.Println("Error: In AcceptTCP:")
		fmt.Println(err)
		return
	}
	if signerProperty.AggregatorKeepAlive > 0 {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Second * time.Duration(signerProperty.AggregatorKeepAlive))
	}

	if signerProperty.Threads < 0 {
		println("Handler Released")
		go signer.DataSourceHandler(tcpConn, messageBufferChannel, signerProperty)
	} else {
		signer.SingleThreadSigner(remoteTCPConn, tcpConn, chTermination, falltime, signerProperty)
	}

	//Manual interaction
	if signerProperty.Mode == 2 {
		utils.GetInput(chMode, chTermination) //for debug
	}
	<-chTermination
}
