#!/bin/sh

case "$1" in
  start)
    npm run migrate up
    exec npm run start:dev
    ;;
esac
