.PHONY: iadocs

iadocs: ../dist/iadocs/cover_page.htm ../dist/iadocs/crita_planning.pdf ../dist/iadocs/critb_design.pdf ../dist/iadocs/critb_recordoftasks.pdf ../dist/iadocs/critc_development.pdf ../dist/iadocs/critd_functionality.pdf ../dist/iadocs/crite_evaluation.pdf ../dist/iadocs/appendix.pdf

../dist/iadocs/%.htm: %.htm
	mkdir -p ../dist/iadocs
	cp $< $@

../dist/iadocs/%.pdf: %.tex header.inc
	mkdir -p ../dist/iadocs
	lualatex $<
	lualatex $<
	mv $$(basename $@) $@
