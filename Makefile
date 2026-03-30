# Variables
SERVER_NAME=server
CLIENT_NAME=client
DAYS_VALID=365

.PHONY: all cert server-cert client-cert clean help

all: cert

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
