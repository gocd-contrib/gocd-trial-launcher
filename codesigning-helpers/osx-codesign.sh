#!/bin/bash

set -e

cd "$(dirname "$0")/../../codesigning"
rake --trace osx:sign_single_binary[../dist/darwin/amd64/run-gocd,../osx-launcher.zip]
