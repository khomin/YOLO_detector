#!/bin/bash
SCRIPT_PATH=$(dirname $(readlink -f $0))
cd $SCRIPT_PATH

echo "Using protobuf compiler:" `which protoc`

export PATH=$PATH:../.lib_pack/build_grpc/x86/bin/
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/khomin/Documents/PROJECTS/YOLO_detector/.lib_pack/build_grpc/x86/lib/

mkdir -p ../cpp/protobuf/generated
set -x

protoc -I=../protobuf \
--cpp_out=../cpp/protobuf/generated \
--grpc_out=../cpp/protobuf/generated \
--plugin=protoc-gen-grpc=$(which grpc_cpp_plugin) \
--proto_path=../protobuf \
tracker.proto