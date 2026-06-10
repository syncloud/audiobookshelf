#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
VERSION=$1

SRC=${DIR}/../build/abs-src
rm -rf ${SRC}

git clone --depth 1 --branch v${VERSION} https://github.com/advplyr/audiobookshelf.git ${SRC}

git -C ${SRC} apply ${DIR}/patches/chunked-upload.patch
grep -q "/api/upload/finalize" ${SRC}/client/pages/upload/index.vue
grep -q "handleUploadFinalize" ${SRC}/server/controllers/MiscController.js
grep -q "/upload/chunk" ${SRC}/server/routers/ApiRouter.js
