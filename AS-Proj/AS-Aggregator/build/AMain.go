package main

import (
	"fmt"
	"os"

	"time"

	"aggregator.local/pkg/aggregator"
	"aggregator.local/pkg/infra"
	"github.com/xxixuuu/netcp"
)

func main() {

	//Prepare variables
	var aggregatorProperty = new(aggregator.AggregatorProperty)

	aggregator.NewAggregatorProperty(aggregatorProperty, os.Args)
	// bls.Init(bls.BLS12_381)
	fmt.Printf("["+time.Now().String()+"]\n"+"-AGGA: Using maxmum goroutine up to :%d\n", aggregatorProperty.MaxMultiHandler+3)

	chBufferHandlerTasks := make(chan uint32, 4096)

	//Tables
	childrenMap := infra.NewChildMap()
	handlerMap := infra.NewMultiHandlerMap()

	// //Dial to Parent, setup Parent Handler, Get AID Parent
	ParentTCPConn := netcp.CheckAndResolveDialAddress(aggregatorProperty.ParentAddress, aggregatorProperty.ParentPort)
	defer ParentTCPConn.Close()

	go aggregator.SyncAggregator(ParentTCPConn, aggregatorProperty, childrenMap)

	//go MultiChildHandler
	fmt.Println("MultiHandler x", aggregatorProperty.MaxMultiHandler)
	for i := aggregatorProperty.MaxMultiHandler - 1; i >= 0; i-- {
		go aggregator.ChildrenHandler(i, chBufferHandlerTasks, childrenMap, handlerMap, aggregatorProperty)
	}

	//Listen for Dials
	localTCPListenerConn := netcp.CheckAndListeningOnPort("0.0.0.0:", aggregatorProperty.AggregatorListenPort)
	defer localTCPListenerConn.Close()
	fmt.Println("-AGGA: AGGA tcp4 handler listning on 0.0.0.0:" + aggregatorProperty.AggregatorListenPort + "\n \n[I'm all ears]")

	aggregator.TCPRecvicer(ParentTCPConn, localTCPListenerConn, aggregatorProperty, chBufferHandlerTasks, childrenMap, handlerMap)

}
