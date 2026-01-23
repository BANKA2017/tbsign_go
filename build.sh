#!/bin/bash

## from env
PUBLISH_TYPE="${PUBLISH_TYPE:-binary}"
EXTERNAL_LDFLAGS="${EXTERNAL_LDFLAGS:-}"

flag=$(pwd)
export NUXT_BASE_PATH="/api"
export NUXT_USE_COOKIE_TOKEN="1"

mkdir tbsign_build
cd tbsign_build
git clone --depth=1 https://github.com/BANKA2017/tbsign_go_fe tbsign_go_fe
git clone --depth=1 https://github.com/BANKA2017/tbsign_go tbsign_go
curl -o assets/ca/cacert.pem https://curl.se/ca/cacert.pem

# fe
echo "build: frontend"
cd tbsign_go_fe
fe_commit_hash=$(git rev-parse HEAD)
export NUXT_COMMIT_HASH=$fe_commit_hash
yarn install
yarn run postinstall
yarn run generate
unset NUXT_BASE_PATH
unset NUXT_USE_COOKIE_TOKEN
unset NUXT_COMMIT_HASH
cd $flag/tbsign_build

# be
echo "build: backend"
cd tbsign_go
rm assets/ca/.gitkeep
rm -r assets/dist
cp -R ../tbsign_go_fe/.output/public/ assets/dist
commit_hash=$(git rev-parse HEAD)
build_at="$(date -u "+%Y-%m-%dT%H:%M:%SZ")"
go_runtime=$(go env GOOS)/$(go env GOARCH)

CURRENT_LDFLAGS="-X 'github.com/BANKA2017/tbsign_go/share.BuiltAt=$build_at' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildRuntime=$go_runtime' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildGitCommitHash=$commit_hash' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildEmbeddedFrontendGitCommitHash=$fe_commit_hash' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildPublishType=$PUBLISH_TYPE'"

if [ -n "$EXTERNAL_LDFLAGS" ]; then
    LDFLAGS="$CURRENT_LDFLAGS $EXTERNAL_LDFLAGS"
else
    LDFLAGS="$CURRENT_LDFLAGS"
fi

CGO_ENABLED=1 go build -ldflags "$LDFLAGS" -tags netgo
mv tbsign_go $flag
cd $flag

## clean
# rm -r tbsign_build
