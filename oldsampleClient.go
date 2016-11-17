package main

import (
	"./rpcc"
	"encoding/json"
	//"fmt"
	//"net"
)

type ValReply struct {
	Reply string
}

type GetArgs struct {
	Key string
}

var debuggingMode = true

func main() {

	var args rpc.ArgToSF

	var getArg GetArgs
	getArg.Key = "hello"

	b, err := json.Marshal(getArg)
	if err != nil {
		// do something
	}

	args.JsonArgString = string(b)

	var reply rpc.ReturnValSF

	var cfInfo rpc.ChainingFunctionInfo
	cfInfo.GitRepo = "https://github.com/Makoz/SampleRepo.git"
	cfInfo.FileName = "test2.go"
	cfInfo.CFName = "1"
	cfInfo.RepoName = "SampleRepo"
	cfInfo.ClientIpPort = "localhost:7000"
	// cfInfo.DebuggingPort = "localhost:5000"

	serv, _ := rpc.Dial("tcp", "localhost:3000")
	// serv.InitialCall("SampleServ.Reverse", args, &reply, cfInfo)
	serv.Call("SampleServ.Reverse", args, &reply, cfInfo)

	/*
	   serv.Call(ServiceFunction1)
	   An INtermediate layer that then calls ServiceFunction1
	   SF1 returns result to Intermediate Layer --> Chaning logic

	*/

}
