package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gRPC "github.com/Trindkr/DistributedSystemsExam2022/proto"

	"google.golang.org/grpc"
)

//Same principle as in Server. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var tcpServer = flag.String("server", "5400", "Tcp server")

var _ports [5]string = [5]string{*tcpServer, "5401", "5402", "5403", "5404"} //List of ports the client tries to connect to._ports

var ctx context.Context                                       //Client context
var servers []gRPC.MessageServiceClient                       //list of servers.
var ServerConn map[gRPC.MessageServiceClient]*grpc.ClientConn //Map of server connections

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("=================================--Client--======================================")

	//Setup method for the log.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)//connect to log.txt file
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	fmt.Println("================= join Server =================")
	ServerConn = make(map[gRPC.MessageServiceClient]*grpc.ClientConn) //Make map of server connections
	joinServer()                                                      //Method call
	defer closeAll()       //when main method exits, close all the connections to the servers      

	parseValueFromClient()									   
}

func joinServer() {
	//connect to server

	//dial options
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())

	//use context for timeout on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel() //cancel the connection when we are done

	for _, port := range _ports { //try to connect to all the ports in _ports with dial
		log.Printf("Client %s: Attempts to dial on port %s\n", *clientsName, port)
		conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", port), opts...) //dials the port with the given timeout
		if err != nil {
			log.Printf("Client %s: Fail to Dial : %v", *clientsName, err)//log error
			continue
		}
		log.Printf("Client %s: connected to servre on port %s\n", *clientsName, port)//Log succesful connection to server
		var s = gRPC.NewMessageServiceClient(conn)
		servers = append(servers, s)          //add the new MessageServiceClient
		ServerConn[s] = conn                  // maps the MessageServiceClient to its connection
		fmt.Println(conn.GetState().String()) //prints connected if it's connected
	}
	ctx = context.Background()
}

func parseValueFromClient(){

	reader := bufio.NewReader(os.Stdin) //reads input from console.
 	fmt.Println("In order to insert a key value pair \n Type \"Put key val\" where key is your key and val is the value. ")
 	fmt.Printf("")
	fmt.Println("In order to retrieve a value \n Type \"Get key \" where key is your key.")
	fmt.Println("--------------------")
	
	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		in, err := reader.ReadString('\n') //Read input into var in
		if err != nil {
			log.Fatal(err) //Log error
		}
		in = strings.TrimSpace(in) //Trim the clients input
		splitInput:= strings.Split(in," ")

		if strings.ToLower(splitInput[0]) == "get" {
			GetValue(splitInput[1])

		} else if strings.ToLower(splitInput[0]) == "put"{
			PutValue(splitInput[1],splitInput[2])

		} else {

			fmt.Println("Something went wrong, please ensure that your input is either a put or a get request.")
		}
	}

}

func GetValue(keyString string){

	keyValue, err := strconv.ParseInt(keyString, 10, 32) //Convert string to int
	if err != nil {
		log.Fatal(err)
	}
	
	//Create a instance of key type with the key value
	key := &gRPC.Key{
		Key: keyValue,
	}

	//This checks the connection to the servers before calling get method.
	for _, s := range servers {
		if conReady(s) { //If the connection to the server is ready
			fmt.Println(s)
			value, err := s.Get(ctx, key) //Make gRPC call to server with key, and recieve corrosponding value back.
			if err != nil {
				log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
				log.Println(err)
			}
			log.Printf("Client %s: Succesful Get request with key: %d", *clientsName, keyValue)
			fmt.Printf("Value: %d\n",value.Value) //Print the value in console.
		}
	}
}

func PutValue (keyString string, valueString string){

	keyVal, err := strconv.ParseInt(keyString, 10, 32) //Convert string to int
	if err != nil {
		log.Fatal(err)
	}

	valueVal, err := strconv.ParseInt(keyString, 10, 32) //Convert string to int
	if err != nil {
		log.Fatal(err)
	}

	keyValue := &gRPC.KeyValue{
		Key:   keyVal,
		Value: valueVal,
	}

	//This checks the connection to the servers before calling get method.
	for _, s := range servers {
		if conReady(s) { //If the connection to the server is ready
			fmt.Println(s)
			con, err := s.Put(ctx, keyValue) //Make gRPC call to server with keyValue, and recieve confirmation.
			if err != nil {
				log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
				log.Println(err)
			}
			if con.Confirmation{ //If confirmation bool is true print success
				log.Printf("Client %s: Succesful Pet request with key: %d and value: %d", *clientsName, keyValue.Key,keyValue.Value)
				fmt.Println("Put request was successful!")
			}else{ //else print unsuccessful
				log.Printf("Client %s: Unsuccesful Pet request with key: %d and value: %d", *clientsName, keyValue.Key,keyValue.Value)
				fmt.Println("Put request unsuccessful, please try again!")
			}
		}
	}
}

//Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.MessageServiceClient) bool {
	return ServerConn[s].GetState().String() == "READY"
}

func closeAll() {
	for _, c := range ServerConn {
		c.Close()
	}
}