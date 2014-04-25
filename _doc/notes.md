Publish the site:

    git commit ...snip...
    ./_build/build.sh ./_build/aws/publish.sh "$AWS_S3CMD_CONFIG"

Create a photo gallery:

    osascript enumerate.scpt 'London-Iceland Trip' | php -dmemory_limit=1G convert.php ~/code/dpb587.me/ 2014-london-iceland-trip
    ./_build/aws/publish-asset.sh "$AWS_S3CMD_CONFIG" gallery/2014-london-iceland-trip

Sync through new blog assets:

    for NAME in $(ls asset/blog/) ; do ./_build/aws/publish-asset.sh "$AWS_S3CMD_CONFIG" blog/$NAME ; done
