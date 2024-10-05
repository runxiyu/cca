.PHONY: docs

docs: ../dist/docs/admin_handbook.html ../dist/docs/cca.scfg.example ../dist/docs/handbook.css

../dist/docs/%: %
	mkdir -p ../dist/docs
	cp $< $@
