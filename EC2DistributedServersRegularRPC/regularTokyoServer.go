package main

import (
  "net/rpc"
  "encoding/json"
  "fmt"
  "log"
  "net"
  // "os"
  // "strings"
  // "io"
  "bufio"
  "time"
  "sync"
)

type MsgServer int

type ArgToSF struct {
  JsonArgString string
}

type ReturnValSF struct {
  JsonArgString string
}

// should use a string array
type Message struct {
  UserName string
  Message string
}

type ListOfMessages struct {
  Messages []Message
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

func (ser *MsgServer) AppendMessages(args *ArgToSF, reply *ReturnValSF) error {
  fmt.Println("IN APPEND MESSAGES")
  str := args.JsonArgString
  var lom ListOfMessages
  // messageLocal = append(messageLocal, Message{"", "SERVER TWO MESSAGE"})
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
    } else {
      reply.JsonArgString = args.JsonArgString
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
  l, e := net.Listen("tcp", "172.31.25.85:4001")
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
    
    // lom.Return address only useful in this case as a unique ID
    lom.ReturnAddress = "52.68.74.211:4010"
    b, err := json.Marshal(lom)
    if err != nil {
      fmt.Println(err)
    }
    var args ArgToSF
    args.JsonArgString = string(b)
    var reply ReturnValSF
    // address of first hop is 4000
    serv, err := rpc.Dial("tcp", "52.79.92.65:4002")
    if err != nil {
      for {
        serv, err = rpc.Dial("tcp", "52.79.92.65:4002")
        if err == nil {
          break;
        }
      }
    }
    time1 = time.Now()
    err = serv.Call("MsgServer.AppendMessages", &args, &reply)

    // err = client.Call("FortuneServerRPC.GetFortuneInfo", remoteClientAddress, &reply)
    if err != nil {
      fmt.Println(err)

    } else {
      
      var arg2 ListOfMessages
      err := json.Unmarshal([]byte(reply.JsonArgString), &arg2)
      b, err = json.Marshal(arg2)
      if err != nil {
        fmt.Println(err)
      }
      args.JsonArgString = string(b)
      var reply ReturnValSF
      serv, err = rpc.Dial("tcp", "52.35.92.187:4000")
      if err != nil {
        for {
          serv, err = rpc.Dial("tcp", "52.35.92.187:4000")
          if err == nil {
            break;
          }
        }
      }
      err = serv.Call("MsgServer.AppendMessages", &args, &reply)
      time2 = time.Since(time1)
      fmt.Println("TIME ELAPSED : ", time2)
      if err != nil {
        fmt.Println(reply)
      } else {
        // print message to clients
        var arg3 ListOfMessages
        err := json.Unmarshal([]byte(reply.JsonArgString), &arg3)
        if err != nil {
          fmt.Println(err)
        } else {
          connections.Lock()

          fmt.Println("MY MESSAGES ", arg3.Messages, "MY CONNECTIONS: ", connections)
          for _, ele := range connections.connections {
            for _, msg := range arg3.Messages {
              fmt.Println("WRITING MSG TO CLIENT: ", msg)
              length := len(msg.Message)
              ele.Write([]byte(msg.Message[:(length-1)] + "\n"))
            }
          }
          connections.Unlock()
        }
        
      }

      
    }
  }
}



func main() {
  serverToLastMsg = make(map[string]int)
 
  appServ := new(MsgServer)
  rpc.Register(appServ)
  fmt.Println("RUNNING MSG SERVER 1 on port 3001")

  // w := bufio.NewWriter(f)
  go handleRPCC()
  go GrabMessagesFromOtherServers()
  l, _ := net.Listen("tcp", "172.31.25.85:3001")
  defer l.Close()
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
        message, err := bufio.NewReader(conn).ReadString('\n')     // output message received 
        if err != nil {
          fmt.Println(err)
          // remove the connection
          connections.Lock()
          for idx, ele := range connections.connections {
            if ele == c {
              fmt.Println("Found a matching connection, remove it")
              connections.connections = append(connections.connections[:idx], connections.connections[idx+1:]...)
              break;
            }
          }
          connections.Unlock()
          c.Close()
          break;
        }    
        // lastID = lastID + 1 
        fmt.Print("Message Received from client:", message)     // sample process for string received    
        connections.Lock()
        fmt.Println(connections.connections, " current connections")
        messageLocal = append(messageLocal, Message{"", message})
        fmt.Println("MSG LOCAL: ", messageLocal)
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
      
      // Shut down the connection.
      
    }(conn)


  }

  // 
}


