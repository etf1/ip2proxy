GO           := go
GINKGO       := ginkgo
GOMETALINTER := gometalinter.v2
GOVENDOR     := govendor

pkgs = $(shell $(GO) list ./... | grep -v /vendor/)

all: deps format vet lint test

deps:
	@echo ">> (re)installing deps"
	@$(GO) get -u github.com/onsi/ginkgo/ginkgo
	@$(GO) get -u github.com/onsi/gomega
	@$(GO) get -u github.com/kardianos/govendor
	@$(GO) get -u gopkg.in/alecthomas/gometalinter.v2
	@$(GOMETALINTER) --install
	@$(GOVENDOR) install

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

lint:
	@echo ">> checking code style"
	@$(GOMETALINTER) --config=./.gometalinter.json .

test:
	@echo ">> running tests"
	@$(GO) version
	@$(GINKGO) version
	@$(GINKGO) -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2 .
	@$(GO) tool cover -html  ip2proxy.coverprofile -o cover.html


.PHONY: all deps format style vet test
