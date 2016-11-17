// Writing files in Go follows similar patterns to the
// ones we saw earlier for reading.

package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    // "io/ioutil"
    "os"
)

type Message struct {
    ID int
    Msg string
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

var lastID = 0
var sliceMsg []Message

func main() {
    lastID = lastID+1
    msg := Message{lastID,"abc"}
    sliceMsg = append(sliceMsg, msg)

    // To start, here's how to dump a string (or just
    // bytes) into a file.
    // d1 := []byte(msg.Msg)
    // err := ioutil.WriteFile("/Users/Jun/git/project_e7o8_f8f9_r6c8_v4l8/dat1", d1, 0644)
    // check(err)

    // For more granular writes, open a file for writing.
    f, err := os.Create("/Users/Jun/git/project_e7o8_f8f9_r6c8_v4l8/dat2")
    check(err)

    // It's idiomatic to defer a `Close` immediately
    // after opening a file.
    defer f.Close()

    // You can `Write` byte slices as you'd expect.
    // d2 := []byte{115, 111, 109, 101, 10}
    // n2, err := f.Write(d2)
    // check(err)
    // fmt.Printf("wrote %d bytes\n", n2)

    // A `WriteString` is also available.
    // n3, err := f.WriteString(msg.Msg)
    // fmt.Printf("wrote %d bytes\n", n3)

    // Issue a `Sync` to flush writes to stable storage.
    // f.Sync()

    // `bufio` provides buffered writers in addition
    // to the buffered readers we saw earlier.
    w := bufio.NewWriter(f)
    n4, err := w.WriteString(msg.Msg + "\n")
    n5, err := w.WriteString("def" + "\n")
    fmt.Printf("wrote %d bytes\n", n4)
    fmt.Printf("wrote %d bytes\n", n5)

    // Use `Flush` to ensure all buffered operations have
    // been applied to the underlying writer.
    w.Flush()

    fmt.Println("Launching server...")   // listen on all interfaces   
    ln, _ := net.Listen("tcp", ":8081")   
    // accept connection on port   
       
    defer ln.Close()
    // run loop forever (or until ctrl-c)   
    for {   
        conn, err := ln.Accept()  
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
        // Handle connections in a new goroutine.
        go handleRequest(conn, w) 
    } 
     // See more at: https://systembash.com/a-simple-go-tcp-server-and-tcp-client/#sthash.Fq1Drdfk.dpuf

}

func collectMessages() {

}

func handleRequest(conn net.Conn, w *bufio.Writer) {
    // will listen for message to process ending in newline (\n)     
        message, _ := bufio.NewReader(conn).ReadString('\n')     // output message received     
        lastID = lastID + 1 
        msg := Message{lastID, message}
        fmt.Print("Message Received:", msg.Msg)     // sample process for string received    
        _,_ = w.WriteString(msg.Msg) 
        w.Flush()
        newmessage := strings.ToUpper(message)     // send new string back to client     
        conn.Write([]byte(newmessage + "\n"))  

}
