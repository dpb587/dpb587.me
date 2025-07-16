---
description: Parsing and validating parameters for upstream routing.
params:
    nav:
        tag:
            api: true
            golang: true
            iiif: true
            iiif-image-api: true
            opensource: true
            package: true
publishDate: "2024-02-15"
title: IIIF Image Requests with Go
---

I created and open-sourced a Go package dedicated to parsing and validating [IIIF Image API 3.0](https://iiif.io/api/image/3.0/) request parameters. Unlike related libraries, it does not try to support a specific image processing runtime. I designed it for a frontend service needing to process and route requests to upstream cache and image RPC services.

* [**Source Code**](https://github.com/dpb587/go-iiif-image-api-v3) --- github.com/dpb587/go-iiif-image-api-v3
* [**Documentation**](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3) --- pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3

# Parsing Image Requests {#parsing-image-requests}

It's a standard Go module that can be imported in any project. The "v3" part refers to the IIIF specification version - I'm not trying to version the package at this point. One or both of the following imports will typically be used.

```go
import (
  iiifimageapi "github.com/dpb587/go-iiif-image-api-v3"
  "github.com/dpb587/go-iiif-image-api-v3/imagerequest"
)
```

The [`ParseRawParams`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3/imagerequest#ParseRawParams) function can be used to parse the image parameters from a request URI. Depending on the HTTP routing framework, the parameters may already be split and decoded to build the [`RawParams`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3/imagerequest#RawParams) slice.

```go
var params = imagerequest.RawParams{"pct:25,25,50,50", "max", "90", "default.png"}

parsedParams, err := imagerequest.ParseRawParams(params)
```

Basic validation will be performed to ensure text and numbers are used according to the spec, and to make sure there are no unexpected modifier characters. If something goes wrong, an [`InvalidValueError`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3#InvalidValueError) type will be returned. Typically the error would be propagated back to a client as an HTTP 400 error. An example error looks like the following.

```go
err.(iiifimageapi.InvalidValueError).Error()
// "parsing region: percents[0]: invalid value (expecting float, range [0, 100])"
```

Without an error, the parameters have met the requirements of an IIIF Image request, and the `parsedParams` of low-level values will look like the following.

```go
imagerequest.ParsedParams{
  RegionIsEnum:    false,
  RegionEnum:      "",
  RegionIsPercent: true,
  RegionPercent:   [4]float32{25, 25, 50, 50},
  RegionPixels:    [4]uint32{0x0, 0x0, 0x0, 0x0},

  SizeIsConfined: false,
  SizeIsUpscaled: false,
  SizeIsEnum:     true,
  SizeEnum:       "max",
  SizeIsPercent:  false,
  SizePercent:    float32(0),
  SizePixels:     [2]*uint32{nil, nil},

  RotationIsMirrored: false,
  RotationAmount:     float32(90),

  Quality: "default",

  Format: "png",
}
```

If you only needed to filter requests for spec-valid parameters, you could now safely forward the request directly to an upstream image server. Use the original request URI, or you can use [`parsedParams.String()`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3/imagerequest#ParsedParams.String) to get a reconstructed IIIF Image Request path.

# Resolving with Image Metadata {#resolving-with-image-metadata}

Although the parameters are spec-valid, there are two other factors which can affect the validity, too:

1. Image metadata, including height and width, helps make sure requests don't exceed the readable dimensions.
2. Optional API features which may or may not be offered for images. For example, whether or not an image can be upscaled from its intrinsic size to a larger size.

The image metadata typically requires an external lookup to a cache or object store, so it doesn't always make sense to perform it outside an image processor. But if that metadata is available, you can call [`parsedParams.Resolve`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3/imagerequest#ParsedParams.Resolve) to further validate and transform a request.

```go
resolvedParams, err := parsedParams.Resolve(imagerequest.ResolveOptions{
  ImageInformation: iiifimageapi.ImageInformation{
    Profile: iiifimageapi.ComplianceLevel2Name,
    Width:   3024,
    Height:  4032,
  },
  DefaultQuality: "color",
})
```

If there's an error, it will be one of [`InvalidValueError`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3#InvalidValueError) (e.g. referring to pixels outside the image's dimensions) or [`FeatureNotSupportedError`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3#FeatureNotSupportedError) (e.g. trying to upscale when it's not allowed). Otherwise, in a successful call, the `resolvedParams` offers the following functions to access the resulting values.

```go
resolvedParams.RegionPixels()     // [4]uint32{756, 1008, 1512, 2016}
resolvedParams.SizePixels()       // [2]uint32{1512, 2016}
resolvedParams.RotationIsMirred() // false
resolvedParams.RotationAmount()   // float32(90)
resolvedParams.Quality()          // "color"
resolvedParams.Format()           // "png"
```

Here, the resolved parameters are exact pixel values based on the image metadata -- no more relative percentages and aliased terms. Depending on your use case, the resolve step might only be helpful for validating against the optional API features. In my case, I directly use the pixels in my RPC requests to image services.

## Canonical Link {#canonical-link}

The IIIF Image API also describes the notion of a canonical link for image requests. For example, if a request uses pixel coordinates of the full image, the canonical link would replace the pixels with the `full` alias. The [`resolvedParams.Canonical()`](https://pkg.go.dev/github.com/dpb587/go-iiif-image-api-v3/imagerequest#ResolvedParams.Canonical) function returns the canonical parameters.

```go
resolvedParams.Canonical().String()
// "756,1008,1512,2016/1512,2016/90/default.png"
```

In my case, I use the canonical parameters as part of the cache key at the frontend services.

# Testing {#testing}

I used a few strategies for testing the behavior in the package, and each ended up providing some additional insights. I started with translating the [specification examples](https://iiif.io/api/image/3.0/#4-image-requests) into standard Go [unit tests](https://github.com/dpb587/go-iiif-image-api-v3/blob/064f8e4333deefd391c76bbe07dda14eece99b00/imagerequest/resolver_test.go#L23).

* I found a typo in the specification's example for upscaling. I treated them as normative examples, so reconciled the test to meet the specification's other descriptions. I later found another mention for it in a [GitHub issue](https://github.com/dpb587/go-iiif-image-api-v3/blob/064f8e4333deefd391c76bbe07dda14eece99b00/imagerequest/resolver_test.go#L246).

Next, I used [libvips](https://www.libvips.org/), a popular image processing library, to test some of the tile-centric logic. I used [`vips dzsave` command](https://www.libvips.org/API/current/Making-image-pyramids.html) to generate IIIF image tiles for several sample images and made sure its results matched the package's enumeration logic.

* I found an apparent [off-by-one bug](https://github.com/dpb587/go-iiif-image-api-v3/blob/064f8e4333deefd391c76bbe07dda14eece99b00/pixelset/package_test.go#L55-L57) in the vips implementation where it uses an extra scale factor for image generation, but does not advertise it in the `info.json` manifest. I reconciled my logic by checking the behavior of another existing tool and experimenting. I still want to investigate the behavior further and see about sending a patch.

Finally, once I had a server running with an image processing backend, I was able to use the [Image Validator](https://iiif.io/api/image/validator/) offered by IIIF. This turned out to be more complicated than I expected, but overall helpful in building confidence of the package.

* I was discouraged when it started failing for quite a few cases. After a lot of digging, I learned the failures were mostly due to the validator apparently not having been tested after its upgrade to Python 3. I ended up submitting a [pull request](https://github.com/IIIF/image-validator/pull/97) with several fixes, but it seems like the repository is no longer being monitored.
* It did help me find a behavioral difference related to upscaling and maximum area constraints. I think the specification is a bit ambiguous about it, so I went with the validator's behavior which also will result in more stable server performance. I also found a Slack question from someone else asking about the ambiguity.
* I also found a couple differences in my server implementation (outside the scope of this package) for some of the HTTP headers. These ended up being mostly stylistic opinions, so decided to go with the validator's preference.

Aside from those strategies, I manually tested with popular IIIF Image viewers to make sure they were able to access IIIF Images. I haven't found any problems through them yet. At least not around this package for the IIIF Image API handling; low-level image processing bugs is a different story.

# Next Steps {#next-steps}

There are a few things I still have on my list to improve the package:

1. Better support `square` size requests. Currently it always uses the center of the image, but there are other cropping strategies this should be aware of (e.g. the "attention" strategy in libvips). I think it'll be handled with a new `CenterCoordinates` resolver option alongside `DefaultColor`.
1. Add some benchmark tests to understand its performance better. In a world of image processing and file systems, these validation calls have been negligible, but I'd like to understand if there's room for any improvements.

# Resources {#resources}

* [International Image Interoperability Framework](https://iiif.io) for more information about IIIF open standards and the consortium leading it.
* [Image API 3.0](https://iiif.io/api/image/3.0/#4-image-requests) to read the technical specification behind this package.
* [go-iiif](https://github.com/go-iiif/go-iiif) is an alternative Go-based, all-in-one image server. Its configuration and parameter parsing were implemented in a different way than I needed.
