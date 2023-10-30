package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vpaklatzis/conduit-go/api"
	"github.com/vpaklatzis/conduit-go/config"
	"github.com/vpaklatzis/conduit-go/grpcapi"
	"github.com/vpaklatzis/conduit-go/logger"
	"github.com/vpaklatzis/conduit-go/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	var (
		log          logger.Logger
		conf         config.Config
		work, output chan *pb.Command
	)
	env := os.Getenv("ENVIRONMENT")
	if env == "" || env == "dev" {
		env = "dev"
		Logger, _ := zap.NewDevelopment()
		defer func(logger *zap.Logger) {
			err := logger.Sync()
			if err != nil {
				log.Fatal("Failed to sync logger")
			}
		}(Logger)
		log = Logger.Sugar()
		conf = config.LoadConfig("dev.env", "./env")
	} else {
		Logger, _ := zap.NewProduction()
		defer func(logger *zap.Logger) {
			err := logger.Sync()
			if err != nil {
				log.Fatal("Failed to sync logger")
			}
		}(Logger)
		log = Logger.Sugar()
		conf = config.LoadConfig("test.env", "./env")
	}
	work, output = make(chan *pb.Command), make(chan *pb.Command)

	startNewGrpcAdminServer(work, output, conf, log)

	startNewGrpcImplantServer(work, output, conf, log)

	srv := configureNewHttpServer(conf, log)

	// Implemented graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listening on: %s", err)
		}
	}()
	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so I don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")
	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	log.Info("Server exited.")
}

func configureNewHttpServer(conf config.Config, log logger.Logger) *http.Server {
	server := api.NewServer(conf, log)
	server.MountHandlers()

	addr := fmt.Sprintf("%s:%s", conf.HttpHost, conf.HttpPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: server.Router(),
	}
	return srv
}

func startNewGrpcAdminServer(work chan *pb.Command, output chan *pb.Command, conf config.Config, log logger.Logger) {
	service := grpc.NewServer()
	adminServer := grpcapi.NewAdminServer(work, output, conf, log)

	pb.RegisterAdminServer(service, adminServer)
	reflection.Register(service)

	addr := fmt.Sprintf("%s:%s", conf.GrpcAdminHost, conf.GrpcAdminPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Admin listener failed to bind to port %s: ", conf.GrpcAdminPort, err)
	}
	log.Infof("Starting gRPC admin server on port %s", conf.GrpcAdminPort)
	go func() {
		err = service.Serve(listener)
		if err != nil {
			log.Fatal("Admin server failed to serve.")
		}
	}()
}

func startNewGrpcImplantServer(work chan *pb.Command, output chan *pb.Command, conf config.Config, log logger.Logger) {
	service := grpc.NewServer()
	implantServer := grpcapi.NewImplantServer(work, output, conf, log)

	pb.RegisterImplantServer(service, implantServer)
	reflection.Register(service)

	addr := fmt.Sprintf("%s:%s", conf.GrpcImplantHost, conf.GrpcImplantPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Implant listener failed to bind to port %s: ", conf.GrpcImplantPort, err)
	}
	log.Infof("Starting gRPC implant server on port %s", conf.GrpcImplantPort)
	go func() {
		err = service.Serve(listener)
		if err != nil {
			log.Fatal("Implant server failed to serve.")
		}
	}()
}
