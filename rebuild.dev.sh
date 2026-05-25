#!/bin/bash

set -e

docker build -f Dockerfile.dev -t "finance-tracker-dev" .