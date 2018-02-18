# tj - stdin line timestamps, JSON-friendly

`tj` timestamps lines read from standard input.

- [Get it](#get-it)
- [Use it](#use-it)
    - [JSON output](#json-output)
    - [Time format](#time-format)
    - [Template output](#template-output)
    - [Color output](#color-output)
    - [JSON input](#json-input)
    - [Stopwatch regex](#stopwatch-regex)
    - [Stopwatch regex template](#stopwatch-regex-template)
    - [Stopwatch condition](#stopwatch-condition)
- [Example](#example)
- [Comments](https://github.com/sgreben/tj/issues/1)


## Get it

Using go get:

```bash
go get -u github.com/sgreben/tj/cmd/tj
```

Or [download the binary](https://github.com/sgreben/tj/releases/latest) from the releases page.

Also available as a [docker image](https://quay.io/repository/sergey_grebenshchikov/tj?tab=tags):

```bash
docker pull quay.io/sergey_grebenshchikov/tj
```

Or using [docker-get](https://github.com/32b/docker-get):

```bash
docker-get https://github.com/sgreben/tj
```

## Use it

`tj` reads from stdin and writes to stdout.

```text
Usage of tj:
  -template string
    	either a go template (https://golang.org/pkg/text/template) or one of the predefined template names
  -time-format string
    	either a go time format string or one of the predefined format names (https://golang.org/pkg/time/#pkg-constants)
  -time-zone string
    	time zone to use (default "Local")
  -match-regex string
    	a regex pattern. if given, only tokens matching it (re)start the stopwatch
  -match-template string
    	go template, used to extract text used for -match-regex
  -match-condition string
    	go template. if given, only tokens that result in 'true' (re)start the stopwatch
  -match-buffer
    	buffer lines between matches of -match-regex / -match-condition, copy delta values from final line to buffered lines
  -match string
    	alias for -match-template
  -condition string
    	alias for -match-condition
  -regex string
    	alias for -match-regex
  -read-json
    	parse a sequence of JSON objects from stdin
  -scale string
    	either a sequence of hex colors or one of the predefined color scale names (colors go from fast to slow)
      (default "BlueToRed")
  -scale-fast duration
    	the lower bound for the color scale (default 100ms)
  -scale-slow duration
    	the upper bound for the color scale (default 2s)
  -scale-linear
    	use linear scale (default true)
  -scale-cube
    	use cubic scale
  -scale-cubert
    	use cubic root scale
  -scale-sqr
    	use quadratic scale
  -scale-sqrt
    	use quadratic root scale
  -version
    	print version and exit
```

### JSON output

The default output format is JSON, one object per line:

```bash
$ (echo Hello; echo World) | tj
```

```json
{"timeSecs":1517592179,"timeNanos":1517592179895262811,"time":"2018-02-02T18:22:59+01:00","deltaSecs":0.000016485,"deltaNanos":16485,"delta":"16.485µs","totalSecs":0.000016485,"totalNanos":16485,"total":"16.485µs","text":"Hello"}
{"timeSecs":1517592179,"timeNanos":1517592179895451948,"time":"2018-02-02T18:22:59+01:00","deltaSecs":0.000189137,"deltaNanos":189137,"delta":"189.137µs","totalSecs":0.000205622,"totalNanos":205622,"total":"205.622µs","text":"World"}
```

### Time format

You can set the format of the `time` field using the `-time-format` parameter:

```bash
$ (echo Hello; echo World) | tj -time-format Kitchen
```

```json
{"timeSecs":1517592194,"timeNanos":1517592194875016639,"time":"6:23PM","deltaSecs":0.000017142,"deltaNanos":17142,"delta":"17.142µs","totalSecs":0.000017142,"totalNanos":17142,"total":"17.142µs","text":"Hello"}
{"timeSecs":1517592194,"timeNanos":1517592194875197515,"time":"6:23PM","deltaSecs":0.000180876,"deltaNanos":180876,"delta":"180.876µs","totalSecs":0.000198018,"totalNanos":198018,"total":"198.018µs","text":"World"}
```

The [constant names from pkg/time](https://golang.org/pkg/time/#pkg-constants) as well as regular go time layouts are admissible values for `-time-format`:

| Name       | Format                              |
|------------|-------------------------------------|
| ANSIC      | `Mon Jan _2 15:04:05 2006`          |
| Kitchen    | `3:04PM`                            |
| RFC1123    | `Mon, 02 Jan 2006 15:04:05 MST`     |
| RFC1123Z   | `Mon, 02 Jan 2006 15:04:05 -0700`   |
| RFC3339    | `2006-01-02T15:04:05Z07:00`         |
| RFC3339Nano| `2006-01-02T15:04:05.999999999Z07:00`
| RFC822     | `02 Jan 06 15:04 MST`               |
| RFC822Z    | `02 Jan 06 15:04 -0700`             |
| RFC850     | `Monday, 02-Jan-06 15:04:05 MST`    |
| RubyDate   | `Mon Jan 02 15:04:05 -0700 2006`    |
| Stamp      | `Jan _2 15:04:05`                   |
| StampMicro | `Jan _2 15:04:05.000000`            |
| StampMilli | `Jan _2 15:04:05.000`               |
| StampNano  | `Jan _2 15:04:05.000000000`         |
| UnixDate   | `Mon Jan _2 15:04:05 MST 2006`      |

### Template output

You can also specify an output template using the `-template` parameter and [go template](https://golang.org/pkg/text/template) syntax:

```bash
$ (echo Hello; echo World) | tj -template '{{ .I }} {{.TimeSecs}} {{.Text}}'
```

```json
0 1516649679 Hello
1 1516649679 World
```

The fields available to the template are specified in the [`token` struct](cmd/tj/main.go#L18).

Some templates are pre-defined and can be used via `-template NAME`:

| Name       | Template                                         |
|------------|--------------------------------------------------|
| Color      | `{{color .}}█{{reset}} {{.Text}}`                |
| ColorText  | `{{color .}}{{.Text}}{{reset}}`                  |
| Delta      | `{{.DeltaNanos}} {{.Text}}`                      |
| Text       | `{{.Text}}`                                      |
| Time       | `{{.TimeString}} {{.Text}}`                      |
| TimeDelta  | `{{.TimeString}} +{{.DeltaNanos}} {{.Text}}`     |
| TimeColor  | `{{.TimeString}} {{color .}}█{{reset}} {{.Text}}`|

### Color output

To help identify durations at a glance, `tj` maps durations to a color scale. The pre-defined templates `Color` and `ColorText` demonstrate this:

```bash
$ (echo fast; 
   sleep 1; 
   echo slower; 
   sleep 1.5; 
   echo slow; 
   sleep 2; 
   echo slowest) | tj -template Color
```
![Color output](docs/images/colors.png)

The terminal foreground color can be set by using `{{color .}}` in the output template. The default terminal color can be restored using `{{reset}}`.

The color scale can be set using the parameters `-scale`, `-scale-fast`, and  `-scale-slow`:

- The `-scale` parameter defines the colors used in the scale.  
- The `-scale-fast` and `-scale-slow` parameters define the boundaries of the scale: durations shorter than the value of `-scale-fast` are mapped to the leftmost color, durations longer than the value of `-scale-slow` are mapped to the rightmost color.

The scale is linear by default, but can be transformed:

- `-scale-sqr`, `-scale-sqrt` yields a quadratic (root) scale
- `-scale-cube`, `-scale-cubert` yields a cubic (root) scale

There are several pre-defined color scales:

| Name                | Scale                  |
|---------------------|----------------------- |
| BlackToPurple       | `#000 -> #F700FF`      |
| BlackToRed          | `#000 -> #F00`         |
| BlueToRed           | `#00F -> #F00`         |
| CyanToRed           | `#0FF -> #F00`         |
| GreenToRed          | `#0F0 -> #F00`         |
| GreenToGreenToRed   | `#0F0 -> #0F0 -> #F00` |
| WhiteToPurple       | `#FFF -> #F700FF`      |
| WhiteToRed          | `#FFF -> #F00`         |
| WhiteToBlueToRed    | `#FFF -> #00F -> #F00` |

You can also provide your own color scale using the same syntax as the pre-defined ones.

### JSON input

Using `-read-json`, you can tell `tj` to parse stdin as a sequence of JSON objects. The parsed object can be referred to via `.Object`, like this:

```bash
$ echo '{"hello": "World"}' | tj -read-json -template "{{.TimeString}} {{.Object.hello}}"
```

```
2018-01-25T21:55:06+01:00 World
```

The exact JSON string that was parsed can be recovered using `.Text`:

```bash
$ echo '{"hello"   :    "World"} {   }' | tj -read-json -template "{{.TimeString}} {{.Text}}"
```

```
2018-01-25T21:55:06+01:00 {"hello"   :    "World"}
2018-01-25T21:55:06+01:00 {   }
```

### Stopwatch regex

Sometimes you need to measure the duration between certain *tokens* in the input.

To help with this, `tj` can match each line against a regular expression and only reset the stopwatch (`delta`, `deltaSecs`, `deltaNanos`) when a line matches. The regular expression can be specified via the `-match-regex` (alias `-regex`) parameter.

### Stopwatch regex template

When using `-match-regex`, you can also specify a template `-match-template` (alias `-match`) to extract text from the current token. The output of this template is matched against the stopwatch regex. 

This allows you to use only specific fields of JSON objects as stopwatch reset triggers. For example:

```bash
$ (echo {}; sleep 1; echo {}; sleep 1; echo '{"reset": "yes"}'; echo {}) | 
    tj -read-json -match .reset -regex yes -template "{{.I}} {{.DeltaNanos}}"
```

```
0 14374
1 1005916918
2 2017292187
3 79099
```

The output of the match template is stored in the field `.MatchText` of the `token` struct:

```bash
$ echo '{"message":"hello"}' | tj -read-json -match-template .message -template "{{.TimeString}} {{.MatchText}}"
```

```
2018-01-25T22:20:59+01:00 hello
```

### Stopwatch condition

Additionally to `-match-regex`, you can specify a `-match-condition` go template. If this template produces the literal string `true`, the stopwatch is reset - "matches" of the `-match-condition` are treated like matches of the `-match-regex`.

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
$ docker build . |
    tj -regex ^Step |
    jq -s 'max_by(.deltaNanos) | {step:.start.text, duration:.delta}'
```

```json
{"step":"Step 3/4 : RUN sleep 10","duration":"10.602026127s"}
```

Alternatively, using color output and buffering:

```bash
$ docker build . |
    tj -regex ^Step -match-buffer -template Color -scale-cube
```

![Docker build with color output](docs/images/docker.png)

## Comments

Feel free to [leave a comment](https://github.com/sgreben/tj/issues/1) or create an issue.