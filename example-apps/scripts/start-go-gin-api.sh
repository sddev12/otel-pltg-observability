#!/bin/bash
export LOG_LEVEL=INFO
export GIN_MODE=release
cd ../go-gin-api
go run main.go
