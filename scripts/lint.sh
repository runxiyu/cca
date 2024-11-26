#!/bin/sh
set -e
golangci-lint run --color=always --enable-all --disable=wsl,funlen,exportloopref,godox,lll,gochecknoglobals,depguard,cyclop,gosmopolitan,nlreturn,varnamelen,nestif,musttag,mnd,tagliatelle,gocognit,gocyclo,maintidx,dogsled,unparam,nonamedreturns,godot
