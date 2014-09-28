Publish the site:

    git commit ...snip...
    ./_build/build.sh ./_build/aws/publish.sh "$AWS_S3CMD_CONFIG"

Sync through new blog assets:

    for NAME in $(ls asset/blog/) ; do ./_build/aws/publish-asset.sh "$AWS_S3CMD_CONFIG" blog/$NAME ; done

Re-generating galleries...

    # 2014-london-iceland-trip
    $ osascript ../jekyll-gallery/export-iphoto.applescript 'London-Iceland Trip' \
      | php -dmemory_limit=1G ../jekyll-gallery/convert.php 2014-london-iceland-trip \
        --export 96x96 \
        --export 200x200 \
        --export 640w \
        --export 1280
    # 2014-colorado-aspens
    $ osascript ../jekyll-gallery/export-iphoto.applescript 'Colorado Aspens' \
      | php ../jekyll-gallery/convert.php 2014-colorado-aspens \
        --export 200x200 \
        --export 640w \
        --export 1280

Uploading photo galleries:

    ./_build/aws/publish-asset.sh "$AWS_S3CMD_CONFIG" gallery/2014-london-iceland-trip
