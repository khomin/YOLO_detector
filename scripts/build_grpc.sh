#!/bin/bash

set -e

SRC_DIR="grpc"
REPO_URL="https://github.com/grpc/grpc"
LIB_PACK=".lib_pack"
INSTALL_DIR="$PWD/$LIB_PACK/build_grpc"
TEMP_BUILD_DIR="build_grpc"
ROOT_PATH=$PWD

if [ ! -d $LIB_PACK ]; then
    echo "--- making $LIB_PACK_NAME"
    mkdir $LIB_PACK
fi

cd $LIB_PACK

if [ -d $SRC_DIR ]; then
    echo "--- $REPO_DIR already cloned. Skipping clone. ---"
else
    echo "--- Cloning $REPO_DIR ---"
    git clone --recurse-submodules -b v1.76.0 --depth 1 --shallow-submodules $REPO_URL || exit 1
fi

cd $SRC_DIR

if [ ! -d $TEMP_BUILD_DIR ]; then
    echo "--- making $TEMP_BUILD_DIR $PWD"
    mkdir $TEMP_BUILD_DIR
fi
cd $TEMP_BUILD_DIR

# cmake ../ \
#     -DgRPC_INSTALL=ON \
#     -DgRPC_BUILD_TESTS=OFF \
#     -DCMAKE_CXX_STANDARD=17 \
#     -DABSL_ROOT_DIR=$INSTALL_DIR/x86 \
#     -DBUILD_SHARED_LIBS=OFF \
#     -DCMAKE_INSTALL_PREFIX=$INSTALL_DIR/x86 || exit 1

# make -j16

cmake --install . --prefix $INSTALL_DIR/x86