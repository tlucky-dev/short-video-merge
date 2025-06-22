#!/bin/bash

GOOS=windows GOARCH=amd64 go build -o video_merger.exe video_merger.go