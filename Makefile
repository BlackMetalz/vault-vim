.PHONY: build run local clean test dev-up dev-down dev

test:
	go test ./... -v

build: test
	go build -o vault-vim .

run:
	./vault-vim

local: build run

clean:
	rm -f vault-vim

# Start local Vault with test data
dev-up:
	docker compose up -d

# Stop local Vault
dev-down:
	docker compose down

# Build + run against local dev Vault
dev: build
	VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root-token ./vault-vim
