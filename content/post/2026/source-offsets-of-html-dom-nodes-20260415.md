---
description: A smuggler's guide for capturing line+column offsets in Go.
title: Source Offsets of HTML DOM Nodes
publishDate: 2026-04-15
---

If you need to parse HTML in Go, you probably use the `golang.org/x/net/html` package. The `html.Parse` function tokenizes, parses, and builds a valid HTML5 DOM tree that can be used for content traversal (i.e. a `<body>` element that contains a `<p>` element that contains the `hello` text). But, when it comes to linters and code-mod tools, you probably want to know where those raw bytes originated in source code.

I found a [feature proposal](https://github.com/golang/go/issues/34302) for "offset tracking" like this, but it is stale and limited in its scope. I thought it would be risky to maintain a fork of the `html` library, and it seemed too complex to reimplement the HTML DOM specifications from scratch. Instead, and as a compromise...

I created [**`inspecthtml-go`**](https://github.com/dpb587/inspecthtml-go) which uses the standard `html` implementation, but it "smuggles" just enough metadata to support an extra `*html.Node` metadata service. This technical post describes the algorithm I used along with some of the HTML quirks which influenced its design.

## Node Metadata

First, a quick look at the kinds of offsets metadata which are relevant. In `inspecthtml`, this is represented by the [`NodeMetadata` type](https://pkg.go.dev/github.com/dpb587/inspecthtml-go@v0.0.0-20260306151804-6eee3382343a/inspecthtml#NodeMetadata), although the principles are fairly implementation-agnostic.

```html
<span class="styled"> hello<br/>world </span>  // span element
^-------------------^                          TokenOffsets             L1C1:L1C22;0x0:0x15
 ^--^                                          TagNameOffsets           L1C2:L1C6;0x1:0x5
      ^---^                                    TagAttr[0].KeyOffsets    L1C7:L1C12;0x6:0xb
            ^------^                           TagAttr[0].ValueOffsets  L1C13:L1C21;0xc:0x14
                                      ^-----^  EndTagTokenOffsets       L1C39:L1C46;0x26:0x2d
                     ^---------------^         GetInnerOffsets()        L1C22:L1C39;0x15:0x26
^-------------------------------------------^  GetOuterOffsets()        L1C1:L1C46;0x0:0x2d

<span class="styled"> hello<br/>world </span>  // " hello" text
                     ^----^                    TokenOffsets             L1C22:L1C28;0x15:0x1b

<span class="styled"> hello<br/>world </span>  // br element
                           ^---^               TokenOffsets             L1C28:L1C33;0x1b:0x20
                                               TagSelfClosing           true

<span class="styled"> hello<br/>world </span>  // "world " text
                                ^----^         TokenOffsets             TextToken=L1C33:L1C39;0x20:0x26
```

Internally, I use the [`cursorio` package](https://pkg.go.dev/github.com/dpb587/cursorio-go/cursorio) for tracking byte and line+column offsets (non-trivial in the world of Unicode where 1 byte != 1 column). Each of the `*Offsets` in the example are represented in Go with the [`TextOffsetRange` type](https://pkg.go.dev/github.com/dpb587/cursorio-go/cursorio#TextOffsetRange) which provides `From` and `Until` fields of their respective offsets.

## Token Smuggler

Now on to the tricky part. Since `html.Tokenizer` does not expose any byte offset metadata, and the `html.Parser` does not support correlation between a token and produced `*html.Node` values, I went with a kind of [man-in-the-middle](https://en.wikipedia.org/wiki/Man-in-the-middle_attack) (MITM) implementation using some state across three stages:

* **Byte Transform** (using the standard `html.Tokenizer` type) to capture all relevant byte offsets and inject metadata references into the byte stream.
* **HTML Parse** (using the standard `html.Parse` function) to build a valid HTML5 document based on the altered byte stream.
* **Tree Transform** to resolve and drop internally-smuggled metadata from the DOM.

To make this work (while ensuring the `*html.Node` result from `inspecthtml.Parse` and `html.Parse` remain equivalent), I had to use several different smuggling approaches based on each token type.

### Tags

For each **`StartTagToken`** (and **`SelfClosingTagToken`**), we want to capture the raw offsets of the token and the offsets for each attribute. The captured offsets can be stored in memory, and then, to maintain the reference, we secretly inject our own attribute.

```patch
- <P ID="lead" data-cms = "{&quot;id&quot;:&quot;a1b2c3d4&quot;}" hidden >
+ <P o="TAG-1" ID="lead" data-cms = "{&quot;id&quot;:&quot;a1b2c3d4&quot;}" hidden >
```

During the *Tree Transform* stage, for each `*html.Node` element, we can always remove the first attribute (`o`) and resolve it back to the original metadata.

{{< callout style="idea" >}}{{< markdown >}}

* The attribute is prepended to avoid getting clobbered by user syntax which may be malformed.
* If, after parse, an element has no attributes, it means the processor injected it and there is no metadata available. For example, a `tbody` element, if missing, will be injected to a `table`.

{{< /markdown >}}{{< /callout >}}

This takes care of the element name and container, but attributes are more complicated...

#### Attributes

There is not an equivalent `Raw` method to access the original bytes of attribute data. The closest is `TagAttr`, but it returns the decoded key and value (which masks syntax corrections, entity decoding, and standardization). Here's an example where values cannot be reversed into their original byte counts.

```go
// <P ID="lead" data-cms = "{&quot;id&quot;:&quot;a1b2c3d4&quot;}" hidden >

name, attr = r.tokenizer.TagName()
// name = "p"
// attr = true

for attr {
    attrKey, attrValue, attr = r.tokenizer.TagAttr()
    // [0] attrKey   = "id"
    //     attrValue = "lead"
    //     attr      = true
    // [1] attrKey   = "data-cms"
    //     attrValue = `{"id":"a1b2c3d4"}`
    //     attr      = true
    // [2] attrKey   = "hidden"
    //     attrValue = ""
    //     attr      = false
}
```

Until a [`RawTagAttr`-like function](https://github.com/golang/go/issues/34302#issuecomment-4012985579) is supported, I [implemented](https://github.com/dpb587/inspecthtml-go/blob/6eee3382343aaa8616cbb54f757024ff0a255b0e/inspecthtml/parser_reader.go#L98-L159) a combination of conditional regular expressions to capture the key and value offsets of attributes. Unfortunately, it is a little hacky and dependent on the Go implementation.

During the *Byte Transform* stage as part of tag processing, these attribute offsets are also stored in memory (bound by the injected `o="TAG-#"` attribute).

#### Closing Tags

The **`EndTagToken`** requires a different approach since we cannot inject attributes and we do not yet know which opening tag token it should be bound to (if any!). Instead, we capture and store the raw offsets, and then append a standard HTML comment with the metadata reference.

```patch
- </P>
+ </P><!--END-TAG-2-->
```

During the *Tree Transform* stage, for each `*html.Node` element, we can look ahead to the next sibling to see if it is an `END-TAG-*` comment. If so, drop the comment node from the tree, resolve it back to the original metadata, and update our node's metadata to include the closing tag offsets.

{{< callout style="idea" >}}{{< markdown >}}

* If an unaffiliated `END-TAG-*` comment is found, it was missing a start tag, the processor determined the start tag needed to be closed sooner, or it was re-parented entirely. For deterministic behavior, I opted to ignore and drop the comment.

{{< /markdown >}}{{< /callout >}}

#### Quirky Examples

While this approach for tags' metadata might seem overcomplicated, remember that HTML is very forgiving on syntax errors. Consider how these snippets and typos end up being interpreted.

| Input | Tokenizer + Processor |
| --- | --- |
| `<p</p>text` | `<p< p="">text</p<>` |
| `<p id=/>text` | `<p id="/">text</p>` |
| `<p id="lead"</p>text</p>` | `<p id="lead" <="" p="">text</p>` |
| `<p id = lead>text</p>` | `<p id="lead">text</p>` |
| `<p id=lead>text</p>` | `<p id="lead">text</p>` |
| `<p id="lead"suffix">text</p>` | `<p id="lead" suffix"="">text</p>` |
| `<p id="lead"class="css">text</p>` | `<p id="lead" class="css">text</p>` |
| `<p id=""">text` | `<p id="" "="">text</p>` |
| `<p id="&">text` | `<p id="&amp;">text</p>` |
| `<p id="&amp">text` | `<p id="&amp;">text</p>` |
| `<p id=&#47>text` | `<p id="/">text</p>` |

### Texts

The **`TextToken`** includes both whitespace-only and human-readable text, each of which has different implications during the *HTML Parse* stage.

For whitespace-only, we keep the original data and append a comment-based metadata reference for the original offsets. By keeping the whitespace, the processor can continue to drop irrelevant whitespace when possible (i.e. outside the `body` element).

```patch
- \t\s\s
+ \t\s\s<!--TEXT-3-->
```

For everything else, we record the original offsets and replace the entire text with a metadata reference.

```patch
- hello world
+ TEXT-4
```

During the *Tree Transform* stage, for each `*html.Node` which is a text node, split the data on whitespace and `TEXT-` pattern (since the standard processor will join sequences of text nodes), and:

* With a `TEXT-` prefix, resolve the metadata reference and swap it with a new text node+metadata pair using the original text.
* Without a `TEXT-` prefix (i.e. whitespace), look ahead to the next sibling to see if it is a `WS-*` comment. If so, drop the comment node from the tree, resolve its metadata reference, and update the text node with the metadata.

{{< callout style="idea" >}}{{< markdown >}}

* The Go tokenizer drops `U+0000` (`NUL`) bytes, so we have to emulate that behavior, too. Interestingly, other control characters do not seem to be affected.
* Unlike the standard processor, `inspecthtml` does *not* join sequences of text nodes. By keeping separate `*html.Node` values, each can maintain its own node metadata and original source offsets.
* It seems simpler to *always* append a comment-based reference after text, but it does not work for raw text nodes (e.g. `style`), nor when text nodes are re-parented for a compliant DOM (e.g. text directly within a `table` element being moved before it).

{{< /markdown >}}{{< /callout >}}

### Comments

Since we use **`CommentToken`** for our own markers, we must always transform them. Similar to other tokens, capture the raw offsets of the token, store the original in memory, and swap it for a reference.

```patch
- <!-- hello world -->
+ <!--COMMENT-5-->
```

During the *Tree Transform* stage, for each `*html.Node` which is a comment:

* With a `COMMENT-` prefix, resolve the metadata reference and swap with the original comment data.
* Otherwise, drop the comment as an indirect remnant of invalid user syntax or due to the processor re-parenting a node in an unexpected way.

{{< callout style="idea" >}}{{< markdown >}}

* The Go tokenizer does some magic to normalize non-standard and HTML5-style comments (e.g. `<?php...?>`, `<!tag...>`) that we have to emulate in our offset calculations.

{{< /markdown >}}{{< /callout >}}

## Testing

Since HTML is a very forgiving syntax, I was a little worried about potential discrepancies of the final DOM produced by `inspecthtml` and standard `html`. I created [`dev-compare`](https://github.com/dpb587/inspecthtml-go/blob/6eee3382343aaa8616cbb54f757024ff0a255b0e/examples/dev-compare/main.go) which tests that both DOMs render to the same source code. After using it with some internal and public example HTML repositories (e.g. [content-extractor-benchmark](https://github.com/markusmobius/content-extractor-benchmark)), I was able to fix numerous edge cases and improve code documentation with a few reminders, like:

* The DOMs are expected to be different due to disjoint, sequential text nodes (which is why rendered code is compared instead).
* The standard `html` library recursively parses *all* tags when processing foreign elements (e.g. SVG `style` tags contain code comment nodes, whereas HTML `style` tags are considered raw text). For now, I treat it as an expected difference rather than introducing and matching state management of tags.

I also use traditional [unit tests](https://github.com/dpb587/inspecthtml-go/blob/6eee3382343aaa8616cbb54f757024ff0a255b0e/inspecthtml/parser_test.go) to make assertions about the expected source offsets of nodes. And, more broadly, I have private, system-level tests which help me detect if `inspecthtml` changes end up affecting the `*html.Node` metadata in unexpected ways.

### Performance

As you can guess, all these steps come with notable overhead both in terms of compute and memory. Aside from needing to store several offset values for each piece, we have our own pre-parse tokenization and post-parse tree traversal steps. I haven't introduced formal benchmark or performance testing at this point, so I imagine there are some opportunities for improvement.

In practice, I usually rely on an `captureSourceOffsets` flag which controls whether the `html.Parse` call should be used vs the slower `inspecthtml.Parse` call. Since both result in an equivalent `*html.Node` document, callers can rely on a single DOM manipulation or traversal implementation.

## Future

The `inspecthtml` module already meets my needs to improve the DX/UX of validation tools with source code references, but, eventually, I might need to revisit:

* Performance testing, particularly to reduce memory usage and allocations in the general case.
* Configuration options which limits metadata capture to specific node types or element attributes.
* Optional retention of raw attribute values to support the calculation of intra-value offsets.
* Using a custom parser to efficiently support these goals, hopefully with a similar interface.

In the real world, it's used by my [alternative Schema.org Validator](https://www.namedgraph.com/intro/schema-org-validator) for source highlighting of Microdata and HTML+RDFa. If you are building similar linter or validation tools in Go, maybe you'll find this useful, too. Feel free to [open an issue](https://github.com/dpb587/inspecthtml-go/issues) if you discover a bug or have a suggestion.
