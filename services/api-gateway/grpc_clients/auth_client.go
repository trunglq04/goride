package grpc_clients

import (
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/metrics"
	pb "github.com/trunglq04/goride/shared/proto/auth"
	"github.com/trunglq04/goride/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type authServiceClient struct {
	Client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthServiceClient() (*authServiceClient, error) {
	authServiceURL := env.GetString("AUTH_SERVICE_URL", "auth-service:9094")

	var dialOptions []grpc.DialOption = append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		metrics.DialOptionWithMetrics(),
	)
	conn, err := grpc.NewClient(authServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)

	return &authServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *authServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
