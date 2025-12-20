cmake_minimum_required(VERSION 3.5)

set(PROJECT yolo_opencv_cpp)
set(BUILD_CONFIG_CXX_TESTS OFF)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

project(${PROJECT} LANGUAGES CXX)

add_subdirectory(${CMAKE_SOURCE_DIR}/external/config-cxx)

add_executable(${PROJECT}
    main.cpp
    detector.h detector.cpp
    network_client.h network_client.cpp
    ./protobuf/generated/tracker.pb.h
    ./protobuf/generated/tracker.pb.cc
    ./protobuf/generated/tracker.grpc.pb.h
    ./protobuf/generated/tracker.grpc.pb.cc

)

target_include_directories(${PROJECT} PUBLIC
    ${CMAKE_SOURCE_DIR}/../cpp/
    ${CMAKE_SOURCE_DIR}/../.lib_pack/opencv/include/opencv4
    ${CMAKE_SOURCE_DIR}/../.lib_pack/build_grpc/x86/include
)

file(GLOB OPENCV_LIBRARIES
    ${CMAKE_SOURCE_DIR}/../.lib_pack/opencv/lib/*.so
)

file(GLOB GRPC_AND_DEPS_LIBRARIES
    ${CMAKE_SOURCE_DIR}/../.lib_pack/build_grpc/x86/lib/*.so
)


target_link_libraries(${PROJECT} PUBLIC
    ${OPENCV_LIBRARIES}
    ${GRPC_AND_DEPS_LIBRARIES}
    config-cxx
)

target_precompile_headers(config-cxx PUBLIC "<algorithm>")
