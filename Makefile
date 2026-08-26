PROTO_DIR := proto
PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT := .

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_SRC)

.PHONY: stripe-listen
stripe-listen:
	stripe listen --forward-to localhost:8081/webhook/stripe

.PHONY: build-push-azure
build-push-azure:
	infra/production/azure/build-push-images.sh 