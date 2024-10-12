#!/bin/bash

set -eu

targetfile="$(realpath -- build/iadocs/source.gen)"

printf '\n' > "$targetfile"

printfile() {
	lang="$1"
	tabsize="$2"
	shift 2
	for i in "$@"
	do
		printf '\\section{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
		printf '\\inputminted[breaklines, tabsize=%s]{%s}{%s}\n' "$tabsize" "$lang" "$i" >> "$targetfile"
	done
}

printf '\\chapter{Backend source code}\n' >> "$targetfile"
printfile go 8 *.go
printfile text 8 go.*

printf '\\chapter{Frontend source code}\n' >> "$targetfile"
printfile javascript 4 frontend/*.js
printfile css 8 frontend/*.css

printf '\\chapter{HTML templates}\n' >> "$targetfile"
printfile html 2 templates/*.html

printf '\\chapter{Build system and auxiliary scripts}\n' >> "$targetfile"
printfile makefile 8 Makefile
printfile bash 8 scripts/*.sh

printf '\\chapter{SQL scripts}\n' >> "$targetfile"
printfile postgresql 8 sql/*.sql

printf '\\chapter{Production documentation}\n' >> "$targetfile"
printfile html 2 docs/*.html
printfile css 8 docs/*.css
printfile text 8 docs/*.csv docs/cca.scfg.example

printf '\\chapter{IA documentation}\n' >> "$targetfile"
printfile latex 8 iadocs/*.tex iadocs/*.texinc
