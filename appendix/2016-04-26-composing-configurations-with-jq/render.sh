#!/bin/bash

set -eu

if [ -n "${STDOUT:-}" ] ; then
  exec 1>$STDOUT
fi

manifest="${1:-$0.jq}"
manifest_dir=$( cd "$( dirname "$manifest" )" && pwd )

network_dir=$( dirname "$manifest_dir" )

jq_args="${jq_args:-}"

if [ -e "$network_dir/network.json" ] ; then
  jq_args="$jq_args --argfile network $network_dir/network.json"
fi

if [ -n "$( find . -maxdepth 1 -name '*-stack' 2>/dev/null )" ] ; then
  for stack_dir in *-stack ; do
    stack_name=$( basename "$stack_dir" | sed 's/-stack$//' | tr '-' '_' )
    stack_arn=$( cat $stack_dir/arn.txt )
    stack_arn="${stack_arn:-UNKNOWN}"
  
    jq_args="$jq_args --argfile ${stack_name}_stack $stack_dir/resources.json"
    jq_args="$jq_args --arg ${stack_name}_stack_arn \"$stack_arn\""
    jq_args="$jq_args --argfile ${stack_name}_stack_outputs $stack_dir/outputs.json"
  done
fi

if [ -n "$( cd $manifest_dir ; find config -not -type d 2>/dev/null )" ] ; then
  for config_path in $( cd $manifest_dir ; find config -not -type d ) ; do
    config_name=$( echo "$config_path" | sed 's/[\/\-\.]/_/g' )
  
    if [[ "$config_path" =~ ".json" ]] ; then
      config_name=$( echo "$config_name" | sed 's/_json$//' )
      jq_args="$jq_args --argfile $config_name $manifest_dir/$config_path"
    else
      TMP=$( mktemp -t "$(basename $0).XXXXXXXXXX" )
      jq -s -R '.' < $manifest_dir/$config_path > $TMP
      jq_args="$jq_args --argfile $config_name $TMP"
    fi
  done
fi

jq -S -n $jq_args -f "$manifest"
