# TODO: Use some variables to clean up the massive documentation file specifiers

.PHONY: cca default minifier iadocs docs build_iadocs build_docs

default: dist/cca docs iadocs

cca: dist/cca

docs: dist/docs/admin_handbook.html dist/docs/handbook.css dist/docs/cca.scfg.example

iadocs: dist/iadocs/index.html dist/iadocs/cover_page.htm dist/iadocs/appendix.pdf dist/iadocs/crita_planning.pdf dist/iadocs/critb_design.pdf dist/iadocs/critb_recordoftasks.htm dist/iadocs/critc_development.pdf dist/iadocs/critd_functionality.pdf dist/iadocs/crite_evaluation.pdf

# Final binary which tries to embed stuff
dist/cca: go.* *.go build/static/style.css build/static/student.js templates/* build/docs/admin_handbook.html build/docs/handbook.css build/docs/cca.scfg.example build/iadocs/index.html build/iadocs/cover_page.htm build/iadocs/appendix.pdf build/iadocs/crita_planning.pdf build/iadocs/critb_design.pdf build/iadocs/critb_recordoftasks.htm build/iadocs/critc_development.pdf build/iadocs/critd_functionality.pdf build/iadocs/crite_evaluation.pdf .editorconfig .gitignore .gitattributes scripts/* sql/* docs/* iadocs/* README.md LICENSE Makefile
	mkdir -p dist
	go build -o $@

# Documentation
dist/docs/%: build/docs/%
	mkdir -p dist/docs
	cp $< $@
build/docs/%.html: docs/%.html
	mkdir -p build/docs
	minify --html-keep-end-tags --html-keep-document-tags -o $@ $<
build/docs/handbook.css: docs/handbook.css
	mkdir -p build/docs
	minify -o $@ $<
build/docs/cca.scfg.example: docs/cca.scfg.example
	mkdir -p build/docs
	cp $< $@

# IA documentation
dist/iadocs/%.pdf: build/iadocs/%.pdf
	mkdir -p dist/iadocs
	cp $< $@
dist/iadocs/%.htm: build/iadocs/%.htm
	mkdir -p dist/iadocs
	cp $< $@
dist/iadocs/%.html: build/iadocs/%.html
	mkdir -p dist/iadocs
	cp $< $@
build/iadocs/%.htm: iadocs/%.htm
	mkdir -p build/iadocs
	minify --html-keep-end-tags --html-keep-document-tags -o $@ $<
build/iadocs/index.html: build/iadocs/cover_page.htm
	cp $< $@
build/iadocs/%.pdf: iadocs/%.tex build/iadocs/header.inc
	mkdir -p build/iadocs
	lualatex -interaction batchmode -output-directory=build/iadocs $<
	lualatex -interaction batchmode -output-directory=build/iadocs $<
build/iadocs/appendix.pdf: iadocs/appendix.tex build/iadocs/source.gen build/iadocs/agpl.inc
	mkdir -p build/iadocs
	lualatex -interaction batchmode -shell-escape -output-directory=build/iadocs $<
	lualatex -interaction batchmode -shell-escape -output-directory=build/iadocs $<
build/iadocs/source.gen: go.* *.go frontend/*.css frontend/*.js templates/* scripts/latexify-source.sh docs/* sql/* scripts/* iadocs/*.tex iadocs/*.inc
	mkdir -p build/iadocs
	scripts/latexify-source.sh
build/iadocs/%.inc: iadocs/%.inc
	mkdir -p build/iadocs
	cp $< $@

# Temporary files in build/ to be embedded into the final binary
build/static/style.css: frontend/style.css
	mkdir -p build/static
	minify -o $@ $<
build/static/student.js: frontend/student.js
	mkdir -p build/static
	minify -o $@ $<

# External dependencies
minifier:
	go install github.com/tdewolff/minify/v2/cmd/minify@latest
