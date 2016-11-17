package main 

import (
  "net" 
  "fmt" 
  "bufio" 
  "os" 
)
  

func main() {   

// connect to this socket   
  conn, err := net.Dial("tcp", "52.68.74.211:3001")   
  if err!= nil {
    fmt.Println(err)
  }

  // a go routine to read from server

 
  go func() {
    for {
      // message, err := bufio.NewReader(conn).ReadString('\n')
      scanner := bufio.NewScanner(conn)
      // if err != nil {
      //   fmt.Println(err)
      //   os.Exit(-1)
      // }     
      // fmt.Print("Message from server: "+message)   
      for scanner.Scan() {
        fmt.Println(scanner.Text()) // Println will add back the final '\n'
      }
      if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "reading standard input:", err)
      }
    }
  }()

  // reader := bufio.NewReader(os.Stdin) 
  // fmt.Print("Enter a username: ")
  // userName, err := reader.ReadString('\n')
  // fmt.Println("My name is: ", userName)
  // if err != nil {
  //   fmt.Println(err)
  //   os.Exit(-1)
  // }
  userNameReader := bufio.NewReader(os.Stdin)

  // userName := "Makoz"
  fmt.Print("Enter a username: ")
  userName, err := userNameReader.ReadString('\n')
  userName = userName[:len(userName)-1]
  // fmt.Println("My name is: ", userName)
  if err != nil {
    fmt.Println(err)
    os.Exit(-1)
  }
  for {     
  // read in input from stdin     
        
    // fmt.Print("Text to send: ")  
    reader := bufio.NewReader(os.Stdin) 
    // fmt.Printf(userName + ": ")
    text, err := reader.ReadString('\n')     // send to socket 
    if err != nil {
      fmt.Println(err)
      break;
    }
    // fmt.Println("im sending: " + userName + ": "+ text)    
    fmt.Fprintf(conn, userName + ": "+ text)     // listen for reply     
  } 
} 