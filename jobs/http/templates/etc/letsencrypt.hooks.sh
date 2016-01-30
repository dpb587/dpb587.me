#!/bin/bash

function deploy_challenge {
    local DOMAIN="${1}" TOKEN_FILENAME="${2}" TOKEN_VALUE="${3}"

    mkdir -p /var/vcap/data/http/letsencrypt/.well-known/acme-challenge

    echo "$TOKEN_VALUE" > "/var/vcap/data/http/letsencrypt/.well-known/acme-challenge/$TOKEN_FILENAME"

    chmod -R 755 /var/vcap/data/http/letsencrypt
}

function clean_challenge {
    local DOMAIN="${1}" TOKEN_FILENAME="${2}" TOKEN_VALUE="${3}"

    rm "/var/vcap/data/http/letsencrypt/.well-known/acme-challenge/$TOKEN_FILENAME"
}

function deploy_cert {
    local DOMAIN="${1}" KEYFILE="${2}" CERTFILE="${3}" CHAINFILE="${4}"

    cp "$KEYFILE" /var/vcap/jobs/http/etc/tls.key
    chmod 0600 "$KEYFILE"

    cp "$CHAINFILE" /var/vcap/jobs/http/etc/tls.crt
    chmod 0755 "$CHAINFILE"

    #/var/vcap/jobs/http/bin/nginx-control reload
}

HANDLER=$1; shift; $HANDLER $@
