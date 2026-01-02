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
    "-D CMAKE_INSTALL_PREFIX=$INSTALL_DIR"
    "-D OPENCV_EXTRA_MODULES_PATH=$CONTRIB_DIR/modules"
    "-D BUILD_SHARED_LIBS=ON"
    "-D CMAKE_POSITION_INDEPENDENT_CODE=ON"
    "-D WITH_TBB=ON"
    "-D BUILD_TBB=ON"
    "-D BUILD_SHARED_LIBS=ON"
    "-D BUILD_TESTS=OFF"
    "-D BUILD_PERF_TESTS=OFF"
    "-D OPENCV_ENABLE_NONFREE=ON"
    "-D OPENCV_GENERATE_PKGCONFIG=ON"
    "-D BUILD_opencv_python3=ON"
    "-D INSTALL_PYTHON_EXAMPLES=OFF"
    "-D CMAKE_BUILD_TYPE=RELEASE"
    "-D BUILD_opencv_xphoto=OFF"
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