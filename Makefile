.PHONY: default iadocs docs build_iadocs build_docs setcap clean

default: dist/cca docs iadocs

# Docs file lists

DOCS_FILES := admin_handbook.html cca.scfg.example azure.json courses_example.csv schema.sql drop.sql
IADOCS_FILES := index.html cover_page.htm appendix.pdf crita_planning.pdf critb_design.pdf \
                critb_recordoftasks.htm critc_development.pdf critd_functionality.pdf crite_evaluation.pdf

# Create docs and iadocs targets using patterns

docs: $(DOCS_FILES:%=dist/docs/%)
iadocs: $(IADOCS_FILES:%=dist/iadocs/%)

# Final binary with embedded stuff

dist/cca: go.* *.go build/static/style.css build/static/student.js templates/* \
          $(DOCS_FILES:%=build/docs/%) $(IADOCS_FILES:%=build/iadocs/%) \
          .editorconfig .gitignore .gitattributes scripts/* sql/* docs/* iadocs/* README.md LICENSE Makefile
	mkdir -p dist
	go build -o $@

# Generic docs rules

dist/docs/%: build/docs/%
	mkdir -p $(@D)
	cp $< $@

build/docs/%.sql: sql/%.sql
	mkdir -p $(@D)
	cp $< $@

build/docs/%.csv: docs/%.csv
	mkdir -p $(@D)
	cp $< $@

build/docs/%.html: docs/%.md docs/handbook.css
	mkdir -p $(@D)
	pandoc --embed-resources --wrap none --standalone -t html -f markdown --css docs/handbook.css $< | gominify --type html -o $@

# Extra docs

build/docs/cca.scfg.example: docs/cca.scfg.example
	mkdir -p $(@D)
	cp $< $@

build/docs/azure.json: docs/azure.json
	mkdir -p $(@D)
	cp $< $@

# IA documentation

dist/iadocs/%: build/iadocs/%
	mkdir -p $(@D)
	cp $< $@

build/iadocs/%.htm: iadocs/%.htm
	mkdir -p $(@D)
	gominify --html-keep-end-tags --html-keep-document-tags -o $@ $<

build/iadocs/index.html: build/iadocs/cover_page.htm
	cp $< $@

build/iadocs/%.pdf: iadocs/%.tex build/iadocs/header.texinc build/iadocs/bib.bib
	mkdir -p $(@D)
	lualatex -interaction batchmode -output-directory=build/iadocs $<
	biber --output-directory=build/iadocs build/$(<:.tex=.bcf)
	lualatex -interaction batchmode -output-directory=build/iadocs $<
	lualatex -interaction batchmode -output-directory=build/iadocs $<

# Special case for Criterion C which needs the appendix's references

build/iadocs/critc_development.pdf: iadocs/critc_development.tex build/iadocs/header.texinc build/iadocs/bib.bib build/iadocs/appendix.pdf
	# Technically I need build/iadocs/appendix.aux instead of build/iadocs/appendix.pdf
	mkdir -p $(@D)
	cp iadocs/critc_development.tex $(@D)
	cd $(@D) && \
		lualatex -interaction batchmode -shell-escape critc_development && \
		biber critc_development && \
		lualatex -interaction batchmode -shell-escape critc_development && \
		lualatex -interaction batchmode -shell-escape critc_development

# Special case for the appendix and the PDF'ed source code

build/iadocs/appendix.pdf: iadocs/appendix.tex build/iadocs/source.gen build/iadocs/agpl.texinc
	mkdir -p $(@D)
	lualatex -interaction batchmode -shell-escape -output-directory=build/iadocs $<
	lualatex -interaction batchmode -shell-escape -output-directory=build/iadocs $<

build/iadocs/source.gen: go.* *.go frontend/*.css frontend/*.ts templates/* scripts/latexify-source.sh \
                        docs/* sql/* scripts/* iadocs/*.tex iadocs/*.texinc iadocs/bib.bib Makefile \
                        README.md LICENSE .editorconfig .gitignore .gitattributes
	mkdir -p $(@D)
	scripts/latexify-source.sh

# TeX includes and bibliography files could just be copied over

build/iadocs/%.texinc: iadocs/%.texinc
	mkdir -p $(@D)
	cp $< $@

build/iadocs/%.bib: iadocs/%.bib
	mkdir -p $(@D)
	cp $< $@

# Minified files will be embedded from build/static/

build/static/style.css: frontend/style.css
	mkdir -p $(@D)
	gominify -o $@ $<

build/static/student.js: frontend/student.ts
	mkdir -p $(@D)
	tsc $< --target ES6 --strict --noImplicitAny --outFile $@
	gominify -o $@ $@

# Cleaning (git clean -xfd is a bit too aggressive, I lost my config once)

clean:
	rm -rf dist build
