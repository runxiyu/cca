#!/bin/bash

set -eu

targetfile="$(realpath -- build/iadocs/source.gen)"

printf '\n' > "$targetfile"

printf '\\section{Backend source code}\n' >> "$targetfile"
for i in *.go
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{go}{%s}\n' "$i" >> "$targetfile"
done
for i in go.*
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{text}{%s}\n' "$i" >> "$targetfile"
done

printf '\\section{Frontend source code}\n' >> "$targetfile"
cd frontend
for i in *.js
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=4]{javascript}{frontend/%s}\n' "$i" >> "$targetfile"
done
for i in *.css
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{css}{frontend/%s}\n' "$i" >> "$targetfile"
done

printf '\\section{HTML templates}\n' >> "$targetfile"
cd ../tmpl
for i in *.html
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=2]{html}{tmpl/%s}\n' "$i" >> "$targetfile"
done

printf '\\section{Build system and auxiliary scripts}\n' >> "$targetfile"
cd ..
for i in Makefile
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{makefile}{%s}\n' "$i" >> "$targetfile"
done
cd scripts
for i in *
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{bash}{scripts/%s}\n' "$i" >> "$targetfile"
done

printf '\\section{SQL scripts}\n' >> "$targetfile"
cd ../sql
for i in *
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{postgresql}{sql/%s}\n' "$i" >> "$targetfile"
done

printf '\\section{Production documentation}\n' >> "$targetfile"
cd ../docs
for i in *.html
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=2]{html}{docs/%s}\n' "$i" >> "$targetfile"
done
for i in *.css
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{css}{docs/%s}\n' "$i" >> "$targetfile"
done
for i in cca.scfg.example
do
	printf '\\subsection{%s}\n' "$(sed 's/_/\\_/g' <<< "$i")" >> "$targetfile"
	printf '\\inputminted[breaklines, tabsize=8]{text}{docs/%s}\n' "$i" >> "$targetfile"
done
