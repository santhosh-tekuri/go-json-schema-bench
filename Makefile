GOLANGCI_LINT_VERSION := "v1.31.0"
UNIT_TEST_COUNT := 1

# The head of Makefile determines location of dev-go to include standard targets.
GO ?= go
export GO111MODULE = on

ifneq "$(GOFLAGS)" ""
  $(info GOFLAGS: ${GOFLAGS})
endif

ifneq "$(wildcard ./vendor )" ""
  $(info >> using vendor)
  modVendor =  -mod=vendor
  ifeq (,$(findstring -mod,$(GOFLAGS)))
      export GOFLAGS := ${GOFLAGS} ${modVendor}
  endif
  ifneq "$(wildcard ./vendor/github.com/bool64/dev)" ""
  	DEVGO_PATH := ./vendor/github.com/bool64/dev
  endif
endif

ifeq ($(DEVGO_PATH),)
	DEVGO_PATH := $(shell GO111MODULE=on $(GO) list ${modVendor} -f '{{.Dir}}' -m github.com/bool64/dev)
	ifeq ($(DEVGO_PATH),)
    	$(info Module github.com/bool64/dev not found, downloading.)
    	DEVGO_PATH := $(shell export GO111MODULE=on && $(GO) mod tidy && $(GO) list -f '{{.Dir}}' -m github.com/bool64/dev)
	endif
endif

-include $(DEVGO_PATH)/makefiles/main.mk
-include $(DEVGO_PATH)/makefiles/test-unit.mk
-include $(DEVGO_PATH)/makefiles/lint.mk
-include $(DEVGO_PATH)/makefiles/github-actions.mk

## Run tests
test: test-unit

## Ensure external test suite
deps:
	@git submodule init && git submodule update

BENCH_COUNT ?= 10

benchstat:
	@test -s $(GOPATH)/bin/benchstat || GO111MODULE=off GOFLAGS= GOBIN=$(GOPATH)/bin $(GO) get -u golang.org/x/perf/cmd/benchstat

## Benchmark performance suite from ajv
bench-ajv: benchstat
	@VALIDATOR=santhosh $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkAjv$$ -run ^$$ >santhosh-ajv.txt
	@VALIDATOR=qri $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkAjv$$ -run ^$$ >qri-ajv.txt
	@VALIDATOR=xeipuuv $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkAjv$$ -run ^$$ >xeipuuv-ajv.txt
	@echo "Results"
	@benchstat santhosh-ajv.txt qri-ajv.txt xeipuuv-ajv.txt

## Benchmark draft-07 test cases
bench-draft7: benchstat
	@VALIDATOR=santhosh $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkDraft7$$ -run ^$$ >santhosh-draft7.txt
	@VALIDATOR=qri $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkDraft7$$ -run ^$$ | tee >qri-draft7.txt
	@VALIDATOR=xeipuuv $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkDraft7$$ -run ^$$ >xeipuuv-draft7.txt
	@echo "Results"
	@benchstat santhosh-draft7.txt qri-draft7.txt xeipuuv-draft7.txt
