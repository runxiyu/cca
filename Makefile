# TODO: Use some variables to clean up the massive documentation file specifiers

.PHONY: cca default iadocs docs build_iadocs build_docs setcap

default: dist/cca docs iadocs

cca: dist/cca

docs: dist/docs/admin_handbook.html dist/docs/cca.scfg.example dist/docs/azure.json dist/docs/courses_example.csv dist/docs/schema.sql dist/docs/drop.sql

iadocs: dist/iadocs/index.html dist/iadocs/cover_page.htm dist/iadocs/appendix.pdf dist/iadocs/crita_planning.pdf dist/iadocs/critb_design.pdf dist/iadocs/critb_recordoftasks.htm dist/iadocs/critc_development.pdf dist/iadocs/critd_functionality.pdf dist/iadocs/crite_evaluation.pdf

# Final binary which tries to embed stuff
dist/cca: go.* *.go build/static/style.css build/static/student.js templates/* build/docs/admin_handbook.html build/docs/cca.scfg.example build/docs/azure.json build/iadocs/index.html build/iadocs/cover_page.htm build/iadocs/appendix.pdf build/iadocs/crita_planning.pdf build/iadocs/critb_design.pdf build/iadocs/critb_recordoftasks.htm build/iadocs/critc_development.pdf build/iadocs/critd_functionality.pdf build/iadocs/crite_evaluation.pdf .editorconfig .gitignore .gitattributes scripts/* sql/* docs/* iadocs/* README.md LICENSE Makefile
	mkdir -p dist
	go build -o $@

# Documentation
dist/docs/%: build/docs/%
	mkdir -p dist/docs
	cp $< $@
build/docs/%.sql: sql/%.sql
	mkdir -p build/docs
	cp $< $@
build/docs/%.csv: docs/%.csv
	mkdir -p build/docs
	cp $< $@
build/docs/%.html: docs/%.md docs/handbook.css
	mkdir -p build/docs
	pandoc --embed-resources --wrap none --standalone -t html -f markdown --css docs/handbook.css $< | gominify --type html -o $@
build/docs/cca.scfg.example: docs/cca.scfg.example
	mkdir -p build/docs
	cp $< $@
build/docs/azure.json: docs/azure.json
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
	gominify --html-keep-end-tags --html-keep-document-tags -o $@ $<
build/iadocs/index.html: build/iadocs/cover_page.htm
	cp $< $@
build/iadocs/%.pdf: iadocs/%.tex build/iadocs/header.texinc
	mkdir -p build/iadocs
	lualatex -interaction batchmode -output-directory=build/iadocs $<
	lualatex -interaction batchmode -output-directory=build/iadocs $<
build/iadocs/appendix.pdf: iadocs/appendix.tex build/iadocs/source.gen build/iadocs/agpl.texinc
	mkdir -p build/iadocs
	lualatex -interaction batchmode -shell-escape -output-directory=build/iadocs $<
	lualatex -interaction batchmode -shell-escape -output-directory=build/iadocs $<
build/iadocs/source.gen: go.* *.go frontend/*.css frontend/*.js templates/* scripts/latexify-source.sh docs/* sql/* scripts/* iadocs/*.tex iadocs/*.texinc
	mkdir -p build/iadocs
	scripts/latexify-source.sh
build/iadocs/%.texinc: iadocs/%.texinc
	mkdir -p build/iadocs
	cp $< $@

# Temporary files in build/ to be embedded into the final binary
build/static/style.css: frontend/style.css
	mkdir -p build/static
	gominify -o $@ $<
build/static/student.js: frontend/student.js
	mkdir -p build/static
	gominify -o $@ $<

# Quick target to set capabilities
setcap: dist/cca
	setcap 'cap_net_bind_service=+ep' dist/cca
