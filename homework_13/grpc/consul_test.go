package grpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"testing"
	"time"
)

type ConsulTestSuite struct {
	suite.Suite
	cli *api.Client
}

func (s *ConsulTestSuite) SetupSuite() {
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(s.T(), err)
	s.cli = client
}

func (s *ConsulTestSuite) TestClient() {
	t := s.T()
	catalog := s.cli.Catalog()
	services, _, err := catalog.Service("interactive", "", nil)
	service := services[0]

	// Use the service's Address and Port for gRPC connection
	grpcServerAddress := fmt.Sprintf("%s:%d", service.Address, service.ServicePort)
	log.Println(grpcServerAddress)
	cc, err := grpc.Dial(grpcServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	require.NoError(t, err)
	grpcClient := NewUserServiceClient(cc)
	resp, err := grpcClient.GetByID(context.Background(), &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}

func (s *ConsulTestSuite) TestServer() {
	t := s.T()
	err := s.cli.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      "service-231",
		Name:    "interactive",
		Port:    8090,
		Address: "127.0.0.1",
		Check: &api.AgentServiceCheck{
			TTL:     (time.Second * 30).String(),
			Timeout: time.Minute.String(),
		},
	})
	require.NoError(s.T(), err)
	go func() {
		checkid := "service:service-231"
		for range time.Tick(time.Second * 30) {
			err := s.cli.Agent().PassTTL(checkid, "")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()

	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)

	require.NoError(t, err)

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	server.Serve(l)
}

func TestConsul(t *testing.T) {
	suite.Run(t, new(ConsulTestSuite))
}
