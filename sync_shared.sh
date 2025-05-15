#!/usr/bin/env bash

# Service directories that use shared module
Services=("poggadaj-tcp")

for service in "${Services[@]}"
do
  echo "Updating shared module for $service"
  rm -r "$service"/shared/*
  cp poggadaj-shared/* "$service"/shared/
done

echo "Updated"
