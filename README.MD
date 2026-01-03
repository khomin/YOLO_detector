git submodule update --init


# build protobuf

./scripts/gen-proto.sh 

# trying to get docker to build opencv for arm

docker pull ubuntu:20.04

# run first time
docker run -it --name opencv_for_arm ubuntu:20.04 /bin/bash

# requirements
apt update
apt install -y crossbuild-essential-arm64 cmake git pkg-config
apt install golang-go
apt install wget

wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin

# start
docker start -ai 2c0e4f042af5

./scripts/build_grpc.sh arm64
./scripts/build_opencv.sh arm64

cd ./go_service/
go install

# may help
go get google.golang.org/genproto/googleapis/rpc@latest
go get google.golang.org/grpc@latest
go mod tidy

./scripts/build_go_arm.sh

# copy project to docker
docker cp ~/Documents/PROJECTS/YOLO_detector 2c0e4f042af5:/root/YOLO_detector

# copy go_service to docker
docker cp ~/Documents/PROJECTS/YOLO_detector/go_service 2c0e4f042af5:/root/YOLO_detector/

# to build cpp
# cmake was too old

wget https://github.com/Kitware/CMake/releases/download/v3.28.1/cmake-3.28.1-linux-x86_64.sh -O /tmp/cmake-install.sh
chmod +x /tmp/cmake-install.sh
/tmp/cmake-install.sh --skip-license --prefix=/usr/local --exclude-subdir
cmake --version

# flags to help linking
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/khomin/Documents/PROJECTS/YOLO_detector/.lib_pack/build_grpc/x86/lib/
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/khomin/Documents/PROJECTS/YOLO_detector/.lib_pack/opencv/lib/


# connect to the tablet
ssh teamhd@192.168.1.15
123

# copy from docker to host
mkdir ~/Documents/PROJECTS/Yolo_resources/bin/cpp/
mkdir ~/Documents/PROJECTS/Yolo_resources/bin/go_service/

docker cp 2c0e4f042af5:/root/YOLO_detector/.lib_pack ~/Documents/PROJECTS/Yolo_resources/bin/.lib_pack
docker cp 2c0e4f042af5:/root/YOLO_detector/cpp/build/yolo_opencv_cpp ~/Documents/PROJECTS/Yolo_resources/bin/cpp/
docker cp 2c0e4f042af5:/root/YOLO_detector/go_service/service ~/Documents/PROJECTS/Yolo_resources/bin/go_service/

# copy from host to tablet
scp -r ~/Documents/PROJECTS/Yolo_resources/bin/ teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector
scp ~/Documents/PROJECTS/YOLO_detector/cpp/config.json teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector/bin/cpp/config

# copy from docker to tablet
scp -r /root/YOLO_detector/cpp/build/yolo_opencv_cpp teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector/bin/cpp/
scp -r /root/YOLO_detector/go_service teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector/bin/

cd ~/Documents/yolo_detector/bin
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$PWD/.lib_pack/grpc/lib/
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$PWD/.lib_pack/opencv/lib/

./cpp/yolo_opencv_cpp

# copy resources
scp -r ~/Documents/PROJECTS/YOLO_detector/resources teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector/
scp ~/Documents/PROJECTS/YOLO_detector/go_service/config.conf teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector/bin/go_service/config.conf


# copy videos from tablet
scp -r teamhd@192.168.1.15:/home/teamhd/Documents/yolo_detector/record_temp ~/Desktop/