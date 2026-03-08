# Variables
CERT_OUT=server.crt
KEY_OUT=server.key
COMMON_NAME=localhost
DAYS_VALID=365

.PHONY: all cert clean help

all: cert

## cert: Generates a self-signed TLS certificate and private key
cert:
	@echo "Generating self-signed certificate with SAN..."
	openssl req -x509 -newkey rsa:4096 -keyout $(KEY_OUT) -out $(CERT_OUT) \
		-days $(DAYS_VALID) -nodes -subj "/CN=$(COMMON_NAME)" \
		-addext "subjectAltName = DNS:localhost, IP:127.0.0.1"
	@echo "Done. Files created: $(CERT_OUT), $(KEY_OUT)"

## clean: Removes generated certificate and key files
clean:
	@echo "Removing certificates..."
	rm -f $(CERT_OUT) $(KEY_OUT)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' -e 's/: /	- /g'
