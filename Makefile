.PHONY: all

all:
	@echo "**********************"
	@echo "** bondb build tool **"
	@echo "**********************"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  test        - standard go test"
	@echo "  build       - build the package"
	@echo "  install     - install the package"
	@echo ""
	@echo "  tools       - go get's a bunch of tools for dev"
	@echo "  deps        - pull and setup dependencies"
	@echo "  update_deps - update deps lock file"

test:
	@go test -v ./...

retest:
	@make test; reflex -r "^*\.go$$" -- make test

build:
	@go build ./... 

install:
	@go install ./...

tools:
	@go get github.com/robfig/glock
	@go get github.com/cespare/reflex

deps:
	@glock sync -n github.com/pressly/bondb < Glockfile

update_deps:
	@glock save -n github.com/pressly/bondb > Glockfile
