#!/bin/bash

set -euo pipefail

sudo chmod 1777 /tmp

sudo apt update
sudo apt install -y vim

# exiftool
mkdir /tmp/exiftool \
  && cd /tmp/exiftool \
  && wget -LO Image-ExifTool.tar.gz 'https://download.sourceforge.net/exiftool/Image-ExifTool-13.59.tar.gz' \
  && tar -xzf Image-ExifTool.tar.gz \
  && cd Image-ExifTool-* \
  && perl Makefile.PL \
  && sudo make install \
  && sudo rm -fr /tmp/exiftool

# tailwindcss
sudo wget -O /usr/local/bin/tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v4.2.2/tailwindcss-linux-arm64 \
  && sudo chmod +x /usr/local/bin/tailwindcss

# gcloud (https://docs.cloud.google.com/sdk/docs/install-sdk#deb)
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg \
  && echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list \
  && sudo apt-get update \
  && sudo apt-get install -y google-cloud-cli

# build scripts
sudo apt install -y xxd brotli zstd
