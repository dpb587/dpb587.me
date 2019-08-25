---
'@context': http://schema.org
'@type': BlogPosting
datePublished: "2017-10-09"
description: A general format for documenting checksums, signatures, and origins of
  blobs.
keywords:
- blobs
- checksum
- dependencies
- golang
- meta4
- metalink
- signatures
name: Documenting Blobs with Metalink Files
url:
- /blog/2017/10/09/documenting-blobs-with-metalink-files.html
---

There are many blobs around the web, with different organizations and teams publishing artifacts through different channels and with varying security. Often a single project will have many dependencies from multiple different sources, and developers need to know specifics about where to download blobs and how to verify them. I started looking for a solution to help unify the way I was both consuming and sharing blobs across my own projects. Eventually I found my way to something called "metalink" files, and it has become a very useful method for managing blobs (and resources) in my projects.

<!--more-->


## About Metalinks

Briefly, [metalinks][1] are a file format for documenting content digests, signing details, and download sources of other files. The Metalink Download Description Format RFC ([RFC 5484][2], June 2010) describes the core structure of the XML file. As an example, here is a simple metalink file which documents a blob which can be downloaded via FTP or HTTP and can be verified with a SHA256 checksum.

```xml
<metalink xmlns="urn:ietf:params:xml:ns:metalink">
  <published>2009-05-15T12:23:23Z</published>
  <file name="example.ext">
    <size>14471447</size>
    <version>1.0</version>
    <hash type="sha-256">f0ad929cd259957e160ea442eb80986b5f01...</hash>
    <url priority="1">ftp://ftp.example.com/example.ext</url>
    <url priority="10">http://example.com/example.ext</url>
  </file>
</metalink>
```

For me, I typically find the hash/checksum/digest and the download URLs to be the most useful pieces of consolidated information - they are often the most difficult to automatically consume from upstream sources. Typically, sites will publish a downloadable link to the file and alongside it will be some sort of checksum within the HTML or as a sibling file. Rarely are both values provided as part of a consistent API (for example, the GitHub API provides download URLs but does not currently provide checksum information).

One drawback of metalink files is that they are not widely used (the RFC was never accepted TODO). After discovering the metalink RFC I noticed a few utilities which provided metalinks (like [curl][5]) and a few tools which supported downloading from metalink files (like [curl][5] and [aria2c][7]). While widespread use would be an encouraging reason to adopt the specification, it doesn't discourage me from using it since it documents the values I care about and I haven't found a better solution yet.

Another large drawback of metalink files is that they are XML which means they are more difficult to automate and write simple scripts for. To help hide the inconvenience of the format, I created a [golang][7] utility in [dpb587/metalink][6] which can be used to create, read, update, and use metalink files.


## Creating Metalinks

From a release maintainer perspective, one common task when releasing is to publish a list of binaries. Here's an example using the `meta4` binary from the `dpb587/metalink`. Before getting started, there are a couple useful variables to set which we'll use...

```bash
version=1.2.3                 # the release version
xmlpath=build/v$version.meta4 # the metalink file
```

The first step would be to create an empty metalink file that we will then populate with our binaries (and we'll also set a publication date since we happen to know it)...

```bash
meta4 create --metalink="$xmlpath"
meta4 set-published --metalink="$xmlpath" "$( date -u +%Y-%m-%dT%H:%M:%SZ )"
```

Next, we can use the `import-file` command for every file built as part of the release. The command will automatically add the file to the metalink, generate checksums, record the file size, and set the version.

```bash
for file in $( find build -type f -maxdepth 1 | cut -d/ -f2 | sort ); do
  meta4 import-file --metalink="$xmlpath" --file="$file" --version="$version" "build/$file"
done
```

At this point, we have a metalink file which has a list of all the release artifacts and their checksums, but it does not provide any URLs for where a user could download the blobs. If part of our release process will upload those files later, we could use the `file-set-url` command to record a URL for the file. For example, if a GitHub release will host the files, we could do the following after `import-file`...

```bash
meta4 file-set-url --metalink="$xmlpath" --file="$file" "https://github.com/example/repo/releases/v$version/downloads/$file"
```

Alternatively, if we will be self-hosting the objects we can let `meta4` take care of uploading with the `file-upload` command. For example, to upload to an S3 bucket we could to the following after `import-file` (assuming the `AWS_*`authentication environment variables were already configured)...

```bash
meta4 file-upload --metalink="$xmlpath" --file="$file" "build/$file" "s3://s3-external-1.amazonaws.com/example-repo-us-east-1/releases/v$version/$file"
```

Or, if publishing to both an S3 bucket and GitHub releases, use both commands to list both URLs and ensure users can download the blob if one of the services is unavailable.

Now the metalink file contains checksums and download sources, so it is ready to be published or committed somewhere for external users.

For a more complete example, see the [build][8] step of my [ssoca][10] project.


## Publishing Metalinks

Once a metalink file is created, you will usually want to share it for others to reference. In my own uses, I have been committing the metalink files to a git repository because it can provide strong audit and integrity features. Depending on the project, I have been committing them to a couple different locations.

If the project already contains build metadata for releases (e.g. [BOSH][11] releases), I'll commit the metalink file alongside the other build metadata. This helps keep the build information in an expected location.

Otherwise, or if trying to use multiple release channels, I'll commit the metalinks to a separate branch (e.g. `artifacts`) within a sub-directory to represent the channel. For example, during CI there may be a couple channels of artifacts: first a `dev` build which has not been fully tested, then perhaps a `rc` build once all tests are successful, and then perhaps a `final` build once an official version has been published. By using separate directories to represent status, consumers can refer to the directory which is most appropriate for their needs.

For a more complete example, see the [build pipeline][12] of my [ssoca][10] project which publishes the binaries at a couple stages.


## Consuming Metalinks

The metalink file itself isn't that interesting - it's the information inside it which is useful. There are a few different use cases which might be useful.


### Downloading Files

The most common use of metalinks is for users to be able to download the files referenced by the metalink file. The `meta4` tool offers a simple `file-download` command which can be used to download and verify files from a working mirror...

```bash
meta4 file-download --metalink-file=$xmlpath --file=$file ~/Downloads/$file
```

For users familiar with `aria2c`, they could use the command directly...

```bash
aria2c --metalink-file=$xmlpath
```

For users of `curl` with metalink support...

```bash
curl --metalink=$xmlpath
```


### Extracting Checksums

When publishing a release, authors will often provide checksums alongside the binaries to make it easy for users who download files with traditional methods. Rather than re-run checksum commands, we can extract that information from our existing metalink file. For example, the following could be used to append release notes with the SHA256 checksum for all release binaries...

```bash
for file in $( meta4 files --metalink=$xmlpath ); do
  echo "$( meta4 file-hash --metalink=$xmlpath --file=$file sha-256 )" "$file"
done
```

For a more complete example, see the [build-release][9] step of my [ssoca][10] project.


### Repository Querying

Assuming you have committed the metalinks to a shared location, a natural next step is to treat the metalinks as a lightweight database. To help with that, I created a `meta4-repo` tool with a few commands for enumerating lists of metalinks. For example, to query metalinks stored in a local directory and show all versions available I could do...

```
$ meta4-repo filter --format=version -n8 file:///tmp/nginx-metalinks/
1.13.4
1.13.3
1.13.2
1.13.1
1.13.0
1.12.0
1.11.13
1.11.12
```

Or, if I have a specific version constraint of using `1.11.*` I could do...

```
$ meta4-repo filter --format=version --filter fileversion:~1.11 -n4 file:///tmp/nginx-metalinks/
1.11.13
1.11.12
1.11.11
1.11.10
```

With the assumption that one metalink represents a single version, I could then export the matching metalink file and use it for downloading the files.

```bash
meta4-repo show --filter fileversion:1.11.13 file:///tmp/nginx-metalinks > metalink.meta4
meta4 file-download nginx.tgz
```

If the metalinks are stored in a git repository, a `git` scheme can be used as well. For example...

```
$ meta4-repo filter --format=version --filter fileversion:~1.11 -n2 git+https://github.com/dpb587/upstream-blob-receipts.git//repository/nginx
1.11.13
1.11.12
```

The repository querying approach becomes useful when sharing assets across environments or projects. If you use [Concourse][13] for CI/CD, you may want to try the [dpb587/metalink-repository-resource][14] which allows you to securely download artifacts from arbitrary locations for your build tasks. By supporting arbitrary repositories, it becomes trivial to add intermediate repositories to represent custom stability channels within a team or organization. For example, a development environment can consume from an official upstream repository, then after development sign-off the metalink is promoted to a different repository that staging or production is watching.


### Programmatic Usage

For more advanced uses, the `github.com/dpb587/metalink` package can be imported to a golang project to easily parse and consume metalink files. For example, the following will generate the same file sha256+name list as the previous example.

```go
package main

import "github.com/dpb587/metalink"

func main() {
  var meta4 metalink.Metalink

  meta4bytes, err := ioutil.ReadFile(os.Args[1])
  if err != nil {  panic(err) }

  err = metalink.Unmarshal(meta4bytes, &meta4)
  if err != nil {  panic(err) }

  for _, file := range meta4.Files {
    for _, hash := range file.Hashes {
      if hash.Type != "sha-256" { continue }

      fmt.Printf("%s %s\n", hash.Hash, file.Name)
    }
  }
}
```


[1]: http://www.metalinker.org/
[2]: https://tools.ietf.org/html/rfc5854
[3]: https://tools.ietf.org/html/rfc6249
[4]: https://stedolan.github.io/jq/
[5]: https://curl.haxx.se/metalink.cgi?curl=tar.gz
[6]: https://github.com/dpb587/metalink
[6]: https://golang.org/
[7]: https://aria2.github.io/
[8]: https://github.com/dpb587/ssoca/blob/1a48a8700f320105b85ea2b1007af3d4053f60e8/ci/tasks/build/execute.sh#L19
[9]: https://github.com/dpb587/ssoca/blob/1a48a8700f320105b85ea2b1007af3d4053f60e8/ci/tasks/build-release/execute.sh#L27
[10]: https://github.com/dpb587/ssoca
[11]: https://bosh.io/
[12]: https://github.com/dpb587/ssoca/blob/1a48a8700f320105b85ea2b1007af3d4053f60e8/ci/pipeline.yml
[13]: https://concourse.ci/
[14]: https://github.com/dpb587/metalink-repository-resource
