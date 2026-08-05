package grpc_clients

import (
	"os"

	pb "github.com/trunglq04/goride/shared/proto/trip"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type tripServiceClient struct {
	pb.UnimplementedTripServiceServer
	Client pb.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*tripServiceClient, error) {
	tripServiceURL := os.Getenv("TRIP_SERVICE_URL") // container name setup in infra
	if tripServiceURL == "" {
		tripServiceURL = "trip-service:9093"
	}
	conn, err := grpc.NewClient(tripServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewTripServiceClient(conn)

	return &tripServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *tripServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
