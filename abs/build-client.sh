#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )

SRC=${DIR}/../build/abs-src
OUT=${DIR}/../build/client-dist
rm -rf ${OUT}

cd ${SRC}/client
npm ci
npm run generate

mkdir -p ${OUT}
cp -r dist/. ${OUT}/
ls ${OUT}
