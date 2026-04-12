package main

import (
	"blackhatgo/c2c/grpcapi"
	"blackhatgo/c2c/server"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const banner = `
   ____             _                     ____ ____  
  / ___| ___  _ __ | |__   ___ _ __      / ___|___ \ 
 | |  _ / _ \| '_ \| '_ \ / _ \ '__|____| |     __) |
 | |_| | (_) | |_) | | | |  __/ | |_____| |___ / __/ 
  \____|\___/| .__/|_| |_|\___|_|        \____|_____|
             |_|                                     
`

const version = "v1.0.0"

func printBanner() {
	separator := "============================================================"

	fmt.Println(separator)
	fmt.Print(strings.TrimLeft(banner, "\r\n"))
	fmt.Println(separator)
	fmt.Printf("  BUILD DATE : %s\n", time.Now().Format("2006-01-02"))
	fmt.Printf("  LISTENERS  : :4444 (Implant), :9090 (Admin)\n")
	fmt.Printf("  DATABASE   : sqlite3 (server.db)\n")
	fmt.Println(separator)
	fmt.Println()
}

func main() {
	var (
		implantListener, adminListener net.Listener
		err                            error
		opts                           []grpc.ServerOption
	)

	db, err := sql.Open("sqlite3", "server.db?_loc=auto&parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	err = server.CreateTables(db)
	if err != nil {
		log.Fatalln(err)
	}

	implantCreds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatalln(err)
	}
	implantOpts := append(opts, grpc.Creds(implantCreds))

	clientCreds, err := server.LoadTLSServerCreds()
	if err != nil {
		log.Fatalln(err)
	}
	clientOpts := append(opts, grpc.Creds(clientCreds))

	work, results := make(map[string]chan *grpcapi.Command), make(map[string]chan *grpcapi.Command)
	sessions := server.NewSessionManager(work, results)
	implants := make(map[uuid.UUID]time.Time)
	implant := server.NewImplantServer(sessions, implants, db)
	admin := server.NewAdminServer(sessions, implants, db)

	if implantListener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", 4444)); err != nil {
		log.Fatal(err)
	}
	if adminListener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", 9090)); err != nil {
		log.Fatal(err)
	}

	grpcAdminServer, grpcImplantServer := grpc.NewServer(clientOpts...), grpc.NewServer(implantOpts...)
	grpcapi.RegisterImplantServer(grpcImplantServer, implant)
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)

	printBanner()

	go func() {
		grpcImplantServer.Serve(implantListener)
	}()
	grpcAdminServer.Serve(adminListener)
}
