.PHONY: default backend tmpl frontend docs sql iadocs

default: backend tmpl frontend docs sql iadocs

backend:
	make -C backend

tmpl:
	make -C tmpl

frontend:
	make -C frontend

docs:
	make -C docs

sql:
	make -C sql

iadocs:
	make -C iadocs
