PROTO_DIR := proto
PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT := .

# PostgreSQL connection string for migrations
POSTGRES_URI ?= postgres://postgres:your-secure-password@localhost:5432/goride?sslmode=disable
MIGRATIONS_DIR := services/auth-service/migrations

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_SRC)

.PHONY: proto-auth
proto-auth:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=shared/proto/auth \
		--go_opt=paths=source_relative \
		--go-grpc_out=shared/proto/auth \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/auth.proto

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(POSTGRES_URI)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(POSTGRES_URI)" down

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

.PHONY: gen-rsa-keys
gen-rsa-keys:
	@mkdir -p infra/development/keys
	openssl genpkey -algorithm RSA -out infra/development/keys/jwt_private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -in infra/development/keys/jwt_private.pem -pubout -out infra/development/keys/jwt_public.pem
	@echo "RSA keys generated in infra/development/keys/"

.PHONY: stripe-listen
stripe-listen:
	stripe listen --forward-to localhost:8081/webhook/stripe

.PHONY: build-push-azure
build-push-azure:
	infra/production/azure/build-push-images.sh 