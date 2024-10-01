#!/bin/sh

LC_ALL="C" exec /opt/homebrew/opt/postgresql@15/bin/postgres -D /opt/homebrew/var/postgresql@15 -d 2 "$@"
