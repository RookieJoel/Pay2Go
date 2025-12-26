#!/bin/bash

go build -o main main.go
mv main cmd/
cd cmd
./main