#!/bin/sh

set -e
set -u

cd updated-infra

git clone file://$PWD/../infra .

if [ -n "$target_ca_crt" ] ; then
  echo "$target_ca_crt" > /tmp/ca.crt
  target_args="--ca-cert /tmp/ca.crt"
fi

echo "$target bosh" >> /etc/hosts

gobosh target "https://bosh:25555" ${target_args:-}
gobosh --user "$username" --password "$password" log-in
gobosh deployment "$deployment"

gobosh ssh -r -c 'sudo /var/vcap/jobs/http/bin/letsencrypt --cron' | cut -f2 | tee /tmp/cron

if grep -q ' + Valid till' /tmp/cron ; then
  exit
fi

domain=$( grep '^Processing ' /tmp/cron | awk '{ print $2 }' )

echo 'Downloading new certificate'

gobosh ssh -r -c 'cat /var/vcap/jobs/http/etc/tls.crt' | cut -f2 > "$tls_crt"

echo 'Downloading new key'

gobosh ssh -r -c 'sudo cat /var/vcap/jobs/http/etc/tls.key' | cut -f2 > "$tls_key"

echo 'Committing'

git config user.email "$git_user_email"
git config user.name "$git_user_name"

git commit -m "Renew letsencrypt for $domain" -- "$tls_crt" "$tls_key"

echo 'Done'
