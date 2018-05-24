package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "github.com/hypnoglow/kubernetes-cookbook/grpc-load-balancing/dns-resolver/greeting"
)

func main() {
	ip := os.Getenv("POD_IP")
	port := os.Getenv("PORT")

	srv := grpc.NewServer(
		// These keepalive parameters force clients to reconnect and possibly
		// discover new server replicas.
		//
		// This solves the problem when your deployment got scaled, but your
		// grpc clients will never connect to new replicas because grpc-go DNS
		// resolver does not trigger in this case.
		//
		// For details, see: https://github.com/grpc/grpc-go/issues/1663
		//
		// WARNING: there is currently a bug https://github.com/grpc/grpc-go/issues/1795
		// where the client will not reconnect to the server when it becomes
		// available.
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      time.Minute * 2,
			MaxConnectionAgeGrace: time.Minute * 1,
		}),
	)

	pb.RegisterGreeterServer(srv, Greeter{ip: ip})

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go serve(&wg, srv, lis)

	wg.Add(1)
	go handleSignals(&wg, srv, ip)

	wg.Wait()
}

func serve(wg *sync.WaitGroup, srv *grpc.Server, lis net.Listener) {
	defer wg.Done()

	if err := srv.Serve(lis); err != nil {
		log.Printf("[ERROR] serve failed: %s", err)
	}
}

func handleSignals(wg *sync.WaitGroup, srv *grpc.Server, ip string) {
	defer wg.Done()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	log.Printf("[INFO] Shutting down %s", ip)
	srv.GracefulStop()
}

type Greeter struct {
	ip string
}

func (srv Greeter) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetReply, error) {
	log.Printf("[INFO] got name %s on %s", req.GetName(), srv.ip)

	time.Sleep(time.Second * 1) // Just for demonstration purposes.

	greeting := fmt.Sprintf("Hello %s from %s", req.GetName(), srv.ip)

	return &pb.GreetReply{Greeting: greeting}, nil
}
