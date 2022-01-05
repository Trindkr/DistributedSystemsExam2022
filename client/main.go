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

