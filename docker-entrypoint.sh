#!/bin/sh
sleep 3

if [ ! -d /work/cassl ]; then
    echo "Error: The /work/cassl directory does not exist!"
    echo "Please mount this directory before running the Docker container."
    exit 1
fi

if [ ! -f /work/cassl/rootCA.crt ] || [ ! -f /work/cassl/rootCA.key ]; then
    echo "The certificate and key files are missing in /work/cassl. Generating new CA..."
    /work/anyproxy/bin/anyproxy-ca --generate
    cp -a /root/.anyproxy/certificates/rootCA.* /work/cassl/
else
    echo "Using existing certificate and key files..."
    mkdir -p /root/.anyproxy/certificates
    cp -a /work/cassl/rootCA.* /root/.anyproxy/certificates/
fi

node /work/server.js
