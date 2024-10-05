# GNU Make is required because BSD Make has bad pattern matching

.SUFFIXES: .tex .pdf

.PHONY: all

all: crita_planning.pdf critb_design.pdf critb_recordoftasks.pdf critc_development.pdf critd_functionality.pdf crite_evaluation.pdf

%.pdf: %.tex header.inc
	lualatex $<
	lualatex $<
