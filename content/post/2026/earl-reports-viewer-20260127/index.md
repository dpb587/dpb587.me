---
title: EARL Reports Viewer
description: Browse reports and tests from the browser.
publishDate: "2026-01-27"
params:
  nav:
    tag:
      earl: true
      go: true
      rdfkit: true
---

While generating [EARL Reports](https://www.w3.org/WAI/standards-guidelines/earl/) from [Go unit tests](../earl-reports-and-go-tests-20260114), I wanted a nicer method of browsing both my own and peer reports. In my search for tools, I only found [earl-report](https://github.com/gkellogg/earl-report) which is used to generate the static *Conformance Reports* often referenced from W3C documentation ([example](https://www.w3.org/2013/N-QuadsReports/index.html)). I was hoping for something with more cross-referenced metadata and ability to integrate decentralized sources.

With the help of [Copilot](https://github.com/features/copilot), through both [Visual Studio Code](https://code.visualstudio.com) and [Agents on GitHub](https://github.com/features/copilot/agents), I was able to hack together some of the ideas into a useful prototype. The [source code](https://github.com/dpb587/earl-index-app) is now on GitHub, but I feel obligated to mention the code isn't particularly elegant and wasn't the main purpose of this experiment. Instead, I was focused on researching EARL practices, learning more about AI-based workflows, expanding the use cases of [rdfkit-go](https://github.com/dpb587/rdfkit-go), and having another method of reviewing my own EARL reports.

* [**Source Code**](https://github.com/dpb587/earl-index-app) --- github.com/dpb587/earl-index-app
* [**Preview Site**](https://earl.dpb.io/) --- earl.dpb.io

You can navigate the [earl.dpb.io](https://earl.dpb.io/) site yourself (assuming my dev server is still up and running); or continue reading to see some screenshots and more background notes about the main pages.

## Sources

Since there is not a single list for all publishers of report or test manifest files, I created a [config file](https://github.com/dpb587/earl-index-app/blob/ebf7ab6931ea6dc5cdab2c176d89c8976dddf1fc/config/public.yaml) listing the ones that I had been referencing. The *Sources* page acts as an index for them.

{{< image alt="Screenshot: Sources" src="./media/source-list.png" caption="Sources | EARL Index" href="https://earl.dpb.io/source-list" >}}

W3C and others publish files through traditional `http`/`https` URLs or `git` repositories, so I support reading from both. Personally, I wanted my own EARL Reports to be produced as part of automated testing and release pipelines, so it also supports pulling from artifacts produced by [GitHub Actions](https://github.com/features/actions).

## Source Revision

For each source, it has a details page showing the Report files that were found. For example, [`rdfkit-go`](https://github.com/dpb587/rdfkit-go) has multiple report files, each one from testing a Go packages against a public test suite.

{{< image alt="Screenshot: github.com/dpb587/rdfkit-go" src="./media/source-rev.png" caption="github.com/dpb587/rdfkit-go | EARL Index" href="https://earl.dpb.io/source-rev?ref=git%3buri%3dhttps%253A%252F%252Fgithub.com%252Fdpb587%252Frdfkit-go.git%2fcommit%3brev%3dheads%252Fmain" >}}

Similarly, when a source has Test Manifest files, the page also includes a section for those.

{{< image alt="Screenshot: github.com/rdfa/rdfa.github.io" src="./media/source-rev-test-manifest-files.png" caption="github.com/rdfa/rdfa.github.io | EARL Index" href="https://earl.dpb.io/source-rev?ref=git%3buri%3dhttps%253A%252F%252Fgithub.com%252Frdfa%252Frdfa.github.io.git%2fcommit%3brev%3dheads%252Fmaster#test-manifest-files" >}}

Technically, each source supports multiple revisions for providers like `git`, but for now it only monitors the `HEAD` ref. Once I start tagging releases, I might try to maintain an archive of the tag/release reports to verify changes over time.

## Report File

Each report file link shows details about the source of the file (such as repository, commit, or artifact), and then shows details of the EARL Assertions, grouped by their `earl:subject`.

{{< image alt="Screenshot: w3c-github-json-ld-api-expand.ttl" src="./media/report-file.png" caption="w3c-github-json-ld-api-expand.ttl | EARL Index" href="https://earl.dpb.io/report-file?ref=git%3buri%3dhttps%253A%252F%252Fgithub.com%252Fdpb587%252Frdfkit-go.git%2fcommit%3brev%3dheads%252Fmain%2farchive.zip%3bpath%3dencoding%252Fjsonld%252Finternal%252Fjsonldinternal%252Ftestsuites%252Fw3c-github-json-ld-api-expand.ttl" >}}

Although uncommon, I include a `dc:description` with some of the test results where I want to document more details about the behavior, so I display that property, too.

{{< image alt="Screenshot: w3c-github-json-ld-api-expand.ttl (descriptions)" src="./media/report-file-descriptions.png" caption="w3c-github-json-ld-api-expand.ttl (descriptions) | EARL Index" href="https://earl.dpb.io/report-file?ref=git%3buri%3dhttps%253A%252F%252Fgithub.com%252Fdpb587%252Frdfkit-go.git%2fcommit%3brev%3dheads%252Fmain%2farchive.zip%3bpath%3dencoding%252Fjsonld%252Finternal%252Fjsonldinternal%252Ftestsuites%252Fw3c-github-json-ld-api-expand.ttl" >}}

These example screenshots are showing a file that was expanded from a ZIP archive. For regular GitHub files, each assertion also includes a link back to the line of the original `earl:outcome` statement.

## Test Details

Since each assertion correlates to a specific `earl:test`, a *Test Node* page provides more details about the test case. For example, the `#t0001` test from JSON-LD is based on a test manifest from `w3c.github.io` which I included in my sources list, so it shows a few notable properties and Turtle snippet.

{{< image alt="Screenshot: Test Node (https://w3c.github.io/json-ld-api/tests/expand-manifest#t0001)" src="./media/test-node.png" caption="Test Node (https://w3c.github.io/json-ld-api/tests/expand-manifest#t0001) | EARL Index" href="https://earl.dpb.io/test-node?node=https%3a%2f%2fw3c.github.io%2fjson-ld-api%2ftests%2fexpand-manifest%23t0001" >}}

The *Assertions* section lists all the related test results grouped by their files and subject. Test manifests are occasionally referenced in multiple ways (e.g. `http` vs `https`, `github.com` vs `github.io`) and I don't currently try to normalize the test IRIs, so this doesn't necessarily include *all* relevant assertions.

{{< image alt="Screenshot: Assertions from Test Node (https://w3c.github.io/json-ld-api/tests/expand-manifest#t0001)" src="./media/test-node-assertions.png" caption="Assertions from Test Node (https://w3c.github.io/json-ld-api/tests/expand-manifest#t0001) | EARL Index" href="https://earl.dpb.io/test-node?node=https%3a%2f%2fw3c.github.io%2fjson-ld-api%2ftests%2fexpand-manifest%23t0001" >}}

The parenthetical versions shown after the subject name are based on `doap:release`-related properties that are often included when testing specific versions of software.

## Test Manifest File

Similar to a report, each test manifest page shows details about the source file; and then it shows `mf:Manifest` resources along with the aggregated assertion outcomes for each of its entries.

{{< image alt="Screenshot: json-ld-api/tests/expand-manifest.jsonld" src="./media/test-manifest-file.png" caption="json-ld-api/tests/expand-manifest.jsonld | EARL Index" href="https://earl.dpb.io/test-manifest-file?ref=web%3buri%3dhttps%253A%252F%252Fw3c.github.io%252Fjson-ld-api%252Ftests%252Fexpand-manifest.jsonld" >}}
