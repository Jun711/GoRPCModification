package main 

import (
	"net"
	"fmt"
)
	


func main(){
	govecConn, err := net.Dial("tcp", "128.189.221.124:2000")
		if err != nil {
			fmt.Println(err)
			// return nil
		}
		govecConn.Write([]byte("hey receieve this"))
}