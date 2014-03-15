---
title: A Generic Storage Interface
layout: post
tags: asset php storage
description: Abstracting file storage, whether it's local or cloud.
---

Websites often have a lot of different assets and files for the various areas of a website - content management systems,
photo galleries, e-commerce product photos, etc. As a site grows, so does storage demand and backup requirements, and as
storage demands grow it typically becomes necessary to distribute those files across multiple servers or services.

One method for managing disparate file systems is to use custom PHP [stream wrappers][4] and configurable paths; but
some extensions don't yet support custom wrappers for file access. An alternative that I've been using is an object and
service-oriented approach to keep my application code independent from the storage configuration.


## Interface

At the core of my design, is the asset storage interface which looks something like:

{% highlight php %}
<?php interface StorageEngineInterface {

    // store a file and return back a token that can be used to retrieve it
    function store(SplFileInfo $file);

    // retrieve a locally-accessible SplFileInfo based on the token
    function retrieve($token);

    // remove data from storage based on the token
    function purge($token);

}
{% endhighlight %}

The storage engine is responsible for generating a reusable token that can be used for later retrieval. Generally, I
simply have it generate a UUID as the token, however tokens could have storage-specific meaning.


## Sample Storage Engines

I've used several base implementations:

 * `LocalStorageEngine` - the simplest storage using a local/NFS filesystem
 * `AWSS3StorageEngine` - using [AWS S3][1] for storage
 * `SftpStorageEngine` - using PHP's [ssh2][2] module to access files on servers via SFTP
 * `AtlassianConfluenceStorageEngine` - managing documents within [Confluence][3] wikis

Remote services like AWS S3 and SFTP can cause significant performance issues. To help with that, I use a
`CachedStorageEngine` implementation. It accepts two `StorageEngineInterface` arguments: one as the upstream engine, and
one as the local cache. For example:

{% highlight php %}
<?php

new CachedStorageEngine(
    new AWSS3StorageEngine(new Aws\S3\S3Client(...), 'bucket.example.com', 'my-prefix'),
    new LocalStorageEngine('/tmp/s3-bucket.example.com-cache')
);
{% endhighlight %}

And since `CachedStorageEngine` is just another implementation of `StorageEngineInterface`, it can be used
interchangeably within the application with performance being the only difference.


## Application Usage

Using dependency injection, each of the storage backends becomes an independent service, configured depending on the
application requirements. The application then has no storage-specific calls like `copy`, `file_get_contents`, `fopen`,
etc and the code looks something like:

{% highlight php %}
<?php

// storage service for photos
$storage = $dic->get('photo_storage')

// save a new photo
$photo = new PhotoRecord();
$photo->setAssetToken(
    $storage->store($request->files->get('upload'))
);

// use the photo
$image = (new Imagine\Gd\Imagine())->open(
    $storage->retrieve($photo->getAssetToken())
);

// delete the photo
$storage->purge($photo->getAssetToken());
$photo->delete();
{% endhighlight %}

Since `retrieve` will always return a [`SplFileInfo`][5] instance, it can be referenced and handled like a local file
(as demonstrated by the `open` call in the example.


## Complicating Things

The asset storage interface itself is fairly primitive, but it allows for some more complex configurations:

 * by using dependency injection, it becomes extremely easy to switch storage engines since application code doesn't
   need to change
 * complex storage rules can be combined with meaningful tokens to, for example, store very large files on different
   disks and using a token prefix to identify that class
 * creating a fallback storage class which will go through a chain of storages searching until it's able to store or
   retrieve a token
 * internally deferring operations via queue manager (e.g. instead of storing files immediately to S3 and waiting for
   upload time, write it locally and create a job to upload it in the background)


## Summary

By abstracting storage logic outside of my application code, it makes my life much more easier as a developer and as a
systems administrator when trying to manage where files are located and any relocations, as necessary.


 [1]: http://aws.amazon.com/s3/
 [2]: http://www.php.net/manual/en/book.ssh2.php
 [3]: http://atlassian.com/software/confluence/overview/team-collaboration-software
 [4]: http://www.php.net/manual/en/class.streamwrapper.php
 [5]: http://us.php.net/manual/en/class.splfileinfo.php
