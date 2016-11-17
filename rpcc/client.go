// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpcc

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	// "github.com/arcaneiceman/GoVector/govec"
	"io"
	"log"
	"net"
	"net/http"
	// "reflect"
	"strings"
	"sync"
	"time"
)

type KillHeartBeat struct {
	Kill bool
}

type Msg struct {
	Content, RealTimestamp string
}

func (m Msg) String() string {
	return "content: " + m.Content + "\ntime: " + m.RealTimestamp
}

type ChainingFunctionInfo struct {
	GitRepo       string
	RepoName      string
	FileName      string
	CFName        string
	DebuggingPort string
	ClientIpPort  string
	NextIpPort	  string
}

// TCPConn for goVector
var GovecConn *net.TCPConn
// var ClientLogger *govec.GoLog

// ServerError represents an error that has been returned from
// the remote side of the RPC connection.
var debugLog = false

var numberOfAttempts = 0
var initialCall Call

// Reverse returns its argument string reversed rune-wise left to right.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	fmt.Println("In here")
	return string(r)
}

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")

// Call represents an active RPC.
type Call struct {
	ServiceMethod string      // The name of the service and method to call.
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call  // Strobes when call is complete.
	CFInfo        ChainingFunctionInfo
}

type ArgToSF struct {
	JsonArgString string
}

type ReturnValSF struct {
	JsonArgString string
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	codec ClientCodec

	reqMutex sync.Mutex // protects following
	request  Request

	mutex    sync.Mutex // protects following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

// A ClientCodec implements writing of RPC requests and
// reading of RPC responses for the client side of an RPC session.
// The client calls WriteRequest to write a request to the connection
// and calls ReadResponseHeader and ReadResponseBody in pairs
// to read responses.  The client calls Close when finished with the
// connection. ReadResponseBody may be called with a nil
// argument to force the body of the response to be read and then
// discarded.
type ClientCodec interface {
	// WriteRequest must be safe for concurrent use by multiple goroutines.
	WriteRequest(*Request, interface{}, ChainingFunctionInfo) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(interface{}) error

	Close() error
}

const (  
   GOVEC_PORT = "2700"
)

func (client *Client) send(call *Call) {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}
	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	// Encode and send the request.
	client.request.Seq = seq
	client.request.CFInfo = call.CFInfo
	client.request.ServiceMethod = call.ServiceMethod
	err := client.codec.WriteRequest(&client.request, call.Args, call.CFInfo)
	// err = client.codec.WriteRequest(&client.request, call.CFInfo)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) input() {
	var err error
	var response Response
	for err == nil {
		response = Response{}
		err = client.codec.ReadResponseHeader(&response)
		if err != nil {
			break
		}
		seq := response.Seq
		client.mutex.Lock()
		call := client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		switch {
		case call == nil:
			// We've got no pending call. That usually means that
			// WriteRequest partially failed, and call was already
			// removed; response is a server telling us about an
			// error reading request body. We should still attempt
			// to read error body, but there's no one to give it to.
			err = client.codec.ReadResponseBody(nil)
			if err != nil {
				err = errors.New("reading error body: " + err.Error())
			}
		case response.Error != "":
			// We've got an error response. Give this to the request;
			// any subsequent requests will get the ReadResponseBody
			// error if there is one.
			call.Error = ServerError(response.Error)
			err = client.codec.ReadResponseBody(nil)
			if err != nil {
				err = errors.New("reading error body: " + err.Error())
			}
			call.done()
		default:
			err = client.codec.ReadResponseBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// Terminate pending calls.
	client.reqMutex.Lock()
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.mutex.Unlock()
	client.reqMutex.Unlock()
	if debugLog && err != io.EOF && !closing {
		log.Println("rpcc: client protocol error:", err)
	}
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here.  It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		if debugLog {
			log.Println("rpcc: discarding Call reply due to insufficient Done chan capacity")
		}
	}
}

// NewClient returns a new Client to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
func NewClient(conn io.ReadWriteCloser) *Client {
	encBuf := bufio.NewWriter(conn)
	client := &gobClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf}
	return NewClientWithCodec(client)
}

// NewClientWithCodec is like NewClient but uses the specified
// codec to encode requests and decode responses.
func NewClientWithCodec(codec ClientCodec) *Client {
	client := &Client{
		codec:   codec,
		pending: make(map[uint64]*Call),
	}
	go client.input()
	return client
}

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(r *Request, body interface{}, cfInfo ChainingFunctionInfo) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	if err = c.enc.Encode(cfInfo); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *gobClientCodec) ReadResponseHeader(r *Response) error {
	return c.dec.Decode(r)
}

func (c *gobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobClientCodec) Close() error {
	return c.rwc.Close()
}

// DialHTTP connects to an HTTP RPC server at the specified network address
// listening on the default HTTP RPC path.
func DialHTTP(network, address string) (*Client, error) {
	return DialHTTPPath(network, address, DefaultRPCPath)
}

// DialHTTPPath connects to an HTTP RPC server
// at the specified network address and path.
func DialHTTPPath(network, address, path string) (*Client, error) {
	var err error
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn), nil
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()
	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string) (*Client, error) {
	// ClientLogger = govec.Initialize("client", "ClientLog")
	// str := strings.Split(address, ":")
	// govecPort := str[0] + ":" + GOVEC_PORT
	// GovecConn = SetupGovecConn(govecPort)
	// str := strings.Split(address, ":")
	// serverGoVecPort := str[0] + ":" + GOVEC_PORT
	// fmt.Println("serverGoVecPort Dial", serverGoVecPort)
	// e := SendGoVecMsgAndClosePort("dialing-a-server", address)
	// if e != nil {
	// 	fmt.Println("dial", e)
	// }

	conn, err := net.Dial(network, address)

	if err != nil {
		return nil, err
	}
	// outgoingMessage := Msg{"Dialing!", time.Now().String()}
	// outBuf := Logger.PrepareSend("Dialing success", outgoingMessage)
	// _, errWrite := conn.Write(outBuf)
	// print(errWrite)
	// ClientLogger.LogLocalEvent("Dialing success")
	return NewClient(conn), nil
}

func (client *Client) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	return client.codec.Close()
}

// Go invokes the function asynchronously.  It returns the Call structure representing
// the invocation.  The done channel will signal when the call is complete by returning
// the same Call object.  If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call, cfInfo ChainingFunctionInfo) {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	call.CFInfo = cfInfo
	// outgoingMessage := Msg{"Hello GoVec!", time.Now().String()}
	// outBuf := Logger.PrepareSend("Contacting-a-server", outgoingMessage)
	// // fmt.Println("out-str: ", string(outBuf))
	// _, errWrite := GovecConn.Write(outBuf)
	// if errWrite != nil {
	// 	fmt.Println("GoVec Write Error:", errWrite)
	// }
	// str := strings.Split(cfInfo.ClientIpPort, ":")
	// serverGoVecPort := str[0] + ":" + GOVEC_PORT
	e := SendGoVecMsgAndClosePort("calling-a-server-service", cfInfo.NextIpPort)
	if e != nil {
		fmt.Println("Go ", e)
	}

	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel.  If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpcc: done channel is unbuffered")
		}
	}
	call.Done = done
	
	client.send(call)
	// fmt.Println("Call was sent")
	// return call
}



func SetupGovecConn(servAddr string) net.Conn {
	// strEcho := "Halo"
	// tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	// if err != nil {
 //        fmt.Println("ResolveTCPAddr failed:", err.Error())
 //        return nil
 //    }

	// conn, e := net.DialTCP("tcp", nil, tcpAddr)
	conn, e := net.Dial("tcp", servAddr)
	if e != nil {
		fmt.Println("error dialing to govec port: ", e)
		return nil
	}
	// fmt.Println("Setting up GovecConn")

	// _, e := net.Dial("tcp", "localhost:3001")
	// if e != nil {
	// 	fmt.Println("error dialing to govec port: ", e)
	// }


	// n, err := conn.Write([]byte(strEcho))
 //    if err != nil {
 //        fmt.Println("Write to server failed:", err.Error())
 //        // os.Exit(1)
 //    }

 //    fmt.Println("write to server = ", n)

	// fmt.Println(reflect.TypeOf(conn))
	return conn
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}, cfInfo ChainingFunctionInfo) {
	// call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	// fmt.Println(cfInfo)
	client.Go(serviceMethod, args, reply, make(chan *Call, 1), cfInfo)
	// go func() {
	// 	fmt.Println(client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done)
	// }()
	// return call.Error
}

func (client *Client) InitialCall(serviceMethod string, args interface{}, reply interface{}, cfInfo ChainingFunctionInfo) {
	initialCall.ServiceMethod = serviceMethod
	initialCall.Args = args
	initialCall.Reply = reply
	initialCall.CFInfo = cfInfo

	str := strings.Split(cfInfo.ClientIpPort, ":")
	heartbeatPort := str[0] + ":6000"

	client.Go(serviceMethod, args, reply, make(chan *Call, 1), cfInfo)
	numberOfAttempts++
	if numberOfAttempts == 1 {
		go client.StartHeartbeatPort(heartbeatPort)
	}
	// fmt.Println("debug", cfInfo.DebuggingPort)
	if cfInfo.DebuggingPort != "" {
		client.StartDebugPort(cfInfo.DebuggingPort)
	}
}

func (client *Client) DebugCall(conn io.ReadWriteCloser) {
	var buf [65535]byte
	n, err := conn.Read(buf[:])
	//fmt.Println("server sent to client:",n, string(buf[:]))

	var returnVal CFReturnVal

	// err = json.Unmarshal(buf[23:n], &returnVal)
	err = json.Unmarshal(buf[:n], &returnVal)
	if err != nil {
		fmt.Println("Error when attempting to unmarshal data from Chaining Function err: ", err)
		return
	}

	// var args GetArgs
	var args ArgToSF
	args.JsonArgString = returnVal.JsonArgString
	// MapToStruct(returnVal.FunArgs, &args)
	//fmt.Println("DEBUG: ABOUT TO CALL IP: ", returnVal.Ip)
	serv, err := Dial("tcp", returnVal.Ip)
	if err != nil {
		fmt.Println("Error on Dial to ", returnVal.Ip, " with error: ", err)
		return
	}

	//fmt.Println("SUCCESSFUL DIAL to ", returnVal.Ip)

	// var reply ValReply
	var reply ReturnValSF
	// reply.Reply = ""

	// fmt.Println("args type is: ", reflect.TypeOf(returnVal.FunArgs))

	// fmt.Println("ABOUT TO CALL SF FROM CLIENT IN DEBUG MODE")

	serv.Call(returnVal.ServiceFunName, args, &reply, returnVal.CFInfo)
}

func (client *Client) StartDebugPort(debugIp string) {
	l, e := net.Listen("tcp", debugIp)
	if e != nil {
		fmt.Println("listen error on debug port: ", e)
	}
	fmt.Println("LISTENING IN DEBUG MODE AT PORT 5000")
	for {
		conn, _ := l.Accept()

		// fmt.Println("ACCEPT DEBUG")
		client.DebugCall(conn)
	}
}

var lastTimestamp time.Time

func (client *Client) StartHeartbeatPort(heartbeatIp string) {

	str := strings.Split(heartbeatIp, ":")
	heartbeatPort := str[0] + ":6000"

	l, e := net.Listen("tcp", heartbeatPort)
	if e != nil {
		fmt.Println("listen error on heartbeat port: ", e)
	}
	fmt.Println("LISTENING IN HEARTBEAT MODE AT PORT 6000")


	closeTimeoutCh := make(chan bool)

	go func () {
	var timeout = 20
	lastTimestamp = time.Now()
	for {
		select {
		case <- closeTimeoutCh:
			fmt.Println("CLOSING TIMEOUT!!")
			return
		default:
			time.Sleep(1*time.Second)
			//fmt.Println("since: ", time.Since(lastTimestamp))
			//fmt.Println("duration: ", time.Duration(timeout)*time.Second)
			if time.Since(lastTimestamp) > time.Duration(timeout)*time.Second {
			//TIMEOUT
			//RETRY CALL
			//fmt.Println("NUM OF ATTEMPTS: ", numberOfAttempts)
			if numberOfAttempts > 1 {
				//ERROR CHAIN HAS BROKEN THROW EXCEPTION
				// RESTART WITH DEBUG ON
				fmt.Println("ERROR: Chain has broken, restarting chain with debug on")
				debugPort := str[0] + ":5000"
				initialCall.CFInfo.DebuggingPort = debugPort
				client.InitialCall(initialCall.ServiceMethod, initialCall.Args, initialCall.Reply, initialCall.CFInfo)
			} else {
				//RESTART
				fmt.Println("Server timeout, retrying once")
				client.InitialCall(initialCall.ServiceMethod, initialCall.Args, initialCall.Reply, initialCall.CFInfo)
			}
			}
		}


	}
	}()





	for {
		// fmt.Println("INSIDE FOR")

		// fmt.Println("AFTER ACCEPT")

		conn, err := l.Accept()
		buf := make([]byte, 65535)
		n, err := conn.Read(buf[:])
		buf = buf[:n]
		var killBeat KillHeartBeat
		err = json.Unmarshal(buf, &killBeat)
		if err == nil {
			fmt.Println("Stopping heartbeat")
			closeTimeoutCh <- true
			return
		}

		lastTimestamp = time.Now()
		fmt.Println("ACCEPT HEARTBEAT at ", lastTimestamp)
	}
	//}()
}

