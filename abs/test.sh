#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )

SNAP=${DIR}/../build/snap
NODE=${SNAP}/node/node.sh
APP=${SNAP}/abs

${NODE} --version

. ${SNAP}/node/runtime-env.sh
test -f "${NUSQLITE3_PATH}"
echo "nusqlite3 present: ${NUSQLITE3_PATH}"

cd ${APP}
${NODE} -e "console.log('node runtime ok')"
