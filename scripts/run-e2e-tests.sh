#!/bin/bash
set -e

# Build the test image
docker build -t broom-test -f Dockerfile.test .

# Run the tests in a container
docker run --rm broom-test

# Clean up
docker rmi broom-test
