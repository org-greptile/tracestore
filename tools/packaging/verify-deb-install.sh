#!/usr/bin/env bash

set -euxo pipefail

# Install tracestore and check it's running
dpkg -i ./tracestore_*_linux_amd64.deb
[ "$(systemctl is-active tracestore)" = "active" ] || (echo "tracestore is inactive" && exit 1)

# Wait for tracestore to be ready.
apt update && apt install -y curl
./wait-for-ready.sh
