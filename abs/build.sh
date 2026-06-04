#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
cd ${DIR}

NODE_DIR=${DIR}/../build/snap/node
APP_OUT=${DIR}/../build/snap/abs
rm -rf ${APP_OUT} ${NODE_DIR}
mkdir -p ${APP_OUT} ${NODE_DIR}

cp -r /app/. ${APP_OUT}

cp -r /usr ${NODE_DIR}/usr
cp -r /lib ${NODE_DIR}/lib

cp ${DIR}/node.sh ${NODE_DIR}/node.sh
cp ${DIR}/ffmpeg.sh ${NODE_DIR}/ffmpeg.sh
cp ${DIR}/ffprobe.sh ${NODE_DIR}/ffprobe.sh

: "${NUSQLITE3_DIR:=$(dirname "$(find / -name 'libnusqlite3*.so' 2>/dev/null | head -1)")}"
: "${NUSQLITE3_PATH:=$(find / -name 'libnusqlite3*.so' 2>/dev/null | head -1)}"

cat > ${NODE_DIR}/runtime-env.sh <<EOF
export NUSQLITE3_DIR=\${SNAP}/node${NUSQLITE3_DIR}
export NUSQLITE3_PATH=\${SNAP}/node${NUSQLITE3_PATH}
export FFMPEG_PATH=\${SNAP}/node/ffmpeg.sh
export FFPROBE_PATH=\${SNAP}/node/ffprobe.sh
EOF

cat ${NODE_DIR}/runtime-env.sh

${NODE_DIR}/node.sh --version
du -sh ${APP_OUT} ${NODE_DIR}
