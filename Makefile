# Variables
BIN_DIR=bin
SERVER_NAME=server
ADMIN_NAME=admin
IMPLANT_NAME=implant

CERT_SERVER=server
CERT_CLIENT=client

DAYS_VALID=365
PROTO_DIR=grpcapi
PROTO_FILE=$(PROTO_DIR)/implant.proto
PB_OUT=grpcapi

.PHONY: all build proto cert server-cert client-cert clean help server admin implant

## all: Compiles proto, generates certs, and builds all binaries
all: proto cert build

## build: Builds all project binaries (Server, Admin, Implant)
build: server admin implant

## server: Builds the C2 server binary
server:
	@echo "Building Server..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(SERVER_NAME) ./cmd/server

## admin: Builds the Admin CLI tool binary
admin:
	@echo "Building Admin Tool..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(ADMIN_NAME) ./cmd/client

## implant: Builds the Implant binary
implant:
	@echo "Building Implant..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(IMPLANT_NAME) ./cmd/implant

# Generic function to generate a self-signed certificate
# Usage: $(call gen_cert,filename_base,common_name,extra_args)
define gen_cert
	@echo "Generating certificate for $(2)..."
	openssl req -x509 -newkey rsa:4096 -keyout $(1).key -out $(1).crt \
		-days $(DAYS_VALID) -nodes -subj "/CN=$(2)" $(3)
	@echo "Done: $(1).crt, $(1).key"
endef

## cert: Generates both server and client certificates
cert: server-cert client-cert

## server-cert: Generates the self-signed server certificate with SAN
server-cert:
	$(call gen_cert,$(SERVER_NAME),localhost,-addext subjectAltName=DNS:localhost, IP:127.0.0.1)

## client-cert: Generates the self-signed client certificate
client-cert:
	$(call gen_cert,$(CLIENT_NAME),grpc-client,)

## proto: Compiles the protobuf files for Go and gRPC
proto:
	@echo "Compiling protobuf..."
	mkdir -p $(PB_OUT)
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PB_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PB_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)
	@echo "Done: Protobuf compiled to $(PB_OUT)"

## clean: Removes all generated certificate and key files
clean:
	@echo "Removing certificates..."
	rm -f *.crt *.key *.srl

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' -e 's/: /	- /g'
