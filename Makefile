VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X github.com/paulrose/hatch/cmd.version=$(VERSION) \
	-X github.com/paulrose/hatch/cmd.commit=$(COMMIT) \
	-X github.com/paulrose/hatch/cmd.date=$(DATE)

.PHONY: build build-test run test clean app frontend

build:
	go build -ldflags '$(LDFLAGS)' -o hatch .

build-test:
	go build -ldflags '$(LDFLAGS)' -o testing/hatch .

run:
	go run -ldflags '$(LDFLAGS)' . $(ARGS)

test:
	go test ./...

frontend:
	cd frontend && npm install && npm run build

app: frontend
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop,production -ldflags '$(LDFLAGS)' -o hatch .

clean:
	rm -f hatch
