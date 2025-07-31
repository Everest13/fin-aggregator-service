package server

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/server/handler"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

type Opts struct {
	GrpcPort        string
	GrpcNetwork     string
	HttpPort        string
	HttpHost        string
	GracefulTimeout time.Duration
}

type Server struct {
	opts       *Opts
	impl       *handler.FinAggregatorServer
	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewServer(impl *handler.FinAggregatorServer, opts *Opts) *Server {
	return &Server{
		impl: impl,
		opts: opts,
	}
}

func (s *Server) Run() error {
	err := s.runGRPC()
	if err != nil {
		log.Printf("failed to run gRPC server: %v", err)
		return err
	}

	err = s.runHTTP()
	if err != nil {
		log.Printf("failed to run HTTP server: %v", err)
		return err
	}

	return nil
}

func (s *Server) runGRPC() error {
	s.grpcServer = grpc.NewServer()

	pb.RegisterFinAggregatorServiceServer(s.grpcServer, s.impl)

	lis, err := net.Listen(s.opts.GrpcNetwork, s.opts.GrpcPort)
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return err
	}

	go func() {
		log.Printf("gRPC server started on %s", s.opts.GrpcPort)
		if err = s.grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server failed: %v", err)
			s.stopGRPC()
		}
	}()

	return nil
}

func (s *Server) stopGRPC() {
	s.grpcServer.GracefulStop()
	log.Println("gRPC server gracefully stopped")
}

func (s *Server) runHTTP() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterFinAggregatorServiceHandlerFromEndpoint(ctx, mux, s.opts.GrpcPort, opts)
	if err != nil {
		return err
	}

	s.httpServer = &http.Server{
		Addr:    s.opts.HttpPort,
		Handler: mux,
	}

	log.Printf("Starting HTTP server on port %s\n", s.opts.HttpPort)
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Fatalf("failed to serve HTTP server: %v", err)
	}

	return nil
}

func (s *Server) stopHTTP() {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.GracefulTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	} else {
		log.Println("HTTP server gracefully stopped")
	}
}
