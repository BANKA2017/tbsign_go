#!/bin/bash

mkdir tbsign_build
cd tbsign_build
flag=$(pwd)
export NUXT_BASE_PATH="/api"
git clone https://github.com/BANKA2017/tbsign_go_fe
git clone https://github.com/BANKA2017/tbsign_go

# fe
cd tbsign_go_fe
yarn install
yarn run generate
unset NUXT_BASE_PATH
cd $flag

# be
cd tbsign_go
rm -r assets/dist
cp -R ../tbsign_go_fe/.output/public/ assets/dist
go build
mv tbsign_go $flag
cd $flag
# rm -r tbsign_build
