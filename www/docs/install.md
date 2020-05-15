---
id: install
title: Installation
---

## Download

1. Download a prebuilt executable binary for your operating system from the [GitHub releases page](https://github.com/bojand/ghz/releases).
2. Unzip the archive and place the executable binary wherever you would like to run it from. Additionally consider adding the location directory in the `PATH` variable if you would like the `ghz` command to be available everywhere.

## Homebrew

```sh
brew install ghz
```

## Compile

**Clone**

```sh
git clone https://github.com/bojand/ghz
```

**Build using make**

```sh
make build
```

**Build using go**

```sh
cd cmd/ghz
go build .
```
