#!/bin/bash

export WELLKNOWN="/var/vcap/data/http/letsencrypt"
export PRIVATE_KEY="/var/vcap/jobs/http/etc/letsencrypt.key"
export CONTACT_EMAIL="<%= p('http.letsencrypt.email') %>"

mkdir -p "$WELLKNOWN"
