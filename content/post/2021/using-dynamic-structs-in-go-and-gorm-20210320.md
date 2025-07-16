---
description: An experimental way to import arbitrary JSON into a database.
params:
    nav:
        tag:
            dynamic: true
            golang: true
            gorm: true
            jsonlines: true
            reflect: true
            struct: true
publishDate: "2021-03-20"
title: Using Dynamic Structs in Go and GORM
---

I was working with some JSON-LD data sources that I needed to import into a database for testing. Since I was already in a [Go project](https://golang.org/), I wanted to figure out how to manage the database schema dynamically. I have been using the [GORM package](https://gorm.io/index.html) to manage databases elsewhere, so it became a good excuse to test out the [`reflect` package](https://pkg.go.dev/reflect) with dynamic `struct`s of data.

# Prototypically, statically dynamic {#prototypically-statically-dynamic}

Traditionally, Go is not a very "dynamic" language, so I started off by using a sample record to figure out what it might look like. Starting from the JSON...

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="128-131" stripprefix="	" lang="go" >}}

I can then use the `reflect` package to define a few fields. The [`StructField` type](https://pkg.go.dev/reflect#StructField) can first be used to list out fields of the to-be `struct`. In this case, I know the `@table` must be a string, and any other field can be named `Field#` since it doesn't really matter.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="133-145" stripprefix="	" lang="go" >}}

Notice that the fields are defining both `json` tags (which will be used when reading the JSON input) as well as `gorm` tags (which will be used by the ORM when saving to the database).

Once the list of fields are ready, the [`StructOf` function](https://pkg.go.dev/reflect#StructOf) can be used for preparing a type. From there, I can create a new value of it where I can unmarshal the sample JSON.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="150-158" stripprefix="	" lang="go" >}}

```go
&struct {
  Table string "json:\"@type\" gorm:\"-\"";
  Field0 string "json:\"@id\" gorm:\"column:_id\"";
  Field1 time.Time "json:\"at\" gorm:\"column:at\"";
  Field2 float32 "json:\"value\" gorm:\"column:value\""
} {
  Table: "measurement",
  ID: "a42fadde-ee76-4687-8f8f-303e083461e8",
  Field0: time.Time{wall:0x0, ext:63739610241, loc:(*time.Location)(nil)},
  Field1: 79.3
}
```

Because it is printing a dynamic struct, it uses the more verbose, inline format. But, all the output looks correct!

For the most part, GORM will extract the fields automatically, but we will want to pull out the `Table` field. Since this is a dynamic `struct` we cannot plainly reference `sampleValue.Table`; but we can use `reflect` again to get the string with its [`FieldByName` function](https://pkg.go.dev/reflect#Value.FieldByName). In this case, the [`Elem` function](https://pkg.go.dev/reflect#Value.Elem) is used to make sure we're looking at our dynamic `struct` value (and not a field from the `reflect` internal objects).

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="160-161" stripprefix="	" lang="go" >}}

```
measurement
```

Next, I can use the ORM library to automatically `CREATE`/`ALTER` the table. Note that, in this situation, new fields can safely be added, but changing field types (e.g. float to boolean) is not supported.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="163-165" stripprefix="	" lang="go" >}}

And, finally, use the [`Create` function](https://pkg.go.dev/gorm.io/gorm#DB.Create) to save our record in the database. Although GORM supports defining the table on the model and avoid the repetition, it was easier to use the [`Table` function](https://pkg.go.dev/gorm.io/gorm#DB.Table) directly since this is a dynamic `struct`.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="167-169" stripprefix="	" lang="go" >}}

# Into the Dynamic {#into-the-dynamic}

Now that we have some functional building blocks it's time to make it work from arbitrary data.

## Building a Type {#building-a-type}

To start, we'll use a function that converts a generic JSON object and builds a `reflect.Type` from it.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="89" stripprefix="	" lang="go" >}}

Within it, we prepare `fields` to be a list of the `reflect.StructField`s. Since a `Table` field is required, it gets hard-coded before ranging through `map` of the record.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="98" stripprefix="	" lang="go" >}}

Inside the loop we can perform any special logic around converting keys or values for our data domain. For example, ignore the `@type` key, or converting values to native types (like the [`time.Time` type](https://pkg.go.dev/time#Time)).

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="110-115" stripprefix="		" lang="go" >}}

After we're done making changes, we add our generated `reflect.StructField` in a similar manner to the original prototype.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="117-121" stripprefix="		" lang="go" >}}

And, once all the key/values are added, we can finally return back the generated `struct`.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="124" stripprefix="	" lang="go" >}}

## Building a Value {#building-a-value}

Next, I add a new function which takes care of both building the type, creating a value, and then "remarshal" it - `marshal` back to JSON, then `unmarshal` into the `struct` value - before returning it back.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="74-87" lang="go" >}}

## Adding Some Data {#adding-some-data}

Finally, the main loop uses a [`json.Decoder`](https://pkg.go.dev/encoding/json#Decoder) to read in [JSON Lines](https://jsonlines.org/) before using the functions described earlier to get the record into the database.

{{< snippet dir="appendix/2021-03-20-using-dynamic-structs-in-go-and-gorm" file="main.go" lines="31-54" stripprefix="		" lang="go" >}}

Running the program and piping the same sample data via `STDIN`, it parses it, creates the table, and inserts the record.

{{< terminal>}}

  {{< terminal-input>}}

    ```bash
    go run . <<< '{ "@type": "measurement", "@id": "a42fadde-ee76-4687-8f8f-303e083461e8", "at": "2020-10-29T23:17:21Z", "value": 79.3 }'
    ```

  {{< /terminal-input>}}

  {{< terminal-output>}}

    ```go
    &struct { Table string "json:\"@type\" gorm:\"-\""; Field1 string "json:\"@id\" gorm:\"column:_id\""; Field2 time.Time "json:\"at\" gorm:\"column:at\""; Field3 float64 "json:\"value\" gorm:\"column:value\"" }{Table:"measurement", Field1:"a42fadde-ee76-4687-8f8f-303e083461e8", Field2:time.Time{wall:0x0, ext:63739610241, loc:(*time.Location)(nil)}, Field3:79.3}
    ```

  {{< /terminal-output>}}

{{< /terminal>}}

Looking directly at the database, we can verify the results as well.

{{< terminal>}}

  {{< terminal-input>}}

    ```bash
    sqlite3 main.sqlite \
      '.schema measurement' \
      'SELECT * FROM measurement'
    ```

  {{< /terminal-input>}}

  {{< terminal-output>}}

    ```
    CREATE TABLE `measurement` (`_id` text,`at` datetime,`value` real);
    a42fadde-ee76-4687-8f8f-303e083461e8|2020-10-29 23:17:21+00:00|79.3
    ```

  {{< /terminal-output>}}

{{< /terminal>}}

# Alternatives {#alternatives}

By the end, it was working well enough for testing and I learned more about the `reflect` package. Mission accomplished. Still, if you're working with a similar scenario, you might want to consider alternatives such as:

 * Use something other than Go - this type of dynamic implementation (for both Go and GORM) is not really an approach or use case they're designed for.
 * Use a library - the [`dynamic-struct`](https://github.com/Ompluscator/dynamic-struct) package, for examples, seems to provide a nicer abstraction for working with dynamic `struct`s if you really need them.
 * Use GORM's schema management directly - a [couple](https://github.com/go-gorm/gorm/tree/e1952924e2a844eca52e5030f7b46b78de6ec135/schema) [subpackages](https://github.com/go-gorm/gorm/tree/e1952924e2a844eca52e5030f7b46b78de6ec135/migrator) seem responsible for managing table schemas and could possibly be used directly (instead of defining the schema via `struct`).
