# gimedic

A template of the go-app project

[![PkgGoDev](https://pkg.go.dev/badge/kyoh86/gimedic)](https://pkg.go.dev/kyoh86/gimedic)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/gimedic)](https://goreportcard.com/report/github.com/kyoh86/gimedic)
[![Release](https://github.com/kyoh86/gimedic/workflows/Release/badge.svg)](https://github.com/kyoh86/gimedic/releases)

## Description

```console
$ gimedic man
```

`gimedic` provides a template of the go-app project.

## Install

### For Golang developers

```console
$ go get github.com/kyoh86/gimedic/cmd/gimedic
```

### Homebrew/Linuxbrew

```console
$ brew tap kyoh86/tap
$ brew update
$ brew install kyoh86/tap/gimedic
```

### Makepkg

```console
$ mkdir -p gimedic_build && \
  cd gimedic_build && \
  curl -iL --fail --silent https://github.com/kyoh86/gimedic/releases/latest/download/gimedic_PKGBUILD.tar.gz | tar -xvz
$ makepkg -i
```

## Available commands

Use `gimedic [command] --help` for more information about a command.
Or see the manual in [usage/gimedic.md](./usage/gimedic.md).

## Commands

Manual: [usage/gimedic.md](./usage/gimedic.md).

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This software is released under the [MIT License](http://www.opensource.org/licenses/MIT), see LICENSE.
