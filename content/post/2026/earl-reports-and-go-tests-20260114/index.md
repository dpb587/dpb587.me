---
title: EARL Reports and Go Tests
description: An earltesting package for the W3C vocabulary.
publishDate: "2026-01-14"
params:
  nav:
    tag:
      earl: true
      go: true
      rdfkit: true
---

The Evaluation and Report Language (*EARL*), from [w3.org](https://www.w3.org/WAI/standards-guidelines/earl/), is a "machine-readable format for expressing test results" in a framework-agnostic manner. While working on the [`rdfkit-go` module](https://github.com/dpb587/rdfkit-go), I wanted to build these reports for some of the common, public test suites that I was already testing against. So, I have been experimenting with a small [`earltesting` package](https://pkg.go.dev/github.com/dpb587/rdfkit-go/ontology/earl/earltesting) to generate these reports from Go test cases.

```go
import "github.com/dpb587/rdfkit-go/ontology/earl/earltesting"
```

This post shows how the [`w3c-github-json-ld-api-toRdf` package](https://github.com/dpb587/rdfkit-go/tree/main/encoding/jsonld/testsuites/w3c-github-json-ld-api-toRdf) uses `earltesting` to build an EARL Report for the [Deserialize JSON-LD to RDF algorithm](https://www.w3.org/TR/2020/REC-json-ld11-api-20200716/#deserialize-json-ld-to-rdf-algorithm) based on the public [Test Suite](https://w3c.github.io/json-ld-api/tests/) published by [JSON-LD](https://json-ld.org/).

{{< image alt="Screenshot: EARL Report, Visual Summary" src="./media/earl-report-screenshot.png" caption="w3c-github-json-ld-api-toRdf.ttl (github.com/dpb587/rdfkit-go) | EARL" >}}

## Terminology

As a general purpose solution, EARL Reports use a few extra terms compared a traditional Go test. They are formally defined in the [EARL 1.0 Schema](https://www.w3.org/TR/EARL10-Schema/), but here is a summary:

* **Assertor** - describes the *who* or *what* that is carrying out a test. In Go, I base this on the package performing the test, and typically use a separate package for public test suites.
* **TestSubject** - describes *what* is being tested against. In Go, I base this on the package being tested.
* **Test** - identifies the criterion being evaluated. In Go, I use subtests for these reports, and each descendant `t.Run` represents a test.
* **Assertion** - describes the context of the execution, evaluation, and result of an individual test.
* **Outcome** - categorizes the result of a test from a limited set of options. In Go, this maps from tests failing, passing, or, ambiguously, being marked as skipped.

## Start a Report

Within a test, you can start a report with the `NewReport` function. At a minimum, the assertor and test subject should be configured. Any properties may be used, but the [schema](https://www.w3.org/TR/EARL10-Schema/)-recommended properties will look similar to the following.

```go
func Test(t *testing.T) {
  earlReport := earltesting.NewReport(t).
    WithAssertor(
      rdf.IRI("#assertor"),
      rdfdescription.NewStatementsFromObjectsByPredicate(rdfutil.ObjectsByPredicate{
        rdfiri.Type_Property: rdf.ObjectValueList{
          earliri.Software_Class,
        },
        foafiri.Name_Property: rdf.ObjectValueList{
          xsdobject.String("rdfkit-go test suite"),
        },
        foafiri.Homepage_Property: rdf.ObjectValueList{
          rdf.IRI("https://github.com/dpb587/rdfkit-go/tree/main/encoding/jsonld/testsuites/w3c-github-json-ld-api-toRdf"),
        },
      })...,
    ).
    WithTestSubject(
      rdf.IRI("#subject"),
      rdfdescription.NewStatementsFromObjectsByPredicate(rdfutil.ObjectsByPredicate{
        rdfiri.Type_Property: rdf.ObjectValueList{
          earliri.Software_Class,
          rdf.IRI("http://usefulinc.com/ns/doap#Project"),
        },
        foafiri.Name_Property: rdf.ObjectValueList{
          xsdobject.String("rdfkit-go/encoding/jsonld"),
        },
        foafiri.Homepage_Property: rdf.ObjectValueList{
          rdf.IRI("https://pkg.go.dev/github.com/dpb587/rdfkit-go/encoding/jsonld"),
        },
        rdf.IRI("http://usefulinc.com/ns/doap#programming-language"): rdf.ObjectValueList{
          xsdobject.String("Go"),
        },
        rdf.IRI("http://usefulinc.com/ns/doap#repository"): rdf.ObjectValueList{
          rdf.IRI("https://github.com/dpb587/rdfkit-go"),
        },
      })...,
    )
    
    // ...
}
```

The `With*` methods update the current scope and any future EARL Assertions will reference the resources with the `earl:assertedBy` and `earl:subject` properties.

## Annotating Test Cases

Each test case published by shared conformance tests are assigned a unique identifier, typically an IRI. The identifier helps ensure the results from independent reports can be correlated. The identifier is required when creating an EARL Assertion, but I also happen to use it for the Go subtest name.

```go
for _, sequence := range testdataManifest.Sequences {
  t.Run(string(sequence.ID), func(t *testing.T) {
    tAssertion := earlReport.NewAssertion(t, sequence.ID)

    // ...
  })
}
```

Once created, a cleanup handler (via `t`) automatically captures the EARL Outcome of the test. That is, a failure becomes `earl:failed`, a success becomes `earl:passed`, and a skip becomes `earl:cantTell`. To force a specific outcome, use the `SetTestResultOutcome` method or, if appropriate, the `Skip` wrapper.

```go
if sequence.Option.ProduceGeneralizedRdf {
  tAssertion.Skip(earliri.Inapplicable_NotApplicable, "produceGeneralizedRdf is not supported")
}
```

The `Skip` method requires an EARL Outcome value and optional log values which are also sent to the standard `t.Skip(...)` method. EARL Test Results do not have an explicit property for "log" data, so I record it in a `dc:description` property. The `Error`, `Fatal`, and `Log` family of methods are also supported which similarly record the log data, and then get sent to the standard `t` methods.

Personally, I only use assertion logging for messages that are expected to be user-facing and worth documenting, not generic failures or technical data dumps. For example:

```go
if slices.Contains(sequence.Type, "jld:NegativeEvaluationTest") {
  _, err := decodeAction()
  if err != nil {
    tAssertion.Logf("error (expected): %v", err)
  } else {
    t.Fatal("expected error, but got none")
  }
}
```

Once a test case has finished, its report data will look similar to the following Turtle resource.

```turtle
[]
  a earl:Assertion ;
  earl:assertedBy <#assertor> ;
  earl:mode earl:automatic ;
  earl:result [
    a earl:TestResult ;
    dc:date "2026-01-14T02:01:33Z"^^xsd:dateTime ;
    dc:description "produceGeneralizedRdf is not supported\n" ;
    earl:outcome earl:inapplicable
  ] ;
  earl:subject <#subject> ;
  earl:test <https://w3c.github.io/json-ld-api/tests/toRdf#t0118> .
```

## Saving Reports

Reports do not have a default output. For now, it's easier to use the `NewReportFromEnv` function. This approach takes care of writing the report to a file once the test suite is complete, and it relies on the following environment variables.

* `TESTING_EARL_OUTPUT` - file path, relative to the test package, for the N-Triples output. Or, if the file extension is `.ttl`, a Turtle file will be created.
* `TESTING_EARL_SUBJECT_RELEASE_*` - if non-empty, added to subjects under the `doap:release` property. Intended for use when producing reports for tagged releases.
  * `TESTING_EARL_SUBJECT_RELEASE_NAME` for `doap:name`
  * `TESTING_EARL_SUBJECT_RELEASE_REVISION` for `doap:revision`
  * `TESTING_EARL_SUBJECT_RELEASE_DATE` for `dc:created`

I use the env-configured approach which behaves similar to the following.

```console
$ TESTING_EARL_OUTPUT=earl-report.ttl go test ./...
$ find . -name earl-report.ttl
./encoding/trig/testsuites/w3-2013-TrigTests/earl-report.ttl
./encoding/nquads/testsuites/w3-2013-N-QuadsTests/earl-report.ttl
./encoding/turtle/testsuites/w3-2013-TurtleTests/earl-report.ttl
./encoding/rdfxml/testsuites/w3-2013-RDFXMLTests/earl-report.ttl
./encoding/jsonld/internal/jsonldinternal/testsuites/w3c-github-json-ld-api-expand/earl-report.ttl
./encoding/jsonld/testsuites/w3c-github-json-ld-api-toRdf/earl-report.ttl
./encoding/ntriples/testsuites/w3-2013-N-TriplesTests/earl-report.ttl
./rdfcanon/testsuites/w3c-rdf-canon-tests/earl-report.ttl
```

Before committing them directly to version control, keep in mind that the reports include timestamps which change every time. I include the files in my `.gitignore` and treat them as build artifacts.

## Publishing Reports

Since `rdfkit-go` is now using GitHub Actions, I have a [`test.sh` step](https://github.com/dpb587/rdfkit-go/blob/885f78b0a70d19374b79470035c7684f54cf3c0e/dev/scripts/test.sh#L87-L98) which collects all the EARL Reports, and then an [`upload-artifact` step](https://github.com/dpb587/rdfkit-go/blob/885f78b0a70d19374b79470035c7684f54cf3c0e/.github/workflows/dev.yaml#L22-L26) to retain them for a short period of time.

{{< image alt="Screenshot: GitHub Actions Artifacts" src="./media/github-actions-artifacts.png" caption="GitHub Actions Artifacts (dpb587/rdfkit-go@885f78b)" >}}

As a self-certification method, there is not really a central location to share results with. In practice, standards development tracks seem to rely on manually-aggregated reports to monitor community implementation before a standard is formally recommended. Once recommended, though, the reports may not be referenced or updated again.

In general, EARL Reports are a fairly niche mechanism of reporting, and they seem like a technical solution that few technical developers have a reason to care about. In theory, I think these reports (and the underlying, shared test suites) have more potential for discovering and evaluating software implementations across the web. I hope to experiment with more ideas about this in the future.
