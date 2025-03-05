#!/usr/bin/env bash

set -euxo pipefail

# Import the Acme GPG key
rpm --import https://packages.acme.com/gpg.key

# Install tracestore and check it's running
rpm -i ./tracestore_*_linux_amd64.rpm
[ "$(systemctl is-active tracestore)" = "active" ] || (echo "tracestore is inactive" && exit 1)

# Wait for tracestore to be ready.
./wait-for-ready.sh
