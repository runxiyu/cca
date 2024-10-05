# Frontend

We do not use a JavaScript package manager because we don't use any JavaScript
libraries at all.

## JavaScript linting

eslint may be installed separately via pgx if linting is desired.

## Building

Building is actually just minification.

```sh
go install github.com/tdewolff/minify/v2/cmd/minify@latest
make
```
