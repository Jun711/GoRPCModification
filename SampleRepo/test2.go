package main

import (
	"encoding/json"
	"fmt"
	"os"
	// "os/exec"
)

type CFReturnVal struct {
	Ip      string
	FunName string
	FunArgs interface{}
}

type GetArgs struct {
	Key string
}

func func1() {
	var args GetArgs
	args.Key = "hi"
	var returnVal CFReturnVal
	returnVal.Ip = "localhost:4000"
	returnVal.FunName = "1"
	returnVal.FunArgs = args

	buff, _ := json.Marshal(returnVal)
	fmt.Println(string(buff))
}

func func2() {

	fmt.Println("func2 from test2")
}

func main() {

	usage := fmt.Sprintf("Usage: %s CF Id\n", os.Args[0])
	if len(os.Args) != 2 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	cfId := os.Args[1]
	if cfId == "1" {
		func1()
	} else {
		func2()
	}

}
