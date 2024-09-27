.SUFFIXES: .texi .pdf .html
.PHONY: all

all: admin_handbook.html user_handbook.html developer_handbook.html admin_handbook.pdf user_handbook.pdf developer_handbook.pdf

.texi.pdf:
	texi2pdf $<

.texi.html:
	makeinfo --html --no-split $<
