#!/bin/bash
set -e

# Get the directory where THIS script lives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Define the Project Root relative to the script (one level up)
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Define your paths relative to the Project Root
SOURCE_DIR="$PROJECT_ROOT/.lib_pack/opencv_src"
CONTRIB_DIR="$PROJECT_ROOT/.lib_pack/opencv_contrib"
BUILD_DIR="$PROJECT_ROOT/.lib_pack/opencv_build"
INSTALL_DIR="$PROJECT_ROOT/.lib_pack/opencv"
ARCH=${1:-x86}

echo "--- Building OpenCV in: $PROJECT_ROOT ---"

# 2. Clone Main and Contrib if they don't exist
if [ ! -d "$SOURCE_DIR" ]; then
    git clone --depth 1 https://github.com/opencv/opencv.git "$SOURCE_DIR"
fi

if [ ! -d "$CONTRIB_DIR" ]; then
    git clone --depth 1 https://github.com/opencv/opencv_contrib.git "$CONTRIB_DIR"
fi

echo "CONTRIB_DIR=$CONTRIB_DIR"

# 3. Clean and Setup Build Directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Determine if we need the toolchain
CMAKE_FLAGS=(
    "-DCMAKE_INSTALL_PREFIX=$INSTALL_DIR"
    "-DOPENCV_EXTRA_MODULES_PATH=$CONTRIB_DIR/modules"
    "-DBUILD_SHARED_LIBS=ON"
    "-DCMAKE_POSITION_INDEPENDENT_CODE=ON"
    "-DWITH_TBB=ON"
    "-DBUILD_TBB=ON"
    "-DBUILD_SHARED_LIBS=ON"
    "-DBUILD_TESTS=OFF"
    "-DBUILD_PERF_TESTS=OFF"
    "-DOPENCV_ENABLE_NONFREE=ON"
    "-DOPENCV_GENERATE_PKGCONFIG=ON"
    "-DBUILD_opencv_python3=ON"
    "-DINSTALL_PYTHON_EXAMPLES=OFF"
    "-DCMAKE_BUILD_TYPE=RELEASE"
    "-DBUILD_opencv_xphoto=OFF"
)

if [ "$ARCH" == "arm64" ]; then
    echo "--- Building for ARM64 (Cross-compiling) ---"
    CMAKE_FLAGS+=("-DCMAKE_TOOLCHAIN_FILE=$SCRIPT_DIR/toolchain-arm64.cmake")
else
    echo "--- Building for x86 (Native) ---"
fi

# Run CMake with the expanded array of flags
cmake "${CMAKE_FLAGS[@]}" "$SOURCE_DIR"

make -j$(nproc)
make install