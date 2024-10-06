#!/bin/bash

set -xeu

home_dir="${BASH_SOURCE[0]%/*}/.."
cd "$home_dir"

make build/static/student.js
gsed -i 's;wss://localhost.runxiyu.org:8080/ws;wss://dev.runxiyu.org/ws;g' build/static/student.js

mv dist/cca dist/ccae
GOOS=linux GOARCH=amd64 make dist/cca

rm build/static/student.js

rsync -v dist/cca root@runxiyu.org:/srv/dev/cca

ssh root@runxiyu.org pkill cca

rm dist/cca
mv dist/ccae dist/cca
