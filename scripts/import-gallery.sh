#!/bin/bash

# brew install imagemagick jq yq cloudfoundry/tap/bosh
# go get github.com/dsoprea/go-exif/v2/exif-read-tool

set -eu

# directory of unzipped photos (e.g. ~/Downloads/2019\ Boring\ Gallery)
source_dir="$1"

# directory content subsection (e.g. content/photo/photo/2019-boring-gallery)
target_dir="$2"

# directory for generated assets (e.g. asset/gallery/2019-boring-gallery)
assets_dir="$3"

#
# ready, set, go.
#

mkdir -p "$target_dir" "$assets_dir"

log() { echo "$( date +%Y-%m-%dT%H:%M:%S )" "$@" >&2 ; }

log DEBUG starting import

patchyaml() {
  bosh interpolate \
    --ops-file "$2" \
    "$1" \
    > "${1}2"
  
  mv "${1}2" "$1"
}

resize() {
  metadata="$1"; shift
  input="$1"; shift
  output="$1"; shift
  key="$1"; shift

  if [[ "true" == "$( yq r -j "$metadata" | jq --arg key "$key" '.sizes[$key] != null' )" ]] ; then
    log DEBUG skipping "$output"
    return
  fi

  log DEBUG generating "$output"

  [ ! -f "$output" ] || rm "$output"

  convert "$@" -quality 90 "$input" "$output"
  if guetzli --quality 90 "$output" "$output.1" ; then
    mv "$output.1" "$output"
  else
    log ERROR guetzli failed
  fi

  patchyaml "$metadata" <(
    (
      echo "sha512-$( cat "$output" | openssl dgst -sha512 -binary | openssl base64 -A )"
    ) \
    | jq -csR \
      --arg key "$key" \
      --arg source "$output" \
      --arg bytes "$( stat -f '%z' "$output" )" \
      --arg height "$( sips -g pixelHeight "$output" | tail -n1 | awk '{ print $2 }' )" \
      --arg width "$( sips -g pixelWidth "$output" | tail -n1 | awk '{ print $2 }' )" \
      '
        [
          { "path": "/sizes?/HACKYHACKYPREFIX\($key)", "type": "replace", "value": { "height": ( $height | tonumber ), "width": ( $width | tonumber ), "bytes": ( $bytes | tonumber ), "source": $source, "integrity": ( . | split("\n") | map(select(length > 0)) ) } }
        ]
      '
  )

  log INFO generated "$output"
}

count="$( cd "$source_dir" ; ls -l | grep -v ^total | wc -l | awk '{ print $1 }' )"
ordering=0
firstDate=""
lastDate=""

while read line ; do
  ordering=$(($ordering + 1))

  file="$( echo "$line" | cut -d';' -f1 )"

  target_file="$( echo "salty0220/$file" | openssl md5 | sed -E 's/(........)(....)(....)(....)(............)/\1-\2-\3-\4-\5/' )"
  content_path="$target_dir/$target_file.md"

  log DEBUG "importing $file as $target_file ($ordering/$count)"
  
  touch "$content_path"

  markers=($( grep -n '^---$' "$content_path" | cut -f1 -d: ))
  awk "NR>${markers[0]:-0} && NR<${markers[1]:-0}" "$content_path" > "$content_path.yaml"
  awk "NR>${markers[1]:-0}" "$content_path" > "$content_path.body"

  [ -s "$content_path.yaml" ] || echo '{}' > "$content_path.yaml"

  patchyaml "$content_path.yaml" <(
    exif-read-tool -json -filepath "$source_dir/$file" \
      | grep -v ^WARNING \
      | jq 'map({ "key": "\(.fq_ifd_path)/\(.tag_name)", value }) | from_entries' \
      | jq -c \
          --arg ordering "$ordering" \
          --arg title "$file" \
          ' . as $dot |
            [
              { "path": "/title?", "type": "replace", "value": $title },
              { "path": "/ordering?", "type": "replace", "value": ( $ordering | tonumber ) },
              { "path": "/date?" } + ( .["IFD/DateTime"] | if . then { "type": "replace", "value": ( . | split(" ") | [ ( .[0] | gsub(":"; "-") ) , .[1] ] | join("T") ) } else { "type": "remove" } end ),
              { "path": "/exif?/make" } + ( .["IFD/Make"] | if . then { "type": "replace", "value": . } else { "type": "remove" } end ),
              { "path": "/exif?/model" } + ( .["IFD/Model"] | if . then { "type": "replace", "value": . } else { "type": "remove" } end ),
              { "path": "/exif?/iso" } + ( .["IFD/Exif/ISOSpeedRatings"] | if . then { "type": "replace", "value": .[0] } else { "type": "remove" } end ),
              { "path": "/exif?/aperture" } + ( .["IFD/Exif/ApertureValue"] | if . then { "type": "replace", "value": ( .[0] | ( .Numerator / .Denominator ) * 100 | round / 100 ) } else { "type": "remove" } end ),
              { "path": "/location?/latitude" } + ( .["IFD/GPSInfo/GPSLatitude"] | if . then { "type": "replace", "value": ( map( .Numerator / .Denominator ) | ( .[0] + .[1] / 60 + .[2] / 3600 ) * ( if $dot["IFD/GPSInfo/GPSLatitudeRef"] == "S" then -1 else 1 end )  ) } else { "type": "remove" } end ),
              { "path": "/location?/longitude" } + ( .["IFD/GPSInfo/GPSLongitude"] | if . then { "type": "replace", "value": ( map( .Numerator / .Denominator ) | ( .[0] + .[1] / 60 + .[2] / 3600 ) * ( if $dot["IFD/GPSInfo/GPSLongitudeRef"] == "W" then -1 else 1 end )  ) } else { "type": "remove" } end )
            ]
          '
  )

  lastDate="$( bosh interpolate --path=/date "$content_path.yaml" )"
  [[ -n "$firstDate" ]] || firstDate="$lastDate"

  resize "$content_path.yaml" "$source_dir/$file" "$assets_dir/$target_file~200x200.jpg" 200x200 -resize 200x200^ -gravity center -crop 200x200+0+0
  resize "$content_path.yaml" "$source_dir/$file" "$assets_dir/$target_file~420x420.jpg" 420x420 -resize 420x420^ -gravity center -crop 420x420+0+0
  resize "$content_path.yaml" "$source_dir/$file" "$assets_dir/$target_file~640w.jpg" 640w -resize 640
  resize "$content_path.yaml" "$source_dir/$file" "$assets_dir/$target_file~1080.jpg" 1080 -resize 1080x1080
  resize "$content_path.yaml" "$source_dir/$file" "$assets_dir/$target_file~1280.jpg" 1280 -resize 1280x1280
  resize "$content_path.yaml" "$source_dir/$file" "$assets_dir/$target_file~1920.jpg" 1920 -resize 1920x1920

  (
    echo ---
    bosh interpolate <( sed 's/HACKYHACKYPREFIX//' "$content_path.yaml" )
    echo ---
    cat "$content_path.body"
  ) > "$content_path"

  rm "$content_path.yaml" "$content_path.body"

  log INFO "imported $file as $target_file ($ordering/$count)"
done < <(
  log DEBUG sorting $count files

  (
    for file in $( cd "$source_dir" ; ls ); do
      echo -n "$file;"
      exif-read-tool -json -filepath "$source_dir/$file" \
        | grep -v ^WARNING \
        | jq -r 'map(select(.fq_ifd_path == "IFD" and .tag_name == "DateTime"))[0].value_string // "unknown"'
    done
  ) \
    | sort -t';' -k2

  log INFO sorted $count files
)

content_path="$target_dir/_index.md"

touch "$content_path"

markers=($( grep -n '^---$' "$content_path" | cut -f1 -d: ))
awk "NR>${markers[0]:-0} && NR<${markers[1]:-0}" "$content_path" > "$content_path.yaml"
awk "NR>${markers[1]:-0}" "$content_path" > "$content_path.body"

[ -s "$content_path.yaml" ] || echo '{}' > "$content_path.yaml"

patchyaml "$content_path.yaml" <(
    jq -cn \
      --arg firstDate "$firstDate" \
      --arg lastDate "$lastDate" \
      '
        [
          { "path": "/date?", "type": "replace", "value": ( $lastDate | split("T")[0] )  },
          { "path": "/date_start?", "type": "replace", "value": ( $firstDate | split("T")[0] ) }
        ]
      '
  )

(
  echo ---
  cat "$content_path.yaml"
  echo ---
  cat "$content_path.body"
) > "$content_path"

rm "$content_path.yaml" "$content_path.body"

log DEBUG uploading assets

aws s3 sync --cache-control 'public,max-age=2592000' "$assets_dir" "s3://dpb587-website-us-east-1/$assets_dir"

log INFO uploaded assets

log INFO finished import
