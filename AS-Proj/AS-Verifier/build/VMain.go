package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xxixuuu/netcp"
	"verifier.local/pkg/kvs"
	"verifier.local/pkg/verifier"
)

func main() {
	//Verifier Configure
	var newVerifier verifier.VerifierProperty
	verifier.NewVerifierProperty(&newVerifier, os.Args)

	//Connect to Redis
	ctx := context.Background()
	RedisClient := kvs.NewClient(ctx, newVerifier.RedisAddress)

	//Prepare ECDSA Keyset
	EdDSAkeyset := verifier.NewEdDSAKeyset()

	//TCP Resolv and Listen
	VerifierListner := netcp.CheckAndListeningOnPort("0.0.0.0", ":"+newVerifier.VerifierListenPort)
	fmt.Println("-VIR: tcp4 handler listning on 0.0.0.0" + newVerifier.VerifierListenPort + "\n \n[I'm all ears]")

	//Mainloop listen for Aggregator handler
	AID := uint32(0) //if Distribute AID is needed
	for {
		AggregatorConn, err := VerifierListner.AcceptTCP()
		if err != nil {
			fmt.Println("During VerifierListner.AcceptTCP. \nABORT Connection.\n ERROR:", err)
		} else {
			AID++
			go verifier.AggregatorHandler(AggregatorConn, AID, &newVerifier, EdDSAkeyset, ctx, RedisClient)
		}
	}
}
