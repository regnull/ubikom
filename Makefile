.PHONY: genproto
.PHONY: build
.PHONY: upload
.PHONY: server-stop
.PHONY: mail-server-stop
.PHONY: server-start
.PHONY: mail-server-start
.PHONY: test

UBIKOM_ONE_ADR = alpha.ubikom.cc
MAIL_SERVER_ADR = mail.ubikom.cc
SSH_KEY = $(HOME)/aws/ubikom-one.pem
MAIL_SSH_KEY = $(HOME)/aws/ubikom-two.pem

ROOT_DIR = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
GO_PROTO_DIR = $(ROOT_DIR)pb

genproto:
	cd proto; protoc --go_out=$(GO_PROTO_DIR) *.proto
	cd proto; protoc --go-grpc_out=$(GO_PROTO_DIR) *.proto

server-stop:
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-server
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-dump
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-proxy
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl stop ubikom-web

mail-server-stop:
	ssh -i $(MAIL_SSH_KEY) ubuntu@$(MAIL_SERVER_ADR) sudo supervisorctl stop ubikom-gateway

upload:
	scp -i $(SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-server ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-server
	scp -i $(SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-dump ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-dump
	scp -i $(SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-cli ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-cli
	scp -i $(SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-proxy ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-proxy
	scp -i $(SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-web ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom-web
	scp -i $(SSH_KEY) $(ROOT_DIR)config/ubikom-server.conf ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/ubikom.conf
	scp -i $(SSH_KEY) $(ROOT_DIR)config/supervisor/* ubuntu@$(UBIKOM_ONE_ADR):~/ubikom/supervisor

mail-upload:
	scp -i $(MAIL_SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-gateway ubuntu@$(MAIL_SERVER_ADR):~/ubikom/ubikom-gateway
	scp -i $(MAIL_SSH_KEY) $(ROOT_DIR)build/linux-amd64/ubikom-receiver ubuntu@$(MAIL_SERVER_ADR):~/ubikom/ubikom-receiver

server-start:
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-server
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-dump
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-proxy
	ssh -i $(SSH_KEY) ubuntu@$(UBIKOM_ONE_ADR) sudo supervisorctl start ubikom-web

mail-server-start:
	ssh -i $(MAIL_SSH_KEY) ubuntu@$(MAIL_SERVER_ADR) sudo supervisorctl start ubikom-gateway

build:
	$(ROOT_DIR)scripts/build.sh

test:
	go test -cover -timeout=30s ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
