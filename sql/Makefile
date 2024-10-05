.PHONY: sql

docs: ../dist/sql/schema.sql

../dist/sql/%: %
	mkdir -p ../dist/sql
	cp $< $@
