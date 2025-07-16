#!/bin/bash

set -e

wget -qO- "https://api.wordpress.org/plugins/info/1.0/$( echo $DEP_NAME | cut -c8- ).json" | jq -r '.version'
