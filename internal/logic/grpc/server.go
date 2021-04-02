package grpc

import (
	"context"
	pb "gameserver/api/logic"
	"gameserver/internal/logic/conf"
	"gameserver/internal/logic/logic"
	"gameserver/pkg/common"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type server struct {
	srv *logic.Logic
}

var _ pb.LogicServer = &server{}

// New logic grpc server
func New(c *conf.RPCServer, l *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	pb.RegisterLogicServer(srv, &server{l})
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	common.Println("grpc server start:", c.Network, c.Addr)
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

// Receive receive a message.
func (s *server) Receive(ctx context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	if err := s.srv.Receive(ctx, req.UserID, req.Proto); err != nil {
		return &pb.ReceiveReply{}, err
	}
	return &pb.ReceiveReply{}, nil
}
