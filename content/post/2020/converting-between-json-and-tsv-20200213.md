---
description: Using jq to convert between data file formats.
params:
    nav:
        tag:
            jq: true
            jsonl: true
            tsv: true
publishDate: "2020-02-13"
title: Converting between JSON and TSV
---

Occasionally I want a quick way to convert data between JSON and TSV in BASH scripts. I use these [`jq` scripts](https:/stedolan.github.io/jq) to transform the formats from stdin.

```bash
tsv2jsonl() { jq -cRs 'split("\n") | map(split("\t")) | .[0] as $keys | .[1:-1] | map(. as $values | $keys | keys | map({"key":($keys[.]),"value":($values[.])}) | from_entries)[]' ; }
jsonl2tsv() { jq -sr '( .[0] | keys_unsorted ) as $keys | $keys, map([ .[ $keys[] ] ])[] | @tsv' ; }
```

# Example {#example}

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    cat holidays.tsv
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    date	holiday
    2020-01-01	New Years Day
    2020-01-20	Dr Martin Luther King Jr Day
    2020-02-17	Presidents Day
    2020-05-25	Memorial Day
    2020-07-03	Independence Day
    2020-09-07	Labor Day
    2020-11-26	Thanksgiving
    2020-11-27	Thanksgiving
    2020-12-24	Winter Holiday
    2020-12-25	Winter Holiday
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    # first to json
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    tsv2jsonl < holidays.tsv
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    {"date":"2020-01-01","holiday":"New Years Day"}
    {"date":"2020-01-20","holiday":"Dr Martin Luther King Jr Day"}
    {"date":"2020-02-17","holiday":"Presidents Day"}
    {"date":"2020-05-25","holiday":"Memorial Day"}
    {"date":"2020-07-03","holiday":"Independence Day"}
    {"date":"2020-09-07","holiday":"Labor Day"}
    {"date":"2020-11-26","holiday":"Thanksgiving"}
    {"date":"2020-11-27","holiday":"Thanksgiving"}
    {"date":"2020-12-24","holiday":"Winter Holiday"}
    {"date":"2020-12-25","holiday":"Winter Holiday"}
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    # chain commands
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    tsv2jsonl < holidays.tsv | jq -c 'select(.holiday | contains(" Day") | not)'
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    {"date":"2020-11-26","holiday":"Thanksgiving"}
    {"date":"2020-11-27","holiday":"Thanksgiving"}
    {"date":"2020-12-24","holiday":"Winter Holiday"}
    {"date":"2020-12-25","holiday":"Winter Holiday"}
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    # back to tsv
    ```

  {{< /terminal-input >}}

  {{< terminal-input >}}

    ```bash
    tsv2jsonl < holidays.tsv | jq -c 'select(.holiday | contains(" Day") | not)' | jsonl2tsv
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    date	holiday
    2020-11-26	Thanksgiving
    2020-11-27	Thanksgiving
    2020-12-24	Winter Holiday
    2020-12-25	Winter Holiday
    ```

  {{< /terminal-output >}}

{{< /terminal >}}
