#!/usr/bin/env bash

[ $# -ne 1 ] && echo "Usage: $0 [up|down]" && exit 1

case "$1" in
    up)
        podman-compose up -d
        podman ps -a
        ;;
    down)
        podman stop peerlogger-crawler peerlogger-db && podman rm peerlogger-crawler peerlogger-db
        ;;
    *)
        echo "Usage: $0 [up|down]"
        ;;
esac
