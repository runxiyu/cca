#!/bin/bash

set -eu

targetfile="$(realpath -- build/iadocs/source.gen)"

printf '\n' > "$targetfile"

include_code() {
	lang="$1"
	tabsize="$2"
	shift 2
	for i in "$@"
	do
		printf '\\section{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
		printf '\\label{%s}\n' "$(sed 's/_/_/g' <<< "$i")" >> "$targetfile"
		printf '\\inputminted[breaklines, tabsize=%s, texcomments]{%s}{%s}\n' "$tabsize" "$lang" "$i" >> "$targetfile"
	done
}

chapter() {
	printf '\\chapter{%s}\n' "$*" >> "$targetfile"
}

chapter Backend source code
include_code go 8 *.go
include_code text 8 go.*

chapter Frontend source code
include_code javascript 4 frontend/*.js
include_code css 8 frontend/*.css

chapter HTML templates
include_code html 2 templates/*.html

chapter Build system and auxiliary scripts
include_code makefile 8 Makefile
include_code bash 8 scripts/*.sh

chapter SQL scripts
include_code postgresql 8 sql/*.sql

chapter Production documentation
include_code markdown 2 docs/*.md
include_code css 8 docs/*.css
include_code text 8 docs/*.csv docs/cca.scfg.example

chapter IA documentation
include_code latex 8 iadocs/*.tex iadocs/*.texinc
include_code bib 8 iadocs/*.bib
