#!/bin/sh

set -e

cd static

NPM_CONFIG_REGISTRY=https://registry.npmmirror.com npx eslint . "$@"
