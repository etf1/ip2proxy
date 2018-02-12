GO           := go
GINKGO       := ginkgo
GOMETALINTER := gometalinter.v2
GOVENDOR     := dep

GOTOOLS       = github.com/onsi/ginkgo/ginkgo  \
				github.com/onsi/gomega         \
				github.com/golang/dep/cmd/dep  \
				gopkg.in/alecthomas/gometalinter.v2

pkgs = $(shell $(GO) list ./... | grep -v /vendor/)

all: deps format vet lint test

tools:
	@echo ">> ensuring tools are installed"
	@- $(foreach GOTOOL,$(GOTOOLS),          \
		$(GO) get $(GOTOOL) ;                \
	)

tools-update:
	@echo ">> updating tools"
	@- $(foreach GOTOOL,$(GOTOOLS),          \
		$(GO) get -u $(GOTOOL) ;             \
	)

deps: tools
	@echo ">> (re)installing deps"
	@$(GOVENDOR) ensure

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

lint:
	@echo ">> checking code style"
	@$(GOMETALINTER) --config=./.gometalinter.json $(pkgs)

test:
	@echo ">> running tests"
	@$(GO) version
	@$(GINKGO) version
	@$(GINKGO) -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2 .
	@$(GO) tool cover -html  ip2proxy.coverprofile -o cover.html


.PHONY: all
