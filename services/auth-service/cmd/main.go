package main

import (
	"context"

	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
)

var grpcAddr = env.GetString("GRPC_ADDR", ":9094")

func main() {
	logger.Setup("auth-service")

	ctx := context.Background()

	app, shutdown, err := Setup(ctx)
	if err != nil {
		logger.Fatal("auth-service exited with error", "err", err)
	}
	defer shutdown()

	if err := app.Run(); err != nil {
		logger.Fatal("auth-service exited with error", "err", err)
	}
}
