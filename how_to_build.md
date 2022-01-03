# How to build krakend for use with AvaNet

## Prerequisite

Clone 
* https://github.com/avamonitoring/avanet-gateway-access-control
* https://github.com/avamonitoring/avanet-gateway-access-logging
* https://github.com/avamonitoring/lura

into the same parent directory as this repo.

## Build

```
docker run --rm -it -v //c/path/to/avanet-api-gateway:/go/github.com/avamonitoring/avanet-api-gateway -v //c/path/to/avanet-gateway-access-logging:/go/github.com/avamonitoring/avanet-gateway-access-logging -v //c/path/to/avanet-gateway-access-control:/go/github.com/avamonitoring/avanet-gateway-access-control -v //c/path/to/lura:/go/github.com/avamonitoring/lura golang:1.12

# Needed for old 1.12 image
apt update && apt install ca-certificates libgnutls30 -y

cd /go/github.com/avamonitoring/avanet-api-gateway
make build
```

Original claims to use Go 1.11 but `go.mod` indicates 1.12.
`Makefile` has `build_on_docker` with version 1.16.4.
We'll use 1.13 to avoid the workarounds needed for 1.12.

## Build Docker image

The build above will put `krakend` binary and this directory.

Thereafter, do
```
docker build . -t 059741451001.dkr.ecr.eu-north-1.amazonaws.com/krakend:my_tag
```