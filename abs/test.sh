#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )

export SNAP=${DIR}/../build/snap
export SNAP_DATA=${DIR}/../build/snap-data
NODE=${SNAP}/node/node.sh
APP=${SNAP}/abs

${NODE} --version

. ${DIR}/../config/abs.env
test -f "${NUSQLITE3_PATH}"
echo "nusqlite3 present: ${NUSQLITE3_PATH}"

cd ${APP}
${NODE} -e "console.log('node runtime ok')"
