---
title: Converting between JSON and TSV
description: Using jq to convert between data file formats.
date: 2020-02-13
tags:
- jq
- jsonl
- tsv
aliases:
- /blog/2020/02/13/converting-between-json-and-tsv.html
---

Occasionally I want a quick way to convert data between JSON and TSV in BASH scripts. I use these [`jq` scripts](https:/stedolan.github.io/jq) to transform the formats from stdin.

```bash
tsv2jsonl() { jq -cRs 'split("\n") | map(split("\t")) | .[0] as $keys | .[1:-1] | map(. as $values | $keys | keys | map({"key":($keys[.]),"value":($values[.])}) | from_entries)[]' ; }
jsonl2tsv() { jq -sr '( .[0] | keys_unsorted ) as $keys | $keys, map([ .[ $keys[] ] ])[] | @tsv' ; }
```

## Example

```console
$ cat holidays.tsv
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

# first to json
$ tsv2jsonl < holidays.tsv
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

# chain commands
$ tsv2jsonl < holidays.tsv | jq -c 'select(.holiday | contains(" Day") | not)'
{"date":"2020-11-26","holiday":"Thanksgiving"}
{"date":"2020-11-27","holiday":"Thanksgiving"}
{"date":"2020-12-24","holiday":"Winter Holiday"}
{"date":"2020-12-25","holiday":"Winter Holiday"}

# back to tsv
$ tsv2jsonl < holidays.tsv | jq -c 'select(.holiday | contains(" Day") | not)' | jsonl2tsv
date	holiday
2020-11-26	Thanksgiving
2020-11-27	Thanksgiving
2020-12-24	Winter Holiday
2020-12-25	Winter Holiday
```
