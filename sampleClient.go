package main

import (
	"./rpcc"
	"encoding/json"
	"fmt"
	"net"
)

type ValReply struct {
	Reply string
}

type GetArgs struct {
	Key string
}

var debuggingMode = true

func main() {

	var args rpcc.ArgToSF

	var getArg GetArgs
	getArg.Key = "hello"

	b, err := json.Marshal(getArg)
	if err != nil {
		// do something
	}

	args.JsonArgString = string(b)

	var reply rpcc.ReturnValSF

	var cfInfo rpcc.ChainingFunctionInfo
	cfInfo.GitRepo = "https://github.com/Makoz/SampleRepo.git"
	cfInfo.FileName = "test2.go"
	cfInfo.CFName = "1"
	cfInfo.RepoName = "SampleRepo"
	cfInfo.ClientIpPort = "localhost:7000"
	// cfInfo.DebuggingPort = "localhost:5000"

	l, err := net.Listen("tcp", cfInfo.ClientIpPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	serv, _ := rpcc.Dial("tcp", "localhost:3000")
	serv.InitialCall("SampleServ.Reverse", args, &reply, cfInfo)

	for {
		fmt.Println("in for loop")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
		} else {
			fmt.Println("accepted a connection")
			buf := make([]byte, 65535)
			n, err := conn.Read(buf[:])
			buf = buf[:n]
			// check if its an errr struct
			// var errMsg rpcc.ErrorMsg
			var ret rpcc.ArgToSF
			err = json.Unmarshal(buf, &ret)
			if err != nil {
				fmt.Println("unmarshal err? ", err)
				// check fi its the value we want
			} else {
				fmt.Println(ret)
				// sucess, it was an error message
				
			}

		}
	}

	//serv.Call("SampleServ.Reverse", args, &reply, cfInfo)

	/*
	  serv.Call(ServiceFunction1)
	  An INtermediate layer that then calls ServiceFunction1
	  SF1 returns result to Intermediate Layer --> Chaning logic

	*/

}
