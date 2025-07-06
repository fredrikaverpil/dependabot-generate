#!/bin/sh -l

# This script acts as the entrypoint for the Docker container.
# It executes the compiled Go binary, passing along all the
# command-line arguments it received.

/app/dependabot-generate "$@"
