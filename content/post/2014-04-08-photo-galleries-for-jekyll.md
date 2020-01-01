---
date: 2014-04-08
title: Photo Galleries for Jekyll
description: Easily exporting my iPhoto album to this Jekyll-based site.
tags:
- blog
- gallery
- iphoto
- jekyll
- jekyllrb
- photo
- ruby
aliases:
- /blog/2014/04/08/photo-galleries-for-jekyll.html
---

I had a trip to London and Iceland several weeks ago, and I wanted to share some of those photos with people. In the
past I've put those sorts of photo galleries on Facebook, but some friends don't have accounts there and I figured I
could/should just keep my photos with my other personal stuff here.

Unlike [WordPress][1], [Jekyll][2] doesn't really have a concept of photo galleries, and since Jekyll is a static site
generator it makes things a little more difficult. I looked through [several][3] [other][4] [posts][5] discussing Jekyll
photo galleries, but they all seemed a bit more primitive than what I wanted. I wanted to:

 * stick with existing Jekyll paradigms (e.g. [markdown][8] file to static page),
 * retain metadata about my photos (e.g. location data, camera EXIF data),
 * support multiple views about my galleries (e.g. photo list, map, slideshow),
 * ensure photos can have landing pages and be easily navigated, and
 * avoid committing images to my git repository.

After giving it some thought, I realized this was going to be a multi-step process.

 0. Script the process of exporting my existing photos to Jekyll-friendly structures.
 0. Find a Jekyll/[Liquid][7] plugin to enumerate directories/files and use the results.
 0. Create templates and pages for my gallery and its photos.
 0. Publish the site!


## Step 1: Export existing photo galleries (iPhoto)

I take pretty much all my photos with my phone and those photos then get synced up with iPhoto. At the end of my trip, I
browse through the photos and create an album of interesting ones. Normally I don't go through and give every photo a
title and description, but if I'm planning on sharing them I add brief notes within iPhoto.

I knew my iPhoto metadata was stored in `AlbumData.xml`, but I've always had poor performance with massive XML data
files. I decided to start with a different approach: [AppleScript][9]. The following snippet gets me the file paths of
all the photos (in order) from whatever album I ask for:

```applescript
on run argv
    set output to ""

    tell application "iPhoto"
        set vAlbum to first item of (get every album whose name is (item 1 of argv))
        set vPhotos to get every photo in vAlbum
        
        repeat with vPhoto in vPhotos
            set output to output & original path of vPhoto & "
"
        end repeat
    end tell
        
    return output
end run
```

So, to get the photos in my album named "London-Iceland Trip" I can do:

```
$ osascript export-iphoto-album.applescript 'London-Iceland Trip'
~/Pictures/iPhoto Library.photolibrary/Masters/2014/03/13/20140313-154842/IMG_0303.JPG
~/Pictures/iPhoto Library.photolibrary/Masters/2014/03/13/20140313-154842/IMG_0308.JPG
...snip...
```

With some tweaks I can get more than just the path to a photo:

```
$ osascript export-iphoto-album.applescript 'London-Iceland Trip'
altitude: 16
latitude: 51.50038
longitude: -0.12786667
name: A Classic View
date: Thursday, March 6, 2014 at 4:44:12 PM
path: ~/Pictures/iPhoto Library.photolibrary/Masters/2014/03/13/20140313-154842/IMG_0303.JPG
title: A Classic View
------
QCon was held at The Queen Elizabeth II Conference Centre and this was the view out one of the common areas.
------------
...snip...
```

The next piece is to write something which will clean up the output, resize the photos, and write out all the different
Jekyll files. For that I created a [PHP][10] script since it was going to be easiest for me. Once complete, I then just
pipe the export results to the script and specify the image sizes I want:

```
$ osascript ../jekyll-gallery/export-iphoto.applescript 'London-Iceland Trip' | \
  php ../jekyll-gallery/convert.php 2014-london-iceland-trip \
  --export 96x96 --export 200x200 --export 640 --export 1280
df5150c-a-classic-view...96x96...200x200...640...1280...mdown...done
7cf02b5-night...96x96...200x200...640...1280...mdown...done
...snip...
```

Once complete, all the resized images are in `asset/gallery/2014-london-iceland-trip` and my markdown files with the
photo details are in `gallery/2014-london-iceland-trip` and they're easily [readable][15].


## Step 2: Jekyll plugin

At a minimum, I wanted to have a listing of all the photos in a gallery index page. After some searches, I found
[two][11] [scripts][12] which became the inspiration for my final plugin. My [final plugin][16] looks like:

    Tag:
      loopdir
    Attributes:
      match: a pattern to match files within the path (e.g. "*.md")
      parse: whether to load the file and parse for YAML front matter
      path: a directory, relative to the site root, to find files
      sort: a property to search by (e.g. "path")
    Result:
      An "item" object is exposed to the template with a "page"-like structure.
      If parsing is enabled, the YAML properties are available as "item.title".

Which means I can easily compose a simple photo list with:

```jinja
{% loopdir path:"gallery/2014-london-iceland-trip" match:"*.md" sort:"ordering" %}
    <a href="/{{ item.fullname }}.html">
        <img alt="Photo: {{ item.title }}" height="200" src="/{{ item.fullname }}~200x200.jpg" title="{{ item.title }}" width="200" />
    </a>
{% endloopdir %}
```

I reuse this plugin elsewhere for regular directory listings.


## Step 3: Create templates

I've started out with two reusable templates in my `_includes` directory:

 0. [Gallery List][13] - a simple listing of thumbnails from all the photos in the gallery
 0. [Interactive Map][14] - an interactive map showing where all the photos were taken

I can pass arguments (like the gallery name) to the include which makes it easy to embed a gallery in any page:

```jinja
{% include gallery_list.html gallery='2014-london-iceland-trip' %}
```


## Step 4: Publish

After generating everything locally, I just have to do a couple steps:

 0. Commit all the new `gallery/2014-london-iceland-trip` files (and new templates)
 0. Run `_build/aws/publish-asset.sh $AWS_S3CMD_CONFIG gallery/2014-london-iceland-trip` to upload all the exported JPGs
 0. Run `_build/aws/build.sh _build/aws/publish.sh $AWS_S3CMD_CONFIG` to upload any modifications from the rest of the
    site

To make things easier for myself and, possibly, others I put the conversion scripts in my [jekyll-gallery][17] repo.

Now I'm able to refer people to the [gallery](/gallery/2014-london-iceland-trip/) or embed the gallery somewhere
useful...

<div style="line-height:0;padding:4px 0 0 1px;">
  {% loopdir path:"gallery/2014-london-iceland-trip" match:"*.md" sort:"ordering" %}<a href="/{{ item.fullname }}.html" style="display:inline-block;margin:3px;text-decoration:none;"><img alt="Photo: {{ item.title }}" height="48" src="{{ site.asset_prefix }}/{{ item.fullname }}~96x96.jpg" title="{{ item.title }}" width="48" style="padding:1px;" /></a>{% endloopdir %}
</div>



 [1]: http://wordpress.org/
 [2]: http://jekyllrb.com/
 [3]: https://github.com/ggreer/jekyll-gallery-generator
 [4]: http://www.mgratzer.com/from-wordpress-to-jekyll/
 [5]: https://github.com/tsmango/jekyll_flickr_set_tag
 [6]: https://help.github.com/articles/what-are-github-pages
 [7]: http://liquidmarkup.org/
 [8]: http://daringfireball.net/projects/markdown/
 [9]: https://developer.apple.com/library/mac/documentation/applescript/Conceptual/AppleScriptX/AppleScriptX.html
 [10]: http://www.php.net/
 [11]: https://gist.github.com/jgatjens/8925165
 [12]: http://simon.heimlicher.com/articles/2012/02/01/jekyll-directory-listing
 [13]: https://github.com/dpb587/dpb587.me/blob/master/_includes/gallery_list.html
 [14]: https://github.com/dpb587/dpb587.me/blob/master/_includes/gallery_map.html
 [15]: https://github.com/dpb587/dpb587.me/blob/master/gallery/2014-london-iceland-trip/df5150c-a-classic-view.md
 [16]: https://github.com/dpb587/dpb587.me/blob/master/_plugins/loopdir.rb
 [17]: https://github.com/dpb587/jekyll-gallery
