#!/usr/bin/env bash

# Service directories that use shared module
Services=("poggadaj-tcp" "poggadaj-http" "poggadaj-api")

for service in "${Services[@]}"
do
  echo "Updating shared module for $service"
  rm -r "$service"/shared/*
  cp -r poggadaj-shared/* "$service"/shared/
done

echo "Updated"
