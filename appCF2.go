package main

import (
  "encoding/json"
  "fmt"
  "os"
  "strings"
  // "os/exec"
)

type CFReturnVal struct {
  JsonArgString  string // Json struct htat'll be passed to next hop
  Ip             string
  FunName        string
  ServiceFunName string
  CFInfo         ChainingFunctionInfo
  ReturnToOrigin bool
}

type ChainingFunctionInfo struct {
  GitRepo       string
  RepoName      string
  FileName      string
  CFName        string
  DebuggingPort string
  ClientIpPort  string
}

func formatJsonStringInput(str string) string {
  str = strings.Replace(str, "\\", "", -1)
  last := len(str) - 1
  str = str[1:last]
  return str
}

type ListOfMessages struct {
  Messages []Message
  ReturnAddress string
}

type Message struct {
  UserName string
  Message string
}

func main() {

  usage := fmt.Sprintf("Usage: %s CF Id\n", os.Args[0])
  if len(os.Args) != 3 {
    fmt.Printf(usage)
    os.Exit(1)
  }
  jsonString := os.Args[1]
  jsonString = strings.Replace(jsonString, "\\", "", -1)
  last := len(jsonString) - 1
  jsonString = jsonString[1:last]
  // jsonString = strings.Replace(jsonString, "\\", "", -1)

  cfInfoOrigString := os.Args[2]
  // fmt.Println("jsonstring is1234: ", jsonString)
  var lom ListOfMessages
  err := json.Unmarshal([]byte(jsonString), &lom)
  if err != nil {
    fmt.Println("ERROR 123!!: ", err, " json string is: ", jsonString )

    return
    // fmt.Println(err)
  }
  // fmt.Println(vReply)

  var origCFInfo ChainingFunctionInfo
  // fmt.Println("input CF: ", cfInfoOrigString)
  cfInfoOrigString = strings.Replace(cfInfoOrigString, "\\", "", -1)
  last = len(cfInfoOrigString) - 1
  cfInfoOrigString = cfInfoOrigString[1:last]
  

  err = json.Unmarshal([]byte(cfInfoOrigString), &origCFInfo)
  if err != nil {
    fmt.Println("cf info, ", err)
    return
  }  
  // fmt.Println("args is: ", args)
  
  // send the lom back
  b, err := json.Marshal(lom)

  var returnVal CFReturnVal
  returnVal.Ip = origCFInfo.ClientIpPort
  returnVal.JsonArgString = string(b)
  // fmt.Println("string b", returnVal.JsonArgString)
  // returnVal.FunArgs = args
  returnVal.ReturnToOrigin = true
  returnVal.ServiceFunName = "MsgServer.AppendMessages"
  origCFInfo.CFName = "appCF2.go"

  returnVal.CFInfo = origCFInfo
  buff, err := json.Marshal(returnVal)
  if err != nil {
    fmt.Println(err)
    return
  } else {
    fmt.Println(string(buff))
  }
  // str := string(buff)
  // fmt.Println(string(buff))

  // cfId := os.Args[1]
  // if cfId == "1" {
  //  func1()
  // } else {
  //  func2()
  // }

}
