.SUFFIXES: .texi .pdf .html

.texi.pdf:
	texi2pdf $<

.texi.html:
	makeinfo --html --no-split $<
