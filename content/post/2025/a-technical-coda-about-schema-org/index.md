---
title: A Technical Coda about schema.org
description: Tools for investigating changes between releases.
publishDate: "2025-12-01"
---

I created a [small web application](https://schemaorg-coda.dpb.io/) to help me in my tasks and technical investigations of the [schema.org](https://schema.org/) specifications. The official site is targeted towards publishers looking to embed their data; but, for investigating and monitoring upstream changes, it has always been a bit more tedious. The rest of the post includes a few details about the new pages I rely on as well as some [background notes](#background-notes).

* **[Release Changes](#release-changes)** to compare releases or development commits to identify changes.
* **[Term History](#term-history)** to review changes over time for a specific term.
* **[Term Details](#term-details)** to view properties and examples of a specific term.
* **[Example Details](#example-details)** to view code snippets of a specific example.

In terms of technology... for the frontend, I kept a minimalist design with [Tailwind CSS](https://tailwindcss.com), [Heroicons](https://heroicons.com), and [HTML templates](https://pkg.go.dev/html/template). The backend is all [Go](https://go.dev/) where I depend on several of my own modules. Most notably:

* **[dpb587/rdfkit-go](https://github.com/dpb587/rdfkit-go)** as the core library for parsing all the Turtle definition files, the JSON-LD/Microdata/RDFa code snippets, and working with the small RDF graphs.
* **[dpb587/schemaorg-examples-go](https://github.com/dpb587/schemaorg-examples-go)** as a small library to parse the examples file format used by the schemaorg repository (similar to the official, Python [implementation](https://github.com/schemaorg/schemaorg/tree/main/software/SchemaExamples)).

{{< image alt="Screenshot: Home Page" src="./media/home.png" caption="schemaorg-coda" href="https://schemaorg-coda.dpb.io/" >}}

## Release Changes {#release-changes}

While there is an official [Releases page](https://schema.org/docs/releases.html), the descriptions are often high-level summaries of change. When working on data processors, I am more interested in the data type changes along with the external vocabulary references and end-user recommendation changes.

The [**`ref-diff`** page](https://schemaorg-coda.dpb.io/ref-diff?name=branch%2Fmain) solves this for me by comparing two reference points to show the RDF-based changes in specifications. References can be tagged releases, such as v29.3, and development code, such as `main`. It then shows all the affected terms to clearly see updated data types, descriptions, and other metadata.

{{< image alt="Screenshot: Schema Release Changes" src="./media/ref-diff-terms.png" caption="Term Changes :: Changes (v29.3 to main@69a70e1) :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/ref-diff?name=branch%2Fmain&from=release%2Fv29.3" >}}

The comparison includes embedded links back to specific lines of code for the original definitions which allows me to quickly "blame" changes back to issues or pull requests for more context.

In addition to the detailed statement-level changes, I also include a "by the numbers" section to quickly check for any trends or to discover new meta-conventions I should be aware of. Or... typos?

{{< image alt="Screenshot: Schema Release Changes (Numbers)" src="./media/ref-diff-numbers.png" caption="Numbers :: Changes (v29.3 to main@69a70e1) :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/ref-diff?name=branch%2Fmain&from=release%2Fv29.3#numbers" >}}

I don't typically refer to it, but another section lists the new, dropped, and updated examples.

{{< image alt="Screenshot: Schema Release Changes (Examples)" src="./media/ref-diff-examples.png" caption="Example Changes :: Changes (v29.3 to main@69a70e1) :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/ref-diff?name=branch%2Fmain&from=release%2Fv29.3#examples" >}}

## Term History {#term-history}

While the `ref-diff` page is my primary tool, a [**`term-history`** page](https://schemaorg-coda.dpb.io/term-history?name=inLanguage&from=branch%2Fmain) offers a similar, term-specific changelog. I occasionally use this when I find odd data and I want to check if it was following past recommendations.

{{< image alt="Screenshot: Schema Term History" src="./media/term-history.png" caption="inLanguage History :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/term-history?name=inLanguage&from=branch%2Fmain" >}}

## Term Details {#term-details}

Since schema.org only documents the latest tagged version of a term, the [**`term`** page](https://schemaorg-coda.dpb.io/term?name=inLanguage&ref=branch%2fmain) lets me review the full definition, both direct and inverse properties along with other metadata, from any point in time.

{{< image alt="Screenshot: Schema Term" src="./media/term.png" caption="inLanguage :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/term?name=inLanguage&ref=branch%2fmain" >}}

The page also includes a comprehensive Examples section. The official site shows a few, manually-tagged term examples, but I parse the examples into their RDF graphs in order to list *all* examples using the term. To make it easier to find relevant examples, I include the graph traversal path leading to the term, too.

{{< image alt="Screenshot: Schema Term Examples" src="./media/term-examples.png" caption="Examples :: inLanguage :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/term?name=inLanguage&ref=branch%2fmain#examples" >}}

## Example Details {#example-details}

The examples link to a local [**`example`** page](https://schemaorg-coda.dpb.io/example?name=eg-0452) where I can see all the syntaxes in a single view. In addition to the syntax highlighting, I also bold + link all the Schema terms using the decoded RDF metadata.

{{< image alt="Screenshot: Schema Linked Example" src="./media/example-code.png" caption="eg-0452 :: schemaorg-coda" href="https://schemaorg-coda.dpb.io/example?name=eg-0452" >}}

## Background Notes {#background-notes}

Finally, having covered the basic features and screenshots, a few more random notes about this project...

**Alternatives** -- I was unable to find any existing Schema-specific community tools for this sort of thing, but it's possible I didn't use the right search terms. Since the Schema definitions are based on RDF, I tried a few generic tools over the years, but they never fully worked for my interests. The generic tools are more focused on diff'ing any ontology which may have *very* generic structures, complex definitions, and reporting requirements. Aside from a couple unmaintained tools that I didn't spend much time on, I tried:

* [WiDOCO](https://github.com/dgarijo/Widoco) -- a reasonable changelog report, but was focused on expected, limited predicates.
* [skos-history](https://github.com/jneubert/skos-history) -- uses some very comprehensive, SPARQL-based queries for analysis, but is more focused on SKOS concepts and expects you to be build your own, human-friendly report view.

**Runtime** -- this project currently runs from a container on a spare server, and I am not too concerned about its uptime. Although it caches most datasets in memory, it does take a couple seconds to load both the specification definitions and all the example graphs for a new reference. As a personal dev tool, "availability" just means caching the rendered results and automatic restarts after an OOM kill.

**History** -- over the years I had accumulated a collection of hacked-together scripts and code generation tools which helped implicitly track some of the upstream changes. With some time for a little side project over Thanksgiving break and some extra work from AI code agents, I was able to evolve those scripts into this nicer set of reports. I expect it to help me be a bit more productive and proactive.

**Source Code** -- at this point, only the underlying modules, such as [rdfkit-go](https://github.com/dpb587/rdfkit-go) and [schemaorg-examples-go](https://github.com/dpb587/schemaorg-examples-go), are open source. I haven't been motivated enough to review the iterated history of commits and remove some of my internal scripts that started this project. Maybe another day.
