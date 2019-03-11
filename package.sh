#!/bin/bash

for pkg in tar unzip gunzip; do
  if (which $pkg); then
    echo "has $pkg"
  else
    echo "no $pkg on container"
  fi
done
