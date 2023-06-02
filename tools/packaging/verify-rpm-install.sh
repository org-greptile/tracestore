#!/usr/bin/env sh

set -euxo pipefail

docker ps
image="$(docker ps --filter ancestor=jrei/systemd-centos:8 --latest --format "{{.ID}}")"
echo "Running on container: ${image}"

dir="."
if [ -n "${CI}" ]; then
    dir="/drone/src"
fi
echo "Running on directory: ${dir}"

cat <<EOF | docker exec --interactive "${image}" sh
    set -x

    # Import the Acme GPG key
    rpm --import https://packages.acme.com/gpg.key

    # Install logstore and check it's running
    rpm -i ${dir}/dist/tracestore*_amd64.rpm
    [ "\$(systemctl is-active tracestore)" = "active" ] || (echo "tracestore is inactive" && exit 1)

    # Wait for tracestore to be ready. The script is cat-ed because it is passed to docker exec
    $(cat ${dir}/tools/packaging/wait-for-ready.sh)
EOF