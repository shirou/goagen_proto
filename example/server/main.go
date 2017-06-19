package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/shirou/goagen_proto/example"
)

type userService struct {
	users []*pb.User
	m     sync.Mutex
}

func (us *userService) Get(c context.Context, p *pb.UserGetType) (*pb.User, error) {
	return us.users[0], nil
}

func (us *userService) List(p *pb.Empty, stream pb.UserService_ListServer) error {
	us.m.Lock()
	defer us.m.Unlock()
	for _, p := range us.users {
		if err := stream.Send(p); err != nil {
			return err
		}
	}
	return nil
}

func (us *userService) Create(c context.Context, p *pb.UserCreateType) (*pb.Empty, error) {
	us.m.Lock()
	defer us.m.Unlock()
	u := pb.User{
		Name: p.Name,
	}
	fmt.Println(p)
	us.users = append(us.users, &u)
	return new(pb.Empty), nil
}

func main() {
	lis, err := net.Listen("tcp", ":11111")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()

	pb.RegisterUserServiceServer(server, new(userService))
	server.Serve(lis)
}
