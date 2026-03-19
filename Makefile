.PHONY: build build-with-ui build-without-ui test lint run-cp run-proxy-goal run-proxy-cap dev clean

# ── Build ──

build: build-without-ui

build-without-ui:
	cd control-plane && go build -tags no_ui -o ../bin/capafabric ./cmd/capafabric
	cd proxy && go build -o ../bin/cfproxy ./cmd/cfproxy

build-with-ui:
	cd control-plane/ui && npm install && npm run build
	cd control-plane && go build -o ../bin/capafabric ./cmd/capafabric
	cd proxy && go build -o ../bin/cfproxy ./cmd/cfproxy

# ── Test ──
test:
	cd shared && go test ./...
	cd control-plane && go test ./...
	cd proxy && go test ./...

# ── Lint ──
lint:
	cd shared && go vet ./...
	cd control-plane && go vet ./...
	cd proxy && go vet ./...

# ── Run Components ──

run-cp:
	cd control-plane && go run ./cmd/capafabric --config=configs/capafabric.dev.yaml

run-proxy-goal:
	cd proxy && go run ./cmd/cfproxy --mode=agent --config=configs/proxy-agent.yaml

run-proxy-cap:
	cd proxy && go run ./cmd/cfproxy --mode=capability --config=configs/proxy-capability.yaml \
		--manifest=$(MANIFEST)

run-litellm:
	litellm --config config/litellm_config.dev.yaml --port 4000

# ── Validate manifests ──
validate:
	@find examples -name "manifest.yaml" | while read f; do \
		echo "Validating: $$f"; \
	done

# ── Clean ──
clean:
	rm -rf bin/
	rm -f control-plane/capafabric.exe control-plane/capafabric
	rm -f proxy/cfproxy.exe proxy/cfproxy

# ── Docker ──
docker-build:
	docker build -f control-plane/Dockerfile.slim -t capafabric/control-plane:latest-slim ./control-plane
	docker build -t capafabric/proxy:latest ./proxy

up:
	docker compose up -d

down:
	docker compose down
