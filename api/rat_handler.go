package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vpaklatzis/conduit-go/models"
	"github.com/vpaklatzis/conduit-go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	cr   models.CommandRequest
	conn *grpc.ClientConn
	err  error
)

func (s *Server) RatHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&cr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cr.Command == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide a valid command"})
		return
	}
	requestUri := fmt.Sprintf("%s:%s", s.conf.GrpcAdminHost, s.conf.GrpcAdminPort)
	if conn, err = grpc.Dial(requestUri, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		s.log.Fatal("Admin client could not connect to the server: ", err)
	}
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			s.log.Fatal("Error occurred while closing the connection: ", err)
		}
	}(conn)

	client := pb.NewAdminClient(conn)

	cmd := new(pb.Command)
	cmd.In = cr.Command

	cmd, err = client.RunCommand(c, cmd)
	if err != nil {
		s.log.Fatal("Error occurred while trying to run the command: ", err)
	}
	result := &models.CommandResponse{
		Result: cmd.Out,
	}
	r, err := json.Marshal(result)
	if err != nil {
		s.log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.Data(http.StatusOK, "application/json", r)
}
