package main

import (
	"./rpcc"
	"encoding/json"
	"fmt"
	"log"
	"net"
	// "time"
)

type SampleServ int

type ValReply struct {
	Reply string
}

type GetArgs struct {
	Key string
}

type ErrorMsg struct {
	Err string
}

func (ser *SampleServ) Reverse(args *rpcc.ArgToSF, reply *rpcc.ReturnValSF) error {
	str := args.JsonArgString

	var getArgs GetArgs
	err := json.Unmarshal([]byte(str), &getArgs)
	if err != nil {
		// do something
	}

	r := []rune(getArgs.Key)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	fmt.Println("SF CALL IN SERVER 1")
	fmt.Println(string(r))

	var serviceFunctionFail = false

	if serviceFunctionFail {
		var errMsg ErrorMsg
		errMsg.Err = "FAKE ERROR IN SERVICE FUNCTION"
		a, err := json.Marshal(errMsg)
		if err == nil {
			fmt.Println("FAKE ERR IN SF: ")
			fmt.Println(string(a))
			reply.JsonArgString = string(a)
		}
	} else {
		var vReply ValReply
		vReply.Reply = string(r)
		b, err := json.Marshal(vReply)
		if err != nil {
		
		}
		reply.JsonArgString = string(b)
		fmt.Println(string(b))
	}


	// time.Sleep(10000 * time.Millisecond)
	return nil
}

func main() {
	sampleServ := new(SampleServ)
	rpcc.Register(sampleServ)
	fmt.Println("RUNNING SERVER 1 on port 3000")

	l, e := net.Listen("tcp", "localhost:3000")
	go rpcc.StartGoVectorPort("localhost:3001")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, _ := l.Accept()
		fmt.Println("Server 1 accepted a connection")
		go rpcc.ServeConn(conn)
	}
}
