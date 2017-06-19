package main

import (
	"fmt"
	"io"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/shirou/goagen_proto/example"
)

func create(name string, age int) error {
	conn, err := grpc.Dial("127.0.0.1:11111", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewUserServiceClient(conn)

	user := &pb.UserCreateType{
		Name: name,
		Age:  uint32(age),
	}
	_, err = client.Create(context.Background(), user)
	return err
}

func list() error {
	conn, err := grpc.Dial("127.0.0.1:11111", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewUserServiceClient(conn)

	stream, err := client.List(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}
	for {
		user, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(user)
	}
	return nil
}

func main() { /*
		(&sc.Cmds{
			{
				Name: "list",
				Desc: "list: listing person",
				Run: func(c *sc.C, args []string) error {
					return list()
				},
			},
			{
				Name: "add",
				Desc: "add [name] [age]: add person",
				Run: func(c *sc.C, args []string) error {
					if len(args) != 2 {
						return sc.UsageError
					}
					name := args[0]
					age, err := strconv.Atoi(args[1])
					if err != nil {
						return err
					}
					return add(name, age)
				},
			},
		}).Run(&sc.C{})
	*/

	if err := create("hoge", 11); err != nil {
		log.Fatalf("create %s", err)
	}
	if err := list(); err != nil {
		log.Fatalf("list %s", err)
	}
}
