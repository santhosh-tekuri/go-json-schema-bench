GOLANGCI_LINT_VERSION := "v1.31.0"
UNIT_TEST_COUNT := 1

# The head of Makefile determines location of dev-go to include standard targets.
GO ?= go
export GO111MODULE = on

ifneq "$(GOFLAGS)" ""
  $(info GOFLAGS: ${GOFLAGS})
endif

ifneq "$(wildcard ./vendor )" ""
  $(info Using vendor)
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

run-test:
	@make test-unit > test.txt || echo "Some cases failed"

## Run tests
test: run-test
	@cat test.txt
	@echo "santhosh-tekuri/jsonschema failed tests count"
	@echo "draft 7          | $(shell cat test.txt | grep 'FAIL: TestSanthoshDraft7/' | grep -v 'format.json' | wc -l)"
	@echo "draft 7 format   | $(shell cat test.txt | grep 'FAIL: TestSanthoshDraft7/format.json' | wc -l)"
	@echo "draft 7 optional | $(shell cat test.txt | grep 'FAIL: TestSanthoshDraft7Opt/' | wc -l)"
	@echo "ajv              | $(shell cat test.txt | grep 'FAIL: TestSanthoshAjv/' | wc -l)"
	@echo "qri-io/jsonschema failed tests count"
	@echo "draft 7          | $(shell cat test.txt | grep 'FAIL: TestQriDraft7/' | grep -v 'format.json' | wc -l)"
	@echo "draft 7 format   | $(shell cat test.txt | grep 'FAIL: TestQriDraft7/format.json' | wc -l)"
	@echo "draft 7 optional | $(shell cat test.txt | grep 'FAIL: TestQriDraft7Opt/' | wc -l)"
	@echo "ajv              | $(shell cat test.txt | grep 'FAIL: TestQriAjv/' | wc -l)"
	@echo "xeipuuv/gojsonschema failed tests count"
	@echo "draft 7          | $(shell cat test.txt | grep 'FAIL: TestXeipuuvDraft7/' | grep -v 'format.json' | wc -l)"
	@echo "draft 7 format   | $(shell cat test.txt | grep 'FAIL: TestXeipuuvDraft7/format.json' | wc -l)"
	@echo "draft 7 optional | $(shell cat test.txt | grep 'FAIL: TestXeipuuvDraft7Opt/' | wc -l)"
	@echo "ajv              | $(shell cat test.txt | grep 'FAIL: TestXeipuuvAjv/' | wc -l)"

## Ensure external test suite
deps:
	@git submodule init && git submodule update

BENCH_COUNT ?= 10

benchstat:
	@test -s $(GOPATH)/bin/benchstat || GO111MODULE=off GOFLAGS= GOBIN=$(GOPATH)/bin $(GO) get -u golang.org/x/perf/cmd/benchstat

## Benchmark performance suite from ajv
bench-ajv: benchstat
	@VALIDATOR=santhosh $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkAjv$$ -run ^$$ >santhosh-ajv.txt || echo "Some cases failed"
	@VALIDATOR=qri $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkAjv$$ -run ^$$ >qri-ajv.txt || echo "Some cases failed"
	@VALIDATOR=xeipuuv $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkAjv$$ -run ^$$ >xeipuuv-ajv.txt || echo "Some cases failed"
	@echo "Results"
	@benchstat -geomean santhosh-ajv.txt qri-ajv.txt xeipuuv-ajv.txt

## Benchmark draft-07 test cases
bench-draft7: benchstat
	@VALIDATOR=santhosh $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkDraft7$$ -run ^$$ >santhosh-draft7.txt || echo "Some cases failed"
	@VALIDATOR=qri $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkDraft7$$ -run ^$$ | tee >qri-draft7.txt || echo "Some cases failed"
	@VALIDATOR=xeipuuv $(GO) test . -count $(BENCH_COUNT) -bench ^BenchmarkDraft7$$ -run ^$$ >xeipuuv-draft7.txt || echo "Some cases failed"
	@echo "Results"
	@benchstat -geomean santhosh-draft7.txt qri-draft7.txt xeipuuv-draft7.txt
