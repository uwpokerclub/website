#!/bin/bash

if [ $1 = "--test" ]; then
  shift
  echo "goose -dir ./migrations postgres $TEST_DATABASE_URL?sslmode=disable $@"
  goose -dir ./migrations postgres $TEST_DATABASE_URL?sslmode=disable $@
else
  goose -dir ./migrations postgres $DATABASE_URL?sslmode=disable $@
fi