package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "github.com/hypnoglow/kubernetes-cookbook/grpc-load-balancing/dns-resolver/greeting"
)

func main() {
	target := os.Getenv("TARGET")
	workersEnv := os.Getenv("WORKERS")

	workers, err := strconv.Atoi(workersEnv)
	if err != nil {
		panic(err)
	}

	grpclog.SetLogger(log.New(os.Stdout, "grpc ", log.LstdFlags))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	cc, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBalancerName("round_robin"),
		// NOTE: FailFast option can be used to prevent RPC errors in situations
		// when server instances are down/unreachable. Instead, the client
		// will try to reconnect infinitely without returning an error to the caller.
		// The only way to cancel such RPC is to use context with cancel/timeout.
		//
		// You probably don't want this for clients that don't need to persist
		// connection, e.g. REST APIs where the call can be "safely" failed
		// and optionally retried on a higher level.
		//
		// You probably want this for clients that need to persist connection,
		// e.g. consumers or watchers.
		grpc.WithDefaultCallOptions(grpc.FailFast(false)),
	)
	cancel()
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	ctx, cancel = context.WithCancel(context.Background())
	jobs := gen(ctx, &wg)

	wg.Add(1)
	go handleSignals(&wg, cancel)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(&wg, jobs, pb.NewGreeterClient(cc))
	}

	wg.Wait()
}

func handleSignals(wg *sync.WaitGroup, cancel func()) {
	defer wg.Done()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	cancel()
}

func gen(ctx context.Context, wg *sync.WaitGroup) chan int {
	jobs := make(chan int, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)

		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			jobs <- i
			i++
		}
	}()

	return jobs
}

func worker(wg *sync.WaitGroup, jobs <-chan int, client pb.GreeterClient) {
	defer wg.Done()

	for job := range jobs {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		reply, err := client.Greet(ctx, &pb.GreetRequest{
			Name: fmt.Sprintf("Job-%d", job),
		})
		if err != nil {
			log.Printf("[ERROR] Job number %d: %s", job, err)
			continue
		}

		log.Printf("[INFO] Reply for job %d: %s", job, reply.GetGreeting())
	}
}
