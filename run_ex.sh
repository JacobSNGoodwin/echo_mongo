#!/bin/bash

# Script to provide "environment variables" via the command line
# We can provide two flags
# 1. -dburi : this is the mongo db uri, and defaults to mongodb://root:example@localhost:27017
# 2. -gcconfig : this holds the path the the json file with your Google cloud config
go build -o ./dist/main && ./dist/main -gcconfig=#