#!/bin/bash
SCRIPT_PATH=$(dirname $(readlink -f $0))
cd $SCRIPT_PATH

echo "Using protobuf compiler:" `which protoc`

mkdir -p ../cpp/protobuf/generated
set -x
protoc -I=../.lib_pack/protobuf/bin/protoc --cpp_out=../cpp/protobuf/generated --proto_path=../protobuf tracking_events.proto