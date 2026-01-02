git submodule update --init


# build protobuf

./scripts/gen-proto.sh 

# trying to get docker to build opencv for arm

docker pull ubuntu:20.04

docker run -it --name opencv_for_arm ubuntu:20.04 /bin/bash

apt update
apt install -y crossbuild-essential-arm64 cmake git pkg-config
apt install golang-go

# start

docker start -ai 2c0e4f042af5

git clone https://github.com/opencv/opencv.git
git clone https://github.com/opencv/opencv_contrib.git


ok i built opencv
now i need grpc for arm too

mkdir .lib_pack

docker cp ~/Documents/PROJECTS/YOLO_detector 2c0e4f042af5:/root/YOLO_detector
docker cp ~/Documents/PROJECTS/YOLO_detector/.lib_pack/grpc 2c0e4f042af5:/root/.lib_pack/grpc
docker cp ~/Documents/PROJECTS/YOLO_detector/.lib_pack/opencv 2c0e4f042af5:/root/.lib_pack/opencv

# to build cpp

export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/khomin/Documents/PROJECTS/YOLO_detector/.lib_pack/build_grpc/x86/lib/
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/khomin/Documents/PROJECTS/YOLO_detector/.lib_pack/opencv/lib/


# need create config

mkdir config
cp ./cpp/config.json ./config/

cp -r ./resources/ ~/Documents/PROJECTS/YOLO_detector/cpp/build_arm
