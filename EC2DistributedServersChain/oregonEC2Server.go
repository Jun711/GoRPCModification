package main

import (
  "./rpcc"
  "encoding/json"
  "fmt"
  "log"
  "net"
  // "strings"
  // "io"
  "bufio"
  "sync"
  "time"
)

type MsgServer int

type ValReply struct {
  Reply string
}

type GetArgs struct {
  Key string
}

// should use a string array
type Message struct {
  UserName string
  Message  string
}

type ListOfMessages struct {
  Messages      []Message
  ReturnAddress string
}

var messageLocal []Message
var msgId int

var connections = struct {
  sync.RWMutex
  connections []net.Conn
}{}

// ip port?
var serverToLastMsg map[string]int

func (ser *MsgServer) AppendMessages(args *rpc.ArgToSF, reply *rpc.ReturnValSF) error {
  fmt.Println("IN APPEND MESSAGES")
  str := args.JsonArgString
  var lom ListOfMessages
  messageLocal = append(messageLocal, Message{"", "SERVER ONE MESSAGE"})
  err := json.Unmarshal([]byte(str), &lom)
  if err != nil {
    fmt.Println(err)
    // need to do something diff here
  }
  if lastMsgSentId, ok := serverToLastMsg[lom.ReturnAddress]; ok {
    // We've seen the server before!
    // lastMsgSentId++
    fmt.Println("MAP EXISTS *******")
    if lastMsgSentId < msgId {
      // we have new messages else do nothing
      fmt.Println("GOT MSG LOM MSG ID: ", lastMsgSentId)
      lom.Messages = append(lom.Messages, messageLocal[lastMsgSentId:msgId]...)
      fmt.Println("LOM MESSAGES ", lom)
      b, err := json.Marshal(lom)
      if err != nil {
        fmt.Println("ERROR IN MARSHALL", err)
        // do something
      }
      reply.JsonArgString = string(b)
      serverToLastMsg[lom.ReturnAddress] = msgId
    }
  } else {
    // new Server connected! Add it
    serverToLastMsg[lom.ReturnAddress] = msgId // TODO FIX
    reply.JsonArgString = args.JsonArgString
    // put shit in returnValsf
  }
  fmt.Println("WE GOOD")
  return nil
}

func handleRPCC() {
  l, e := net.Listen("tcp", "172.31.36.149:4000")
  if e != nil {
    log.Fatal("listen error:", e)
  }
  for {
    conn, _ := l.Accept()
    fmt.Println("Server 1 accepted a connection")
    go rpc.ServeConn(conn)
  }
}

var time1 time.Time
var time2 time.Duration
func GrabMessagesFromOtherServers() {
  // hardcoded to
  // every like 5 seconds or something but lets do one for now
  for {
    fmt.Println("SLEEPING")
    time.Sleep(5000 * time.Millisecond)
    var lom ListOfMessages
    lom.ReturnAddress = "52.35.92.187:4010"
    b, err := json.Marshal(lom)
    if err != nil {
      fmt.Println(err)
    }
    var args rpc.ArgToSF
    args.JsonArgString = string(b)

    var cfInfo rpc.ChainingFunctionInfo
    cfInfo.GitRepo = "https://github.com/Makoz/MessageServerCS416.git"
    cfInfo.FileName = "appCFTokyo.go"
    cfInfo.CFName = "1"
    cfInfo.RepoName = "MessageServerCS416"
    // add in debugging mode i guess...
    // cfInfo.ClientIpPort = "52.35.92.187:4010"
    cfInfo.ClientIpPort = "172.31.36.149:4010"

    var reply rpc.ReturnValSF
    // address of first hop is 4000
    serv, _ := rpc.Dial("tcp", "52.68.74.211:4001")
    time1 = time.Now()
    serv.InitialCall("MsgServer.AppendMessages", args, &reply, cfInfo)
  }
}

func ListenForReturnMesages() {
  fmt.Println("LISTENING FOR RETURN ESSAGES")
  l, err := net.Listen("tcp", "172.31.36.149:4010")

    if err != nil {
      fmt.Println(err)
      return
    }
  for {
      fmt.Println("listening for msgs")
      conn, err := l.Accept()
      if err != nil {
        fmt.Println("Error accepting: ", err)
      } else {
        fmt.Println("accepted a connection")
        buf := make([]byte, 65535)
        n, err := conn.Read(buf[:])
        buf = buf[:n]
        // check if its an errr struct
        // var errMsg rpc.ErrorMsg
        time2 = time.Since(time1)
        var ret rpc.ArgToSF
        err = json.Unmarshal(buf, &ret)
        if err != nil {
          fmt.Println("unmarshal err? ", err)
          // check fi its the value we want
        } else {
          // time2 = time.Since(time1)
          fmt.Println("TIME ELAPSED : ", time2)
          fmt.Println(ret)
          // write this hsit to every client
          var arg ListOfMessages
          err := json.Unmarshal([]byte(ret.JsonArgString), &arg)
          if err != nil {
            fmt.Println(err)
          }
          fmt.Println(arg)

          connections.Lock()

          fmt.Println("MY MESSAGES ", arg.Messages, "MY CONNECTIONS: ", connections)
          for _, ele := range connections.connections {
            for _, msg := range arg.Messages {
              fmt.Println("WRITING MSG TO CLIENT: ", msg)
              length := len(msg.Message)
              ele.Write([]byte(msg.Message[:(length-1)] + "\n"))
            }
          }
          connections.Unlock()
          // sucess, it was an error message

        }

      }
    }
  }



func main() {
  serverToLastMsg = make(map[string]int)
  msgId = 0

  appServ := new(MsgServer)
  rpc.Register(appServ)
  fmt.Println("RUNNING MSG SERVER 1 on port 3000")

  // w := bufio.NewWriter(f)
  go handleRPCC()
  l, _ := net.Listen("tcp", "172.31.36.149:3000")
  defer l.Close()
  go GrabMessagesFromOtherServers()
  go ListenForReturnMesages()

  for {

    conn, err := l.Accept()
    fmt.Println("accepted a connection")
    if err != nil {
      log.Fatal(err)
    }

    fmt.Println("connections: ", connections.connections)
    go func(c net.Conn) {
      // Echo all incoming data.
      connections.Lock()
      connections.connections = append(connections.connections, conn)
      connections.Unlock()
      for {
        // check if conn is still active
        message, err := bufio.NewReader(conn).ReadString('\n') // output message received
        if err != nil {
          fmt.Println(err)
          // remove the connection
          connections.Lock()
          for idx, ele := range connections.connections {
            if ele == c {
              fmt.Println("Found a matching connection, remove it")
              connections.connections = append(connections.connections[:idx], connections.connections[idx+1:]...)
              break
            }
          }
          connections.Unlock()
          c.Close()
          break
        }
        // lastID = lastID + 1
        fmt.Print("Message Received from client:", message) // sample process for string received
        if message != "\n" {
          fmt.Println("APPENDING ")

          connections.Lock()
          fmt.Println(connections.connections, " current connections")
          messageLocal = append(messageLocal, Message{"", message})
          fmt.Println(messageLocal)
          msgId++
          for _, ele := range connections.connections {
            if ele != c {
              ele.Write([]byte(message + "\n"))
            }
          }
          connections.Unlock()
          // defer c.Close()
          // io.Copy(c, c)
        }
      }

      // Shut down the connection.

    }(conn)

  }

  //
}
