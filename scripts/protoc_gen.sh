#!/bin/bash
protoc --proto_path=api/protobuf-spec --proto_path=third_party \
	--go_out=plugins=grpc:pkg/grpc/proto auth.proto
