#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
VERSION=$1

SRC=${DIR}/../build/abs-src
OUT=${DIR}/../build/client-dist
rm -rf ${SRC} ${OUT}

git clone --depth 1 --branch v${VERSION} https://github.com/advplyr/audiobookshelf.git ${SRC}

cd ${SRC}/client
npm ci
npm run generate

mkdir -p ${OUT}
cp -r dist/. ${OUT}/
ls ${OUT}
