# tj - stdin line timestamps, JSON-friendly

`tj` timestamps lines read from standard input. 

<!-- TOC -->

- [Get it](#get-it)
- [Use it](#use-it)
    - [JSON output](#json-output)
    - [Time format](#time-format)
    - [Template output](#template-output)
    - [Stopwatch regex](#stopwatch-regex)
    - [JSON input](#json-input)
- [Example](#example)

<!-- /TOC -->

## Get it

Using go get:

```bash
go get -u github.com/sgreben/tj/cmd/tj
```

Or [download the binary](https://github.com/sgreben/tj/releases/latest) from the releases page.

## Use it

`tj` reads from stdin and writes to stdout.

```text
Usage of tj:
  -timeformat string
        either a go time format string or one of the predefined format names (https://golang.org/pkg/time/#pkg-constants)
  -template string
        go template (https://golang.org/pkg/text/template)
  -start string
        a regex pattern. if given, only lines matching it (re)start the stopwatch
  -readjson
        parse each stdin line as JSON
  -jsontemplate string
        go template, used to extract text from json input. implies -readjson
  -plain
        -template='{{.TimeString}} +{{.DeltaNanos}} {{.Text}}'
```

### JSON output

The default output format is JSON, one object per line:

```bash
$ (echo Hello; echo World) | tj
```

```json
{"timeSecs":1516648762,"timeNanos":1516648762008900882,"time":"2018-01-22T20:19:22+01:00","deltaSecs":0.000015003,"deltaNanos":15003,"delta":"15.003µs","totalSecs":0.000015003,"totalNanos":15003,"total":"15.003µs","text":"Hello"}
{"timeSecs":1516648762,"timeNanos":1516648762009093926,"time":"2018-01-22T20:19:22+01:00","deltaSecs":0.000193044,"deltaNanos":193044,"delta":"193.044µs","totalSecs":0.000208047,"totalNanos":208047,"total":"208.047µs","text":"World"}
```

### Time format

You can set the format of the `time` field using the `-timeformat` parameter:

```bash
$ (echo Hello; echo World) | tj -timeformat Kitchen
```

```json
{"timeSecs":1516648899,"timeNanos":1516648899954888290,"time":"8:21PM","deltaSecs":0.000012913,"deltaNanos":12913,"delta":"12.913µs","totalSecs":0.000012913,"totalNanos":12913,"total":"12.913µs","text":"Hello"}
{"timeSecs":1516648899,"timeNanos":1516648899955092012,"time":"8:21PM","deltaSecs":0.000203722,"deltaNanos":203722,"delta":"203.722µs","totalSecs":0.000216635,"totalNanos":216635,"total":"216.635µs","text":"World"}
```

The [constant names from pkg/time](https://golang.org/pkg/time/#pkg-constants) as well as regular go time layouts are admissible values for `-timeformat`:

| Name       | Format                              |
|------------|-------------------------------------|
| RubyDate   | Mon Jan 02 15:04:05 -0700 2006      |
| RFC3339    | 2006-01-02T15:04:05Z07:00           |
| Stamp      | Jan _2 15:04:05                     |
| StampMicro | Jan _2 15:04:05.000000              |
| RFC1123Z   | Mon, 02 Jan 2006 15:04:05 -0700     |
| Kitchen    | 3:04PM                              |
| RFC1123    | Mon, 02 Jan 2006 15:04:05 MST       |
| RFC3339Nano| 2006-01-02T15:04:05.999999999Z07:00 |
| RFC822     | 02 Jan 06 15:04 MST                 |
| RFC850     | Monday, 02-Jan-06 15:04:05 MST      |
| RFC822Z    | 02 Jan 06 15:04 -0700               |
| StampMilli | Jan _2 15:04:05.000                 |
| StampNano  | Jan _2 15:04:05.000000000           |
| ANSIC      | Mon Jan _2 15:04:05 2006            |
| UnixDate   | Mon Jan _2 15:04:05 MST 2006        |

### Template output

You can also specify an output template using the `-template` parameter and [go template](https://golang.org/pkg/text/template) syntax:

```bash
$ (echo Hello; echo World) | tj -template '{{ .I }} {{.TimeSecs}} {{.Text}}'
```

```json
0 1516649679 Hello
1 1516649679 World
```

The fields available to the template are specified in the [`line` struct](cmd/tj/main.go#L15).

### Stopwatch regex

Sometimes you need to measure the duration between certain *tokens* in the input.

To help with this, `tj` can match each line against a regular expression and only reset the stopwatch (`delta`, `deltaSecs`, `deltaNanos`) when a line matches.

The regular expression can be specified via the `-start` parameter.

### JSON input

Using `-readjson`, you can tell `tj` to parse each input line as a separate JSON object.  Fields of this object can be referred to via `.Object` in the `line` struct, like this:

```bash
$ echo '{"hello": "World"}' | tj -readjson -template "{{.TimeString}} {{.Object.hello}}"
```

```
2018-01-25T21:55:06+01:00 World
```

Additionally, you can also specify a template `-jsontemplate` to extract text from the object. The output of this template is matched against the stopwatch regex. 

This allows you to use only specific fields of the object as stopwatch reset triggers. For example:

```bash
$ (echo {}; sleep 1; echo {}; sleep 1; echo '{"reset": "yes"}'; echo {}) | 
    tj -jsontemplate "{{.reset}}" -start yes -template "{{.I}} {{.DeltaNanos}}"
```

```
0 14374
1 1005916918
2 2017292187
3 79099
```

The output of the JSON template is stored in the field `.JSONText` of the `line` struct:

```bash
$ echo '{"message":"hello"}' | tj -jsontemplate "{{.message}}" -template "{{.TimeString}} {{.JSONText}}"
```

```
2018-01-25T22:20:59+01:00 hello
```

## Example

Finding the slowest step in a `docker build` (using `jq`):

```bash
$ cat Dockerfile
FROM alpine
RUN echo About to be slow...
RUN sleep 10
RUN echo Done being slow
```

```bash
docker build . |
    tj -start ^Step |
    jq -s 'max_by(.deltaNanos) | {step:.startText, duration:.delta}'
```

```json
{"step":"Step 3/4 : RUN sleep 10","duration":"10.602026127s"}
```