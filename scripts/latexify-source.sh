#!/bin/bash

set -eu

targetfile="$(realpath -- build/iadocs/source.gen)"

printf '\n' > "$targetfile"

printfile() {
	lang="$1"
	tabsize="$2"
	base="$3"
	shift 3
	for i in "$@"
	do
		printf '\\section{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
		printf '\\inputminted[breaklines, tabsize=%s]{%s}{%s/%s}\n' "$tabsize" "$lang" "$base" "$i" >> "$targetfile"
	done
}

printf '\\chapter{Backend source code}\n' >> "$targetfile"
printfile go 8 ./ *.go
printfile text 8 ./ go.*

printf '\\chapter{Frontend source code}\n' >> "$targetfile"
cd frontend
printfile javascript 4 ./frontend *.js
printfile css 8 ./frontend *.css

printf '\\chapter{HTML templates}\n' >> "$targetfile"
cd ../templates
printfile html 2 ./templates *.html

printf '\\chapter{Build system and auxiliary scripts}\n' >> "$targetfile"
cd ..
printfile makefile 8 ./ Makefile
cd scripts
printfile bash 8 ./scripts *.sh

printf '\\chapter{SQL scripts}\n' >> "$targetfile"
cd ../sql
printfile postgresql 8 ./sql *.sql

printf '\\chapter{Production documentation}\n' >> "$targetfile"
cd ../docs
printfile html 2 ./docs *.html
printfile css 8 ./docs *.css
printfile text 8 ./docs *.csv cca.scfg.example
