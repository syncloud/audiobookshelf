#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
cd ${DIR}

NODE_DIR=${DIR}/../build/snap/node
APP_OUT=${DIR}/../build/snap/abs
rm -rf ${APP_OUT} ${NODE_DIR}
mkdir -p ${APP_OUT} ${NODE_DIR}

cp -r /app/. ${APP_OUT}

OIDC_JS=${APP_OUT}/server/auth/OidcAuthStrategy.js
SETTINGS_JS=${APP_OUT}/server/objects/settings/ServerSettings.js
grep -q 'authOpenIDAdminGroups' ${OIDC_JS} || sed -i -f ${DIR}/oidc-group-mapping.sed ${OIDC_JS}
grep -q 'authOpenIDAdminGroups' ${SETTINGS_JS} || sed -i -f ${DIR}/oidc-group-mapping.sed ${SETTINGS_JS}
grep -q 'adminGroups.includes(group)' ${OIDC_JS}
grep -q 'rolesInOrderOfPriority.includes(defaultRole)' ${OIDC_JS}
grep -q 'user.permissions = Database.userModel.getDefaultPermissionsForUserType(userType)' ${OIDC_JS}
grep -q 'this.authOpenIDAdminGroups = settings.authOpenIDAdminGroups' ${SETTINGS_JS}

FILESYSTEM_JS=${APP_OUT}/server/controllers/FileSystemController.js
grep -q 'FILE_BROWSER_ROOT' ${FILESYSTEM_JS} || sed -i -f ${DIR}/filebrowser-root.sed ${FILESYSTEM_JS}
grep -q 'process.env.FILE_BROWSER_ROOT' ${FILESYSTEM_JS}

LOGGER_JS=${APP_OUT}/server/Logger.js
grep -q 'file logging disabled' ${LOGGER_JS} || sed -i -f ${DIR}/journal-logging.sed ${LOGGER_JS}
grep -q 'file logging disabled' ${LOGGER_JS}

cp -r /usr ${NODE_DIR}/usr
cp -r /lib ${NODE_DIR}/lib

cp ${DIR}/node.sh ${NODE_DIR}/node.sh
cp ${DIR}/ffmpeg.sh ${NODE_DIR}/ffmpeg.sh
cp ${DIR}/ffprobe.sh ${NODE_DIR}/ffprobe.sh

NUSQLITE3_SRC=$(find ${NODE_DIR} -name 'libnusqlite3*.so' 2>/dev/null | head -1)
test -n "${NUSQLITE3_SRC}"
ln -sfn "$(dirname "${NUSQLITE3_SRC}" | sed "s|^${NODE_DIR}/||")" ${NODE_DIR}/nusqlite3
test -f ${NODE_DIR}/nusqlite3/libnusqlite3.so

${NODE_DIR}/node.sh --version
du -sh ${APP_OUT} ${NODE_DIR}
