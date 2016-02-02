#!/bin/bash

current_dir=${PWD##*/}
packages=`go list ... | grep "github.com/mercadolibre/${current_dir}" | tr '\r\n' ','`

go test -coverprofile=/tmp/coverage.out -coverpkg $packages
go tool cover -html=/tmp/coverage.out
