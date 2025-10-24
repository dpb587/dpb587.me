---
title: Add schema.org JSON-LD to Hugo Templates
description: Implementing and testing structured data for this web site.
publishDate: 2025-10-24
params:
  nav:
    tag:
      schema.org: true
      structured data: true
      hugo: true
---

I wanted to restore "structured data" to this site using some of the basic [schema.org types](https://schema.org/) that [Google Search supports](https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data) for their search results. The process was fairly straight forward...

1. **[Add `structured-data` Block](#structured-data-block)** within `baseof.html` for shared template definitions.
1. **[Add `structured-data.html` Partial](#structured-data-partial)** for generation of default types.
   1. **[Use `dict` and `jsonify`](#dict-and-jsonify)** for safely encoding JSON data.
   1. **[Add Global Properties](#global-properties)** based on Hugo's common `Page` metadata.
   1. **[Add Detailed Properties](#detailed-properties)** based on site- and type-specific metadata.
1. **[Update Content Templates](#templates)** to include the partial on relevant pages.
1. **[Custom Pages](#custom-pages)** with unique structured data requirements.
1. **[Validate Structured Data](#validate-structured-data)** to bulk test the generated site before publishing.

Continue reading for more commentary, review the [original commit](https://github.com/dpb587/dpb587.me/commit/c9d27d4bbe8e0749b15fe15d190087111ba36388), or view the [latest code](https://github.com/dpb587/dpb587.me).

## Add `structured-data` Block {#structured-data-block}

I chose JSON-LD as the easiest method to generate and embed data. Google Search requires this data to be embedded within the `head` or `body` tag, so I created a new block within the [`layouts/baseof.html` template](https://github.com/dpb587/dpb587.me/blob/main/hugo/themes/current/layouts/baseof.html).

```go-html-template
<html {{/* ... */}}>
  <head>
    {{/* ... */}}
    {{ block "structured-data" . }}{{ end }}
  </head>
  {{/* ... */}}
```

## Add `structured-data.html` Partial {#structured-data-partial}

Most pages can be generated using the same generic template, so I decided to use a partial named [`structured-data.html`](https://github.com/dpb587/dpb587.me/blob/main/hugo/themes/current/layouts/_partials/structured-data.html).

### Use `dict` and `jsonify` {#dict-and-jsonify}

Rather than use string interpolation which can lead to JSON encoding issues, I used Hugo's [`dict` function](https://gohugo.io/functions/collections/dictionary/) to build in-memory data structures. Within `structured-data.html`, I initialized a `$sd` variable with basic properties...

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="3-17" lang="go-html-template" >}}

Then, at the end of the file after any modifications of `$sd`, I use the [`jsonify` function](https://gohugo.io/functions/encoding/jsonify/) to encode it as JSON (and the [`safeJS` function](https://gohugo.io/functions/safe/js/) to avoid the output being double escaped).

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="253" lang="go-html-template" >}}

### Add Global Properties {#global-properties}

There are a few schema.org properties that can be applied to any types, so I can include this snippet. The `2006-01...` sequence ensures the dates are formatted in the ISO 8601 format recommended by schema.org.

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="19-29" lang="go-html-template" >}}

For hierarchical, section-based content, I use the following snippet to include the [`isPartOf` property](https://schema.org/isPartOf).

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="247-251" lang="go-html-template" >}}

I also have some site-specific, global parameters that I wanted to include. The following snippet generates [`keywords` properties](https://schema.org/keywords) based on my custom tagging conventions.

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="31-41" lang="go-html-template" >}}

The following snippet uses a nested template to generate structured data for my "place" taxonomies. I happened to use a partial, though it could probably be a type-specific template, too. I used the [`unmarshal` function](https://gohugo.io/functions/transform/unmarshal/) to convert the rendered template string back into a dictionary that can be added to `$sd`.

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="43-53" lang="go-html-template" >}}

### Add Detailed Properties {#detailed-properties}

Next, I added more conditional logic based on the page type. I use the `post` type as a generic blog post, so include the [`BlogPosting` type](https://schema.org/BlogPosting) and [`wordCount` property](https://schema.org/wordCount).

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="59-64" lang="go-html-template" >}}

I use a `media` content type for my images, so I include some logic for that. This implementation is very specific to my site's metadata conventions. The full source has some good examples of building nested objects such as wrapping the [`ImageObject` type](https://schema.org/ImageObject) as the [`mainEntity`](https://schema.org/mainEntity), [`PropertyValue` types](https://schema.org/PropertyValue) for the [`exifData` property](https://schema.org/exifData), and several other image-related properties.

```go-html-template
{{- else if eq .Type "media" -}}
  {{- $mainEntity := dict "@type" "ImageObject" -}}

  {{- with .Params.mediaType -}}
    {{- with .captureTime.time -}}
      {{- $mainEntity = merge $mainEntity ( dict "dateCreated" . ) -}}
    {{- end -}}

    {{- with .exifProfile -}}
      {{- $exifData := slice -}}

      {{- with .apertureValue.number -}}
        {{- $exifData = $exifData | append ( dict
          "@type" "PropertyValue"
          "name" "fNumber"
          "value" .
        ) -}}
      {{- end -}}

      {{/* ... */}}

      {{- if gt (len $exifData) 0 -}}
        {{- $mainEntity = merge $mainEntity ( dict "exifData" $exifData ) -}}
      {{- end -}}
    {{- end -}}
  {{ end }}

  {{- with .width -}}
    {{- $mainEntity = merge $mainEntity ( dict "width" . ) -}}
  {{- end -}}

  {{/* ... */}}

  {{- $sd = merge $sd ( dict
    "@type" "WebPage"
    "name" .Title
    "mainEntity" $mainEntity
  ) -}}
```

Finally, if no other type matches, I make sure it falls back to a [`WebPage` type](https://schema.org/WebPage). Google doesn't specifically use this type for their search enhancements, but it helps complete a valid, typed entity.

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="structured-data.html" lines="240-245" lang="go-html-template" >}}

## Update Content Templates {#templates}

With the `_partials/structured-data.html` file available, I can now include it in my existing templates; namely `page.html` and `section.html`.

```go-html-template
{{ define "structured-data" }}
  {{ partial "structured-data" . }}
{{ end }}
```

## Custom Pages {#custom-pages}

The generic, partial template is great for most of the pages, but some pages need one-off structured data. For example, in `home.html` I wanted to use the [`ProfilePage` type](https://schema.org/ProfilePage) with a [`Person` entity](https://schema.org/Person) to describe some basics about myself.

```go-html-template
{{ define "structured-data" }}
  <script type="application/ld+json">{{ ( dict
    "@context" "https://schema.org"
    "@type" "ProfilePage"
    "url" .Permalink
    "name" .Title
    "mainEntity" ( dict
      "@type" "Person"
      "name" "Danny Berger"
      "familyName" "Berger"
      "givenName" "Danny"
      "sameAs" ( slice
        .Permalink
        "https://www.linkedin.com/in/dpb587/"
        "https://twitter.com/dpb587"
        "https://github.com/dpb587"
        "https://gitlab.com/dpb587"
      )
      "homeLocation" ( dict
        "@type" "Place"
        "address" ( dict
          "@type" "PostalAddress"
          "addressLocality" "Albuquerque"
          "addressRegion" "NM"
          "addressCountry" "US"
        )
        "name" "Albuquerque, New Mexico"
      )
      "image" ( absURL "/assets/images/dpb587-20221205b~256.jpg" )
    ) ) | jsonify | safeJS }}</script>
{{ end }}
```

## Validate Structured Data {#validate-structured-data}

While prototyping, I validated my changes using Google's [Rich Results Test](https://search.google.com/test/rich-results). It only supports testing live web pages or manually pasted code, so I copy-pasted a few pages manually. The only errors/warnings it found were related to using `localhost` URLs.

{{< image alt="Rich Results Test - Home" src="./media/rich-results-test-home.png" >}}

To bulk test all my new structured data, I tried to "dog food" the [Structured Data API](https://www.namedgraph.com/toolkit/structured-data) and hacked together a script to iterate all the JSON-LD references after `hugo build`...

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="audit-structured-data.sh" lines="33-35" lang="bash" >}}

And then sent it to the endpoint and print any messages, if they show up, for manual review...

{{< snippet dir="appendix/2025-10-24-add-schema-org-json-ld-to-hugo-templates" file="audit-structured-data.sh" lines="18-28" stripprefix="  " lang="bash" >}}

It found a typo that I was able to fix (Google silently autocorrects these), but otherwise no surprises.

```
./entries/minikube-and-bridged-networking-20200411/index.html
+++ ERROR [Unknown Property] The "datePUblished" term is not a known schema.org/Property.
...
```

I committed [`audit-structured-data.sh` script](https://github.com/dpb587/dpb587.me/blob/main/scripts/audit-structured-data.sh) to manually use later in case I make further changes to the structured data templates.

## Monitor Results

Once published, Google Bot and other crawlers will take some time to discover the new structured data. The [Search Console](https://search.google.com/search-console/about) offers basic insights through their "Enhancements" and "Core Web Vitals" reports.

{{< image alt="Google Search Console - Overview" src="./media/search-console-enhancements.png" >}}

Of course, whether or not any structured data actually makes a difference to search results and rankings is ultimately up to the search engines.
