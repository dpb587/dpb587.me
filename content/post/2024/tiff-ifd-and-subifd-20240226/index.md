---
description: A technical investigation about some internals of TIFF files.
publishDate: "2024-02-26"
title: "TIFF: IFD and SubIFD"
---

TIFF is a popular file format for professional images. One of the key features is being able to include multiple images within a single file. That does make it a little more complicated to write libraries for image processing, though. While working on some custom image services, I did a deep-dive into the TIFF specifications and practical usage of the file format. After quite some research and exploration, I've tried to summarize my learnings about TIFF and its multi-image features in this post.

* **[Image File Directory](#ifd)** -- the internal structure which supports multiple image
* **[Tags](#tags)** -- how metadata is used to describe content and images
* **[SubIFD](#subifd)** -- another method for organizing and relating IFDs to one another
* **[Real-world Usage](#usage)** -- examples of multi-image, pyramid, and industry-specific TIFF files
* **[General Notes](#general-notes)** -- a few brief takeaways about TIFF subfiles
* **[Resources](#resources)** -- links to official standards and related documentation

# Image File Directory {#ifd}

Every TIFF file contains at least one *Image File Directory* (IFD) -- a section within the file to hold the bytes of an encoded image along with its metadata. But, to be clear with the name, it's not a separate file nor directory on the filesystem. An IFD is also commonly called a *subfile* which is probably a better way to think about it.

I'll be using conceptual diagrams in this post to help visualize some of the topics. To start, a single-image TIFF file has a very basic structure with a single IFD.

{{< generated-diagram lang="plantuml" dir="single-image-plantuml" caption="Diagram of a basic TIFF file with a single image." >}}

Some industries and use cases require multiple images within a single file. Since an IFD is a container to hold all the data needed for an individual image, an application can basically just write additional IFDs into a TIFF file.

There are a couple ways IFDs can be used. As the main images of a file, an extra piece of metadata, *Next IFD*, is included which tells applications where it can find the next IFD to read. This sequence of IFDs is called the *main-IFD chain*, and they are implicitly numbered based on the ordering, starting from zero.

{{< generated-diagram lang="plantuml" dir="multi-image-plantuml" caption="Diagram of a multi-image TIFF file." >}}

By following the main-IFD chain (highlighted in white on these diagrams), applications can quickly seek through a TIFF file to whichever image(s) it needs. Externally, if a caller wants to reference an image from this main-IFD chain, it will reference the IFD based on its sequence number. For example, the first and default image of a TIFF file is always IFD0, and the third may be referenced as IFD2.

It's worth noting that not all applications support TIFF files with multiple images (it is considered an optional feature of the specifications). In these cases, an application will typically stop reading the file after it reads the first image (IFD0).

# Tags {#tags}

A piece of metadata, such as Next IFD, is officially named a *tag*. These tags are recorded within the header of an IFD and act as a kind of key/values lookup to further describe the IFD. Official TIFF specifications define common tags which range in purpose from low-level details (such as dimensions and byte offsets) to user-facing properties (such as copyrights and descriptions). Below are some examples referenced in this post.

* **Next IFD**, mentioned previously, is what links an IFD with a subsequent IFD. Without the tag, applications will generally assume the IFD is the last in a file and stop processing. The value is the byte offset for the start of the target IFD within the file.
* **Subfile Type** helps classify the image with a few common markers:
  * Full-resolution - indicates the image is the fullest resolution available.
  * Reduced-resolution - indicates the image has been resized down from full-resolution (which, typically, can also be found in the file).
  * Single page of multiple - indicates the image is one of many and there are typically other sibling IFDs within the file.
  * And there are several other official and vendor-conventional values that might appear, such as Transparency mask.
* **Width** and **Height** (technically, *Length*) are generic tags that should always be present and describe the dimensions of an image (in pixels).
* **Page Number** may be present when there is a particular ordering to multiple pages. Although IFDs are ordered within the TIFF file, that ordering shouldn't be relied on for user-facing information.
* **Image Description** may be a user-supplied description about the subject of an image. Sometimes this is short and a literal description (e.g. company picnic), but other times this is used for application-specific metadata and may not be human-readable.

Below is a more detailed version of the `multipage.tif` diagram shown earlier, but with the addition of some conventional tags.

{{< generated-diagram lang="plantuml" dir="multi-page-plantuml" caption="Diagram of a multi-page TIFF file, including some tags." >}}

# SubIFD {#subifd}

Where the Next IFD tag creates an association to a sibling IFD for the file, the *SubIFD* tag creates an association from an IFD to a child IFD(s). Typically, the top-level/main IFDs are used for different but related images, while SubIFDs are used to add supporting or alternate representations of a specific image. A sequence of child IFDs is called a *SubIFD chain*, and they, too, are implicitly numbered based on their order within the chain, starting from zero.

Between the Next IFD and SubIFD tags, there are several internal layouts that a TIFF file may use for inter-related images. In practice, only one of these approaches is used within a given file, but, in theory, they are not mutually exclusive.

## TIFF Trees {#subifd-tiff-trees}

The standard approach, named *TIFF Trees* by the specification, uses the SubIFD tag with values pointing to one or more IFDs. In this case, an application can read the SubIFD tag once before seeking to any of the listed IFDs.

{{< generated-diagram lang="plantuml" dir="multi-ifd-with-subifd-plantuml" caption="Diagram of a TIFF file with multiple IFDs, each with several SubIFDs." >}}

## Next [Sub]IFD {#subifd-next-ifd}

Another approach relies on both the SubIFD and Next IFD tags. A chain is started with a standard SubIFD tag pointing to a child IFD; but, from there, the Next IFD tag is used to continue the chain. In this case, an application would read the SubIFD tag once, and then sequentially seek through child IFDs until the Next IFD tags stop (similar to how the main-IFD chain would be processed).

{{< generated-diagram lang="plantuml" dir="multi-ifd-with-subifd-chain-plantuml" caption="Diagram of a TIFF file with multiple IFDs, each with a secondary SubIFD chain." >}}

## Common IFD {#common-ifd}

Finally, consider the following example which demonstrates how tags are simply references to IFDs within the file. In the case where two IFDs need to refer to the exact same image, such as a transparency mask, the SubIFD tags may end up referring to a single IFD.

{{< generated-diagram lang="plantuml" dir="multi-ifd-subifd-shared-plantuml" caption="Diagram of a TIFF file with multiple IFDs, SubIFDs, and a common SubIFD." >}}

# Real-world Usage {#usage}

TIFF files, with their IFDs and SubIFDs, can be written and structured using a wide range of styles.

## Multi-page TIFF {#multi-page-tiff}

In *multi-page TIFF* files, independent but related images are encoded in a single file. For example, a camera capturing a 'burst' of images, or a document scanner digitizing the front and back of a page. The structure of the file will typically be a basic main-IFD chain.

{{< generated-diagram lang="plantuml" dir="multi-page-sequential-plantuml" caption="Diagram of a multi-page TIFF file with sequential pages." >}}

## Pyramid TIFF {#pyramid-tiff}

In a *Pyramid TIFF*, the first IFD is a full-resolution image with additional IFDs for smaller versions of the image. Pyramid files, in general, are used for high-performance publishing environments since the extra, downscaled images avoid slow resize operations. Pyramid-aware use the Subfile Type, Width, and Height tags to identify which IFD is most appropriate.

{{< generated-diagram lang="plantuml" dir="pyramid-plantuml" caption="Diagram of a Pyramid TIFF file with multiple resolutions." >}}

{{< details summary="Command Line Example" >}}

  {{< markdown >}}

    The following commands create and then inspect a Pyramid TIFF using a [large image from NASA](https://mars.nasa.gov/resources/27612/curiosity-views-mud-cracks-in-the-clay-sulfate-transition-region/) (29163×8162).

  {{< /markdown >}}

  {{< terminal >}}

    {{< terminal-input >}}

      ```
      vips tiffsave PIA25915.png PIA25915-pyramid.tif --pyramid
      ```

    {{< /terminal-input >}}

    {{< terminal-input >}}

      ```
      tiffinfo PIA25915-pyramid.tif | grep -e '^=' -e '^-' -e 'Image Width:'
      ```

    {{< /terminal-input >}}

    {{< terminal-output >}}

      ```
      === TIFF directory 0 ===
        Image Width: 29163 Image Length: 8162
      === TIFF directory 1 ===
        Image Width: 14581 Image Length: 4081
      === TIFF directory 2 ===
        Image Width: 7290 Image Length: 2040
      === TIFF directory 3 ===
        Image Width: 3645 Image Length: 1020
      === TIFF directory 4 ===
        Image Width: 1822 Image Length: 510
      === TIFF directory 5 ===
        Image Width: 911 Image Length: 255
      === TIFF directory 6 ===
        Image Width: 455 Image Length: 127
      === TIFF directory 7 ===
        Image Width: 227 Image Length: 63
      === TIFF directory 8 ===
        Image Width: 113 Image Length: 31
      ```

    {{< /terminal-output >}}

  {{< /terminal >}}

{{< /details >}}

For single-image files, pyramids are often written using all top-level IFDs as shown in the previous diagram. However, they can also be structured using the SubIFD tag.

{{< generated-diagram lang="plantuml" dir="pyramid-subifd-plantuml" caption="Diagram of a Pyramid TIFF file using SubIFDs for subsampled resolutions." >}}

{{< details summary="Command Line Example" >}}

  {{< markdown >}}

    The following commands create and then inspect a Pyramid TIFF using a [large image from NASA](https://mars.nasa.gov/resources/27612/curiosity-views-mud-cracks-in-the-clay-sulfate-transition-region/) (29163×8162). In this example, the generated TIFF uses SubIFDs for the reduced-resolution images.

  {{< /markdown >}}

  {{< terminal >}}

    {{< terminal-input >}}

      ```
      vips tiffsave PIA25915.png PIA25915-pyrsub.tif --pyramid --subifd
      ```

    {{< /terminal-input >}}

    {{< terminal-input >}}

      ```
      tiffinfo PIA25915-pyrsub.tif | grep -e '^=' -e '^-' -e 'Image Width:'
      ```

    {{< /terminal-input >}}

    {{< terminal-output >}}

      ```
      === TIFF directory 0 ===
        Image Width: 29163 Image Length: 8162
      --- SubIFD image descriptor tag within TIFF directory 0 with array of 8 SubIFD chains ---
      --- SubIFD 0 of chain 0 at offset 0x3571630c (896623372):
        Image Width: 14581 Image Length: 4081
      --- SubIFD 0 of chain 1 at offset 0x381dc55e (941475166):
        Image Width: 7290 Image Length: 2040
      --- SubIFD 0 of chain 2 at offset 0x38cbe790 (952887184):
        Image Width: 3645 Image Length: 1020
      --- SubIFD 0 of chain 3 at offset 0x38f8f9d2 (955840978):
        Image Width: 1822 Image Length: 510
      --- SubIFD 0 of chain 4 at offset 0x3905080c (956631052):
        Image Width: 911 Image Length: 255
      --- SubIFD 0 of chain 5 at offset 0x3908153e (956831038):
        Image Width: 455 Image Length: 127
      --- SubIFD 0 of chain 6 at offset 0x3909a228 (956932648):
        Image Width: 227 Image Length: 63
      --- SubIFD 0 of chain 7 at offset 0x390a6f02 (956985090):
        Image Width: 113 Image Length: 31
      ```

    {{< /terminal-output >}}

  {{< /terminal >}}

{{< /details >}}

This SubIFD approach enables multi-page TIFF files to include pyramid-style, downscaled versions for each page, too.

## OME-TIFF {#ome-tiff}

An *[OME-TIFF](https://docs.openmicroscopy.org/ome-model/5.6.3/ome-tiff/specification.html) file* is a good example of a domain-specific standard built on top of TIFF standards. The Open Microscopy Environment (OME) is a consortium of groups focused on standards for microscopy data -- including images. In these cases, images are captured across several dimensional planes and a single TIFF file can include all of them via IFDs. However, the standard TIFF tags are insufficient for describing microscopy metadata of the IFDs, so [OME-XML](https://www.openmicroscopy.org/Schemas/OME/index.html) data is also embedded in the TIFF file using the *Image Description* tag.

{{< generated-diagram lang="plantuml" dir="ome-tiff-plantuml" caption="Diagram of a TIFF file using OME-TIFF conventions." >}}

# General Notes {#general-notes}

As a reminder, IFD-related tags are simply pointers to IFD locations within a TIFF file. Aside from recommendations, there is no technical enforcement about the kind of structure a TIFF file results in. When building libraries or lower-level tools, there are a couple things to keep in mind:

1. It's important to understand, follow, and test any industry-specific conventions that may be used on top of TIFF. For example, usage of main-IFD vs SubIFD chains. Otherwise, applications may not be able to correctly enumerate TIFF subfiles.
1. When working with untrusted files, it's even more important to avoid assumptions about how TIFF files and their IFD chains may be structured. In theory, those chains could be crafted to cause recursive loops, out of bound reads, or unexpected traversals.

One of my initial reasons for investigating was to figure out a canonical and practical way to reference images within TIFFs. It ended up more complicated than I expected due to the potential combinations of Next IFD + SubIFD tags used for traversal. For now, I decided to stay with `ifd=N,subifd=M`-style references and only support access to main IFDs and their direct, SubIFD tag-based children. It keeps things simple, matches most other image processing libraries, and I think it is inline with changes that might be needed later for more comprehensive IFD access.

# Resources {#resources}

If you want to learn more, I found the following links most helpful.

* [Baseline TIFF](https://www.awaresystems.be/imaging/tiff/specification/TIFF6.pdf) specification about the file format, standard tags, and other details.
* [TIFF Technical Notes](https://www.awaresystems.be/imaging/tiff/specification/TIFFPM6.pdf) specification including details about SubIFDs and TIFF Trees.
* [Multi Page / Multi Image TIFF](http://www.simplesystems.org/libtiff/multi_page.html) documented by the `libtiff` library used to create TIFF files.
