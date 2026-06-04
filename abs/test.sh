#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SNAP=${DIR}/../build/snap
NODE=${SNAP}/node/node.sh
APP=${SNAP}/abs

${NODE} --version

source ${SNAP}/node/runtime-env.sh
test -f "${NUSQLITE3_PATH}"
echo "nusqlite3 present: ${NUSQLITE3_PATH}"

cd ${APP}
${NODE} -e "console.log('node runtime ok')"
