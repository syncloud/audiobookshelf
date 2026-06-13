#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
ROOT=$( cd "${DIR}/../.." && pwd )

PROJECT="${1:-desktop}"
NAME=audiobookshelf
export PLAYWRIGHT_DOMAIN="${PLAYWRIGHT_DOMAIN:-bookworm.com}"
export PLAYWRIGHT_USER="${PLAYWRIGHT_USER:-user}"
export PLAYWRIGHT_PASSWORD="${PLAYWRIGHT_PASSWORD:-Password1}"
export PLAYWRIGHT_PROJECT="${PROJECT}"
export PLAYWRIGHT_DEVICE_HOST="${NAME}.${PLAYWRIGHT_DOMAIN}"
export PLAYWRIGHT_SSH_PASSWORD="${PLAYWRIGHT_PASSWORD}"
export PLAYWRIGHT_ARTIFACT_DIR="${ROOT}/artifact"
export PLAYWRIGHT_TESTDIR=./specs

DOMAIN="$PLAYWRIGHT_DOMAIN"
APP_DOMAIN="${NAME}.${DOMAIN}"
getent hosts $APP_DOMAIN | sed "s/$APP_DOMAIN/auth.$DOMAIN/g" | tee -a /etc/hosts
cat /etc/hosts

apt-get update -qq
apt-get install -y -qq sshpass openssh-client ffmpeg

BOOK_DIR="${DIR}/.book"
rm -rf "${BOOK_DIR}"
mkdir -p "${BOOK_DIR}"
ffmpeg -y -f lavfi -i anullsrc=r=44100:cl=mono -t 60 -b:a 128k \
  -metadata title="Test Book ${PROJECT}" -metadata album="Test Book ${PROJECT}" -metadata artist="Test Author" \
  "${BOOK_DIR}/Test Book ${PROJECT}.mp3" >/dev/null 2>&1
export PLAYWRIGHT_BOOK_FILE="${BOOK_DIR}/Test Book ${PROJECT}.mp3"

ffmpeg -y -f lavfi -i anullsrc=r=44100:cl=mono -t 60 -b:a 128k \
  -metadata title="Rescan Book ${PROJECT}" -metadata album="Rescan Book ${PROJECT}" -metadata artist="Test Author" \
  "${BOOK_DIR}/Rescan Book ${PROJECT}.mp3" >/dev/null 2>&1
export PLAYWRIGHT_RESCAN_BOOK_FILE="${BOOK_DIR}/Rescan Book ${PROJECT}.mp3"
export PLAYWRIGHT_SNAP=$(cd ${ROOT} && realpath "$(cat package.name)")

AUDIOBOOK_DIR="${DIR}/.audiobook/Test Audiobook ${PROJECT}"
rm -rf "${DIR}/.audiobook"
mkdir -p "${AUDIOBOOK_DIR}"
for i in $(seq -w 1 35); do
  ffmpeg -y -f lavfi -i anullsrc=r=44100:cl=mono -t 300 -b:a 256k \
    -metadata title="Chapter ${i}" -metadata album="Test Audiobook" -metadata artist="Test Author" \
    "${AUDIOBOOK_DIR}/${i}.mp3" >/dev/null 2>&1
done
du -sh "${AUDIOBOOK_DIR}"
export PLAYWRIGHT_AUDIOBOOK_DIR="${AUDIOBOOK_DIR}"

cd ${DIR}
npm ci
npx playwright test --project="${PROJECT}"
