#!/bin/bash
goose -dir ./migrations postgres $DATABASE_URL?sslmode=disable $@