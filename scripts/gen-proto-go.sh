#!/bin/bash

SCRIPT_PATH=$(dirname $(readlink -f $0))
cd $SCRIPT_PATH

echo "Using protobuf compiler:" `which protoc`

PATH=$PATH:../.lib_pack/build_grpc/x86/bin/
PATH=$PATH:~/go/bin/

mkdir -p ../go_service/grpc/generated
set -x

protoc \
--go_out=../go_service/grpc/generated  \
--go-grpc_out=../go_service/grpc/generated \
--proto_path=../protobuf \
tracker.proto