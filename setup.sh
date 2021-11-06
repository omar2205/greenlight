#!/bin/sh

# Used to run complex setup to prepare the API

# Run migrations
./migrate "-path ./migrations -database postgres://postgres:hello@db/greenlight up"