#!/bin/sh
sudo ./cetuspoolsui -stop
git pull
go build -o cetuspoolsui ./main.go
sudo ./cetuspoolsui -start
