.PHONY: frontend

frontend: ../dist/static/student.js ../dist/static/style.css

../dist/static/student.js:
	mkdir -p ../dist/static
	minify student.js -o ../dist/static/student.js

../dist/static/style.css:
	mkdir -p ../dist/static
	minify style.css -o ../dist/static/style.css
