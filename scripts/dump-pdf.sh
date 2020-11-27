#!/bin/bash

set -eou

/usr/local/bin/server $PWD &

sleep 1

wkhtmltopdf \
    --grayscale \
    --print-media-type \
    http://localhost:8080/about/ \
    about-danny-berger.pdf

kill %1
