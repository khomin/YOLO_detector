cmake_minimum_required(VERSION 3.5)

set(PROJECT opencv_cpp)
set(BUILD_CONFIG_CXX_TESTS OFF)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

project(${PROJECT} LANGUAGES CXX)


add_subdirectory(${CMAKE_SOURCE_DIR}/external/config-cxx)

add_executable(${PROJECT}
    main.cpp
    detector.h detector.cpp
    network_client.h network_client.cpp
)

target_include_directories(${PROJECT} PUBLIC
    ${CMAKE_SOURCE_DIR}/../cpp/
    ${CMAKE_SOURCE_DIR}/../.lib_pack/opencv/include/opencv4
    ${CMAKE_SOURCE_DIR}/../.lib_pack/protobuf/include
)

file(GLOB OPENCV_LIBRARIES
    ${CMAKE_SOURCE_DIR}/../.lib_pack/opencv/lib/*.so
    ${CMAKE_SOURCE_DIR}/../.lib_pack/protobuf/lib/libprotobuf.a
)

target_link_libraries(${PROJECT} PUBLIC
    ${OPENCV_LIBRARIES}
    config-cxx
)
