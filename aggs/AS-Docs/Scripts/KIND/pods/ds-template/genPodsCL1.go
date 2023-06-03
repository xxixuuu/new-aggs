package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

func useIoutilReadFile(fileName string) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
func main() {
	var err error

	targetSignerNum, err := strconv.Atoi(os.Args[1])
	if err != nil {
		println(err)
		return
	}

	err = exec.Command("bash", "-c", "mkdir "+os.Args[1]+"-CL1").Run()
	if err != nil {
		panic(err)
	}

	for i := 0; i < targetSignerNum; i++ {
		statement1 := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: signer-aggs-"
		statement2 := "  labels:\n    role: woker\nspec:\n    containers:\n      - name: signer\n        image: aggs2:signerTM0-2\n        ports:\n          - name: aggs\n            containerPort: 3000\n            protocol: TCP\n        resources:\n        args: [\"10.96.173.66\", \"3000\", \"3000\" , \"0\" , \"-\", \"" + fmt.Sprint(i+1)
		statement3 := "]\n"
		statement4 := "      - name: dummy\n        image: aggs2:dummyTM0-600\n        ports:\n          - name: aggs\n            containerPort: 3000\n            protocol: TCP\n        args: [\"-\", \"3000\", \"0\", \"100\", \"5\"]" + "\n    restartPolicy: OnFailure\n    nodeSelector:\n      aggs: sir" + fmt.Sprint(i%8+1)

		f, err := os.Create("./" + os.Args[1] + "-CL1" + "/signerDummy" + fmt.Sprint(i) + ".yaml")
		if err != nil {
			println(err)
			return
		}
		f.Write([]byte(statement1 + fmt.Sprint(i) + "\n" + statement2 + "\"" + statement3 + statement4))
		f.Close()
	}

}
