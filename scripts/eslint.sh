#!/bin/sh

set -e

cd frontend

eslint . "$@"
