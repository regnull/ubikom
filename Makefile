.PHONY: genproto
.PHONY: build
.PHONY: upload

UBIKOM_ONE_ADR = ec2-18-191-204-119.us-east-2.compute.amazonaws.com
SSH_KEY = $(HOME)/aws/ubikom-one.pem

ROOT_DIR = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
GO_PROTO_DIR = $(ROOT_DIR)pb

genproto:
	cd proto; protoc --go_out=$(GO_PROTO_DIR) *.proto
	cd proto; protoc --go-grpc_out=$(GO_PROTO_DIR) *.proto

upload:
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-server
	scp -i $(SSH_KEY) $(ROOT_DIR)build/ubikom-server-linux ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-server
	scp -i $(SSH_KEY) $(ROOT_DIR)build/ubikom-dump-linux ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-dump
	scp -i $(SSH_KEY) $(ROOT_DIR)build/ubikom-cli-linux ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-cli
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-server

build:
	$(ROOT_DIR)scripts/build.sh
