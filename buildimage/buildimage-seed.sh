#!/bin/bash
SEED_DIR=latest-test
BUILD_DIR=build-seed

sudo apt-get update && apt-get install -y expect

echo -e "1 create build dir and copy latest version seed to the dir"
if [ -d $BUILD_DIR ];then
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR
else
mkdir -p $BUILD_DIR
fi

cp -r $SEED_DIR/* $BUILD_DIR
cp Dockerfile $BUILD_DIR
cp startup.sh $BUILD_DIR
cp init.sh $BUILD_DIR
cp config.sh $BUILD_DIR
cp setdataforseed.sh $BUILD_DIR

echo -e "2 set password, data and clients"

cd $BUILD_DIR
./init.sh
./config.sh
VERSION=`cat version.txt`

echo -e "3 remove exist docker container"
docker stop jadepool-seed
docker rm jadepool-seed
docker rmi jadepool-seed:$VERSION

echo -e "4 build docker image"
docker build --force-rm --no-cache --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') --build-arg VERSION=$VERSION -t jadepool-seed:$VERSION -f ./Dockerfile . 
#docker build --force-rm --no-cache --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') --build-arg VERSION=$VERSION -t jadepool-seed:$VERSION --env PORT=8888 -f ./Dockerfile . 

#echo -e "5 start seed container"
#docker run -d --name jadepool-seed -p 8899:8899 -v /data/seed-data:/usr/app jadepool-seed:R1.2.0.181212
#docker run -d --name jadepool-seed4db -p 8888:8888 -v /data/seed-data:/usr/app jadepool-seed:R1.2.0.181212

echo -e "----------------------------------"
echo -e "build seed image success"
echo -e "----------------------------------"
