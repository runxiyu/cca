.PHONY: default backend frontend

default: backend tmpl frontend

backend:
	make -C backend

tmpl:
	make -C tmpl

frontend:
	make -C frontend
