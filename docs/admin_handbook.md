---
lang: en
title: CCA Admin Handbook
viewport: width=device-width, initial-scale=1
---

## Introduction

This handbook guides you in installing, configuring, and managing your CCA Selection System (CCASS) instance.

## Downloading

You may obtain a stable or development version. The stable version is recommended for production.

-   To obtain a stable version, go to the [release page](https://git.sr.ht/~runxiyu/cca/refs) and download the latest version that is not a pre-release.
-   To obtain an unstable development version, clone the development repository at [`https://git.sr.ht/~runxiyu/cca`](https://git.sr.ht/~runxiyu/cca), or download the latest development snapshot's tarball at [`https://git.runxiyu.org/ykps/cca.git/snapshot/cca-master.tar.gz`](https://git.runxiyu.org/ykps/cca.git/snapshot/cca-master.tar.gz).

## External dependencies

You need a [Go](https://go.dev) toolchain, [Pygments](https://pygments.org/), [GNU make](https://www.gnu.org/software/make/), [TeX Live](https://tug.org/texlive/) and [minify](https://github.com/tdewolff/minify). If you have everything else, you could install minify via `make minifier`, which would build and install it with your Go toolchain.

The Go toolchain will fetch more dependencies. You may wish to set a Go proxy (such as via `export GOPROXY='https://goproxy.io'`) if it stalls or is too slow. This is likely necessary for users in Mainland China due to firewall restrictions.

## Building

Just type `make`.

The built files will appear in `dist/`. The binary, with all runtime resources other than the configuration file embedded, is located at `dist/cca`. A minified copy of the documentation, including a sample configuration file, is located at `dist/docs/`.

## Configuration

Copy [the example configuration file](./cca.scfg.example) to `cca.scfg` in the working directory where you intend to run CCASS. Then edit it according to the comments, though you may wish to pay attention to the following:

-   CCASS natively supports serving over clear text HTTP or over HTTPS. HTTPS is required for production setups as Microsoft Entra ID does not allow clear-text HTTP redirect URLs for non-`localhost` access.
-   Note that CCASS is designed to be directly exposed to clients due to the lacking performance of standard reverse proxy setups, although there is nothing that otherwise prevents it from being used behind a reverse proxy. Reverse proxies must forward WebSocket connection upgrade headers for all requests to the `/ws` endpoint.
-   You must [create an app registration on the Azure portal](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade) and complete the corresponding configuration options.
-   `perf/sendq` should be set to roughly the number of expected students making concurrent choices.

## Database setup

A working PostgreSQL setup is required. It is recommended to set up UNIX socket authentication and set the user running CCASS as the database owner while creating the database.

Before first run, run <code>psql <i>dbname</i> -f sql/schema.sql</code> to create the database tables, where <code><i>dbname</i></code> is the name of the database.

Using the same database for different versions of CCASS is currently unsupported, although it should be trivial to manually migrate the database.