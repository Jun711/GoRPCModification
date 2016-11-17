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

	var vReply ValReply
	vReply.Reply = string(r)
	b, err := json.Marshal(vReply)
	if err != nil {
		// do something
	}
	reply.JsonArgString = string(b)
	fmt.Println(string(b))
	// time.Sleep(10000 * time.Millisecond)
	return nil
}

func main() {
	sampleServ := new(SampleServ)
	rpcc.Register(sampleServ)

	fmt.Println("RUNNING SERVER 2 on PORT 4000")

	l, e := net.Listen("tcp", "localhost:4000")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, _ := l.Accept()
		fmt.Println("Server 2 accepted a connection")
		go rpcc.ServeConn(conn)
	}
}
