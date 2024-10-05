.PHONY: frontend

frontend: ../dist/static/student.js ../dist/static/style.css

../dist/static/%: %
	mkdir -p ../dist/static
	minify $< -o $@
