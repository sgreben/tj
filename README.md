# ts - stdin line timestamps

`ts` timestamps lines read from standard input. 

## Get it

Using go get:

```bash
go get github.com/sgreben/ts/cmd/ts
```

Or [download the binary](https://github.com/sgreben/ts/releases) from the releases page.

## Use it

`ts` reads from stdin and writes to stdout.

```text
Usage of ts:
  -plain
    	-template='{{.Time}} +{{.DeltaNanos}} {{.Text}}'
  -previous
    	include previous line
  -start string
    	a regex pattern. if given, only lines matching it (re)start the stopwatch
  -template string
    	go template (https://golang.org/pkg/text/template)
  -timeformat string
```

### JSON output

The default output format is JSON, one object per line:

```bash
$ (echo Hello; echo World) | ts
```

```json
{"timeSecs":1516648762,"timeNanos":1516648762008900882,"time":"2018-01-22T20:19:22+01:00","deltaSecs":0.000015003,"deltaNanos":15003,"delta":"15.003µs","totalSecs":0.000015003,"totalNanos":15003,"total":"15.003µs","text":"Hello"}
{"timeSecs":1516648762,"timeNanos":1516648762009093926,"time":"2018-01-22T20:19:22+01:00","deltaSecs":0.000193044,"deltaNanos":193044,"delta":"193.044µs","totalSecs":0.000208047,"totalNanos":208047,"total":"208.047µs","text":"World"}
```

### Time format

You can set the format of the `time` field using the `-timeformat` parameter:

```bash
$ (echo Hello; echo World) | ts -timeformat Kitchen
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
$ (echo Hello; echo World) | ts -template '{{ .I }} {{.TimeSecs}} {{.Text}}'
```

```json
0 1516649679 Hello
1 1516649679 World
```

The fields available to the template are specified in the [`line` struct](cmd/ts/main.go#L14).

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
    ts -start ^Step |
    jq -s 'max_by(.deltaNanos) | {step:.start, duration:.delta}'
```

```json
{"step":"Step 3/4 : RUN sleep 10","duration":"10.602026127s"}
```