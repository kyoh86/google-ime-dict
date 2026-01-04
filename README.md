# gobase

A template of the go-app project

[![PkgGoDev](https://pkg.go.dev/badge/kyoh86/gobase)](https://pkg.go.dev/kyoh86/gobase)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/gobase)](https://goreportcard.com/report/github.com/kyoh86/gobase)
[![Release](https://github.com/kyoh86/gobase/workflows/Release/badge.svg)](https://github.com/kyoh86/gobase/releases)

## Description

```console
$ gobase man
```

`gobase` provides a template of the go-app project.

## Install

### For Golang developers

```console
$ go get github.com/kyoh86/gobase/cmd/gobase
```

### Homebrew/Linuxbrew

```console
$ brew tap kyoh86/tap
$ brew update
$ brew install kyoh86/tap/gobase
```

### Makepkg

```console
$ mkdir -p gobase_build && \
  cd gobase_build && \
  curl -iL --fail --silent https://github.com/kyoh86/gobase/releases/latest/download/gobase_PKGBUILD.tar.gz | tar -xvz
$ makepkg -i
```

## Available commands

Use `gobase [command] --help` for more information about a command.
Or see the manual in [usage/gobase.md](./usage/gobase.md).

## Commands

Manual: [usage/gobase.md](./usage/gobase.md).

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This software is released under the [MIT License](http://www.opensource.org/licenses/MIT), see LICENSE.
