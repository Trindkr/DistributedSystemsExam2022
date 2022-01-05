package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	gRPC "github.com/Trindkr/DistributedSystemsExam2022/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedMessageServiceServer 	//You need this line if you have a server
	name string 							//Not required but useful if you want to name your server
	port string 							//Not required but useful if your server needs to know what port it's listening to
	mutex sync.Mutex 						//used to lock the server to avoid having multiple clients access the same resource at the same time.

	keyValueMap map[int64]int64 			//Map to store keyvalue pairs.
}

var server *Server

//flags used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
var serverName = flag.String("name", "default", "Senders name") 
var port = flag.String("port", "5400", "Server port")

var _ports [5]string = [5]string{*port, "5401", "5402", "5403", "5404"} //Loops through the hardcoded ports to see if it can listen on one of them._ports


func main() {
	
	setLog()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")


	go launchServer(_ports[:]) //starts a goroutine executing the launchServer method. Syntax note: [:] ensures that the entire array/slice is being sent.
	
	for {
		time.Sleep(time.Second*5) //This makes sure that the main method is "kept alive"/keeps running
	}
}

func (s *Server) Put(ctx context.Context, in *gRPC.KeyValue) (*gRPC.Confirmation, error) {

	s.mutex.Lock()//locks the server ensuring no one else insert a key value pair into the map.
	defer s.mutex.Unlock() //this unlocks the mutex when exiting the method

	s.keyValueMap[in.Key]=in.Value

	if s.keyValueMap[in.Key] != in.Value{
		return &gRPC.Confirmation{Confirmation: false}, nil //If the value we just inserted into the map isn't there, return false
	}
	return &gRPC.Confirmation{Confirmation: true}, nil //else return true.
	
}

func (s *Server) Get(ctx context.Context, in *gRPC.Key) (*gRPC.Value, error) { //Since we don't manipulate any data, no lock is needed here.
	
	var value int64 = s.keyValueMap[in.Key] //Get value from map
	return &gRPC.Value{Value: value},nil //return value
}

func launchServer(ports []string) {
	
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, ports[0])

	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", ports[0])) // Create listener tcp on given port or default port 5400
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		if len(ports) > 1 {
			launchServer(ports[1:]) //[1:] makes a slice from item at index 1 to the last index, instead of taking all items in the array/slice. It exclude item at index 0.
		} else {
			log.Fatalf("Server %s: Failed to find open port", *serverName)//if it fails to listen on all ports, log error message and kill process.
		}
	}

	var opts []grpc.ServerOption 			//Server options.
	grpcServer := grpc.NewServer(opts...) 	//makes gRPC server 
	server = newServer(ports[0]) 			//Item at index 0 is at this point the port which we succesfully listened to
	gRPC.RegisterMessageServiceServer(grpcServer, server) //We think this method takes our own server and puts it into a grpc server, but again we not sure :)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
	//code here is unreachable because serve occupies the current thread.
}

//creates a new instance of a server type
func newServer(serverPort string) *Server {
	s := &Server{
		name: *serverName, 		//* is used because of flags
		port: serverPort, 		// not sure why * isnt used here but it isn't
		keyValueMap: make(map[int64]int64), 		// initialize a map. If no value exists in the map, the default value is 0, ensuring propery 2.
	}

	fmt.Println(s) 				//prints the server struct to console
	return s 					//return server
}

func setLog(){
	//Clears the log.txt file when a new server is started
	//This is used for logging information to a log file.
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	//This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

}
