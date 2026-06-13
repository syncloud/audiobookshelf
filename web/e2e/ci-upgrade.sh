#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
ROOT=$( cd "${DIR}/../.." && pwd )

NAME=audiobookshelf
export PLAYWRIGHT_DOMAIN="${PLAYWRIGHT_DOMAIN:-bookworm.com}"
export PLAYWRIGHT_USER="${PLAYWRIGHT_USER:-user}"
export PLAYWRIGHT_PASSWORD="${PLAYWRIGHT_PASSWORD:-Password1}"
export PLAYWRIGHT_PROJECT=desktop
export PLAYWRIGHT_DEVICE_HOST="${NAME}.${PLAYWRIGHT_DOMAIN}"
export PLAYWRIGHT_SSH_PASSWORD="${PLAYWRIGHT_PASSWORD}"
export PLAYWRIGHT_ARTIFACT_DIR="${ROOT}/artifact"
export PLAYWRIGHT_TESTDIR=./specs-upgrade

DOMAIN="$PLAYWRIGHT_DOMAIN"
APP_DOMAIN="${NAME}.${DOMAIN}"
getent hosts $APP_DOMAIN | sed "s/$APP_DOMAIN/auth.$DOMAIN/g" | tee -a /etc/hosts
cat /etc/hosts

apt-get update -qq
apt-get install -y -qq sshpass openssh-client

export PLAYWRIGHT_SNAP=$(cd ${ROOT} && realpath "$(cat package.name)")

cd ${DIR}
npm ci
npx playwright test --project=desktop
