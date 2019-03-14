#!/bin/bash

set -e

cd "$(dirname "$0")"
chmod a+rx gocd/run-gocd

gocd/run-gocd
