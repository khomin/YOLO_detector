cmake_minimum_required(VERSION 3.5)

set(PROJECT opencv_cpp)

project(${PROJECT} LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(${PROJECT}
    main.cpp
)

target_include_directories(${PROJECT} PUBLIC
    ${CMAKE_SOURCE_DIR}/../.lib_pack/opencv/include/opencv4
    ${CMAKE_SOURCE_DIR}/../cpp/
)

file(GLOB OPENCV_LIBRARIES
    ${CMAKE_SOURCE_DIR}/../.lib_pack/opencv/lib/*.so
)

target_link_libraries(${PROJECT} PUBLIC
    ${OPENCV_LIBRARIES}
)
