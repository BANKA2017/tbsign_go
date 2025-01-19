#!/bin/bash

## from env
libc="${CCC:-$(go env CC)}"
EXTERNAL_LDFLAGS="${EXTERNAL_LDFLAGS:-}"

flag=$(pwd)
export NUXT_BASE_PATH="/api"

mkdir tbsign_build
cd tbsign_build
git clone https://github.com/BANKA2017/tbsign_go_fe
git clone https://github.com/BANKA2017/tbsign_go

# fe
echo "build: frontend"
cd tbsign_go_fe
fe_commit_hash=$(git rev-parse HEAD)
export NUXT_COMMIT_HASH=$fe_commit_hash
yarn install
yarn run postinstall
yarn run generate
unset NUXT_BASE_PATH
unset NUXT_COMMIT_HASH
cd $flag/tbsign_build

# be
echo "build: backend"
cd tbsign_go
rm -r assets/dist
cp -R ../tbsign_go_fe/.output/public/ assets/dist
commit_hash=$(git rev-parse HEAD)
builtAt="$(date -Iseconds)"
goRuntime=$(go version | sed 's/go version go[0-9]*\.[0-9]*\.[0-9]* //')

CURRENT_LDFLAGS="\
-X 'github.com/BANKA2017/tbsign_go/share.BuiltAt=$builtAt' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildRuntime=$goRuntime' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildGitCommitHash=$commit_hash' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildEmbeddedFrontendGitCommitHash=$fe_commit_hash' \
-X 'github.com/BANKA2017/tbsign_go/share.BuildLibc=$libc' \
"

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
