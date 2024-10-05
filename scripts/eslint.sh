#!/bin/sh

set -e

cd frontend

NPM_CONFIG_REGISTRY=https://registry.npmmirror.com npx eslint . "$@"
