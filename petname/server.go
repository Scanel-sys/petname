package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	petname "github.com/dustinkirkland/golang-petname"

	petnamepb "yadro.com/course/proto"
)

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
}

func (s *server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Generate(_ context.Context, in *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {

	if in.Words <= 0 {
		slog.Error("[ Generate ] Words <= 0")
		return nil, status.Error(codes.InvalidArgument, "words must be > 0")
	}

	name := petname.Generate(int(in.Words), in.Separator)

	return &petnamepb.PetnameResponse{
		Name: name,
	}, nil
}

func (s *server) GenerateMany(in *petnamepb.PetnameStreamRequest, stream petnamepb.PetnameGenerator_GenerateManyServer) error {
	if in.Words <= 0 {
		slog.Error("[ GenerateMany ] Words <= 0")
		return status.Error(codes.InvalidArgument, "words must be > 0")
	}

	if in.Names <= 0 {
		slog.Error("[ GenerateMany ] Names <= 0")
		return status.Error(codes.InvalidArgument, "names must be > 0")
	}

	for range int(in.Names) {
		if err := stream.Context().Err(); err != nil {
			slog.Error("[ GenerateMany ] Stream.Context().Err()", "error", err)
			return err
		}

		name := petname.Generate(int(in.Words), in.Separator)

		if err := stream.Send(&petnamepb.PetnameResponse{Name: name}); err != nil {
			slog.Error("[ GenerateMany ] Send error:", "error", err)
			return err
		}
	}

	return nil
}

func parsePort() (string, error) {
	configPath := flag.String("config", "", "config path")
	flag.Parse()

	cfg, err := LoadPetnameConfig(*configPath)
	if err != nil {
		return "", err
	}

	return cfg.Port, nil
}

func main() {
	port, err := parsePort()
	if err != nil {
		log.Fatalf("error parsing config %v", err)
	}
	slog.Info("config loaded", "port", port)

	var address string
	flag.StringVar(&address, "address", ":"+port, "server address")
	flag.Parse()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
