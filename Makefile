.PHONY: genproto

ROOT_DIR = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
GO_PROTO_DIR = $(ROOT_DIR)pb

genproto:
	cd proto; protoc --go_out=$(GO_PROTO_DIR) *.proto
