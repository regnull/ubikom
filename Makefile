.PHONY: genproto
.PHONY: build
.PHONY: upload

UBIKOM_ONE_ADR = alpha.ubikom.cc
SSH_KEY = $(HOME)/aws/ubikom-one.pem

ROOT_DIR = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
GO_PROTO_DIR = $(ROOT_DIR)pb

genproto:
	cd proto; protoc --go_out=$(GO_PROTO_DIR) *.proto
	cd proto; protoc --go-grpc_out=$(GO_PROTO_DIR) *.proto

upload:
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-server
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-dump
	scp -i $(SSH_KEY) $(ROOT_DIR)build/ubikom-server-linux ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-server
	scp -i $(SSH_KEY) $(ROOT_DIR)build/ubikom-dump-linux ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-dump
	scp -i $(SSH_KEY) $(ROOT_DIR)build/ubikom-cli-linux ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-cli
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-server
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-dump

build:
	$(ROOT_DIR)scripts/build.sh
