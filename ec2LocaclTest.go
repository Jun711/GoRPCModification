package main 

import (
  "net" 
  "fmt" 
  "bufio" 
  "os" 
)
  

func main() {   

// connect to this socket   
  conn, _ := net.Dial("tcp", "128.189.220.122:4010")   

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
  
  for {     
  // read in input from stdin     
    reader := bufio.NewReader(os.Stdin)     
    // fmt.Print("Text to send: ")     
    text, err := reader.ReadString('\n')     // send to socket 
    if err != nil {
      fmt.Println(err)
      break;
    }    
    fmt.Fprintf(conn, text + "\n")     // listen for reply     
  } 
} 