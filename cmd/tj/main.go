package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/sgreben/tj/pkg/color"
)

type token struct {
	I           uint64        `json:"-"` // token index
	TimeSecs    int64         `json:"timeSecs"`
	TimeNanos   int64         `json:"timeNanos"`
	TimeString  string        `json:"time,omitempty"`
	Time        time.Time     `json:"-"`
	DeltaSecs   float64       `json:"deltaSecs"`
	DeltaNanos  int64         `json:"deltaNanos"`
	DeltaString string        `json:"delta,omitempty"`
	Delta       time.Duration `json:"-"`
	TotalSecs   float64       `json:"totalSecs"`
	TotalNanos  int64         `json:"totalNanos"`
	TotalString string        `json:"total,omitempty"`
	Total       time.Duration `json:"-"`
	MatchText   string        `json:"-"`
	Start       interface{}   `json:"start,omitempty"`
}

func (t *token) copyDeltasFrom(token *token) {
	t.DeltaSecs = token.DeltaSecs
	t.DeltaNanos = token.DeltaNanos
	t.DeltaString = token.DeltaString
	t.Delta = token.Delta
}

type configuration struct {
	timeFormat          string        // -time-format="..."
	timeZone            string        // -time-zone="..."
	template            string        // -template="..."
	matchRegex          string        // -match-regex="..."
	matchTemplate       string        // -match-template="..."
	matchCondition      string        // -match-condition="..."
	buffer              bool          // -match-buffer
	readJSON            bool          // -read-json
	scaleText           string        // -scale="..."
	scaleFast           time.Duration // -scale-fast="..."
	scaleSlow           time.Duration // -scale-slow="..."
	scaleCube           bool          // -scale-cube
	scaleSqr            bool          // -scale-sqr
	scaleLinear         bool          // -scale-linear
	scaleSqrt           bool          // -scale-sqrt
	scaleCubert         bool          // -scale-cubert
	printVersionAndExit bool          // -version
}

type printerFunc func(interface{}) error

var (
	version        string
	config         configuration
	printer        printerFunc
	matchRegex     *regexp.Regexp
	matchCondition *templateWithBuffer
	matchTemplate  *templateWithBuffer
	scale          color.Scale
	tokens         tokenStream
	location       *time.Location
)

func print(data interface{}) {
	if err := printer(data); err != nil {
		fmt.Fprintln(os.Stderr, "output error:", err)
	}
}

const ISO8601 = "2006-01-02T15:04:05Z07:00"

var timeFormats = map[string]string{
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"ISO8601":     ISO8601,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
}

var templates = map[string]string{
	"Text":           "{{.Text}}",
	"Time":           "{{.TimeString}} {{.Text}}",
	"TimeDeltaNanos": "{{.TimeString}} +{{.DeltaNanos}} {{.Text}}",
	"TimeDelta":      "{{.TimeString}} +{{.Delta}} {{.Text}}",
	"DeltaNanos":     "{{.DeltaNanos}} {{.Text}}",
	"Delta":          "{{.Delta}} {{.Text}}",
	"ColorText":      "{{color .}}{{.Text}}{{reset}}",
	"Color":          "{{color .}}█{{reset}} {{.Text}}",
	"DeltaColor":     "{{.Delta}} {{color .}}█{{reset}} {{.Text}}",
	"TimeColor":      "{{.TimeString}} {{color .}}█{{reset}} {{.Text}}",
}

var colorScales = map[string]string{
	"GreenToRed":        "#0F0 -> #F00",
	"GreenToGreenToRed": "#0F0 -> #0F0 -> #F00",
	"BlueToRed":         "#00F -> #F00",
	"CyanToRed":         "#0FF -> #F00",
	"WhiteToRed":        "#FFF -> #F00",
	"WhiteToPurple":     "#FFF -> #F700FF",
	"BlackToRed":        "#000 -> #F00",
	"BlackToPurple":     "#000 -> #F700FF",
	"WhiteToBlueToRed":  "#FFF -> #00F -> #F00",
}

var templateFuncs = template.FuncMap{
	"color": foregroundColor,
	"reset": func() string { return color.Reset },
}

func foregroundColor(o tokenOwner) string {
	token := o.Token()
	c := float64(token.DeltaNanos-int64(config.scaleFast)) / float64(config.scaleSlow-config.scaleFast)
	return color.Foreground(scale(c))
}

func jsonPrinter() printerFunc {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode
}

func templatePrinter(t string) printerFunc {
	template := template.Must(template.New("-template").Funcs(templateFuncs).Option("missingkey=zero").Parse(t))
	newline := []byte("\n")
	return func(data interface{}) error {
		err := template.Execute(os.Stdout, data)
		os.Stdout.Write(newline)
		return err
	}
}

func timeFormatsHelp() string {
	help := []string{}
	for name, format := range timeFormats {
		help = append(help, fmt.Sprint("\t", name, " - ", format))
	}
	sort.Strings(help)
	return "either a go time format string or one of the predefined format names (https://golang.org/pkg/time/#pkg-constants)\n" + strings.Join(help, "\n")
}

func templatesHelp() string {
	help := []string{}
	for name, template := range templates {
		help = append(help, fmt.Sprint("\t", name, " - ", template))
	}
	sort.Strings(help)
	return "either a go template (https://golang.org/pkg/text/template) or one of the predefined template names\n" + strings.Join(help, "\n")
}

func colorScalesHelp() string {
	help := []string{}
	for name, scale := range colorScales {
		help = append(help, fmt.Sprint("\t", name, " - ", scale))
	}
	sort.Strings(help)
	return "either a sequence of hex colors or one of the predefined color scale names (colors go from fast to slow)\n" + strings.Join(help, "\n")
}

func addTemplateDelimitersIfLiteral(t string) string {
	if !strings.Contains(t, "{{") {
		return "{{" + t + "}}"
	}
	return t
}

func init() {
	flag.StringVar(&config.template, "template", "", templatesHelp())
	flag.StringVar(&config.timeFormat, "time-format", "RFC3339", timeFormatsHelp())
	flag.StringVar(&config.timeZone, "time-zone", "UTC", `time zone to use (or "Local")`)
	flag.StringVar(&config.matchRegex, "regex", "", "alias for -match-regex")
	flag.StringVar(&config.matchRegex, "match-regex", "", "a regex pattern. if given, only tokens matching it (re)start the stopwatch")
	flag.StringVar(&config.matchCondition, "condition", "", "alias for -match-condition")
	flag.StringVar(&config.matchCondition, "match-condition", "", "go template. if given, only tokens that result in 'true' (re)start the stopwatch")
	flag.StringVar(&config.matchTemplate, "match", "", "alias for -match-template")
	flag.StringVar(&config.matchTemplate, "match-template", "", "go template, used to extract text used for -match-regex")
	flag.BoolVar(&config.buffer, "match-buffer", false, "buffer lines between matches of -match-regex / -match-condition, copy delta values from final line to buffered lines")
	flag.BoolVar(&config.readJSON, "read-json", false, "parse a sequence of JSON objects from stdin")
	flag.StringVar(&config.scaleText, "scale", "BlueToRed", colorScalesHelp())
	flag.DurationVar(&config.scaleFast, "scale-fast", 100*time.Millisecond, "the lower bound for the color scale")
	flag.DurationVar(&config.scaleSlow, "scale-slow", 2*time.Second, "the upper bound for the color scale")
	flag.BoolVar(&config.scaleCube, "scale-cube", false, "use cubic scale")
	flag.BoolVar(&config.scaleSqr, "scale-sqr", false, "use quadratic scale")
	flag.BoolVar(&config.scaleLinear, "scale-linear", true, "use linear scale")
	flag.BoolVar(&config.scaleSqrt, "scale-sqrt", false, "use quadratic root scale")
	flag.BoolVar(&config.scaleCubert, "scale-cubert", false, "use cubic root scale")
	flag.BoolVar(&config.printVersionAndExit, "version", false, "print version and exit")
	flag.Parse()
	if config.printVersionAndExit {
		fmt.Println(version)
		os.Exit(0)
	}
	var err error
	location, err = time.LoadLocation(config.timeZone)
	if err != nil {
		fmt.Fprintln(os.Stderr, "time zone parse error:", err)
		os.Exit(1)
	}
	if knownFormat, ok := timeFormats[config.timeFormat]; ok {
		config.timeFormat = knownFormat
	}
	if knownTemplate, ok := templates[config.template]; ok {
		config.template = knownTemplate
	}
	if knownScale, ok := colorScales[config.scaleText]; ok {
		config.scaleText = knownScale
	}
	if config.scaleText != "" {
		scale = color.ParseScale(config.scaleText)
	}
	if config.scaleLinear {
		// do nothing
	}
	if config.scaleSqrt {
		scale = color.Sqrt(scale)
	}
	if config.scaleCubert {
		scale = color.Cubert(scale)
	}
	if config.scaleSqr {
		scale = color.Sqr(scale)
	}
	if config.scaleCube {
		scale = color.Cube(scale)
	}
	if config.template != "" {
		printer = templatePrinter(config.template)
	} else {
		printer = jsonPrinter()
	}
	if config.matchRegex != "" {
		matchRegex = regexp.MustCompile(config.matchRegex)
	}
	if config.readJSON {
		tokens = newJSONStream()
	} else {
		tokens = newLineStream()
	}
	if config.matchTemplate != "" {
		config.matchTemplate = addTemplateDelimitersIfLiteral(config.matchTemplate)
		matchTemplate = &templateWithBuffer{
			template: template.Must(template.New("-match-template").Option("missingkey=zero").Parse(config.matchTemplate)),
			buffer:   bytes.NewBuffer(nil),
		}
	}
	if config.matchCondition != "" {
		config.matchCondition = addTemplateDelimitersIfLiteral(config.matchCondition)
		matchCondition = &templateWithBuffer{
			template: template.Must(template.New("-match-condition").Option("missingkey=zero").Funcs(templateFuncs).Parse(config.matchCondition)),
			buffer:   bytes.NewBuffer(nil),
		}
	}
}

type templateWithBuffer struct {
	template *template.Template
	buffer   *bytes.Buffer
}

func (t *templateWithBuffer) executeSilent(data interface{}) (string, error) {
	t.buffer.Reset()
	err := t.template.Execute(t.buffer, data)
	return t.buffer.String(), err
}

func (t *templateWithBuffer) execute(data interface{}) string {
	s, err := t.executeSilent(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "template error:", err)
	}
	return s
}

type tokenOwner interface {
	Token() *token
}

type tokenStream interface {
	tokenOwner
	AppendCurrentToBuffer()
	FlushBuffer()
	CurrentMatchText() string
	CopyCurrent() tokenStream
	Err() error
	Scan() bool
}

func main() {
	token := tokens.Token()
	first := time.Now().In(location)
	last := first
	i := uint64(0)

	for tokens.Scan() {
		now := time.Now().In(location)
		delta := now.Sub(last)
		total := now.Sub(first)

		token.DeltaSecs = delta.Seconds()
		token.DeltaNanos = delta.Nanoseconds()
		token.DeltaString = delta.String()
		token.Delta = delta
		token.TotalSecs = total.Seconds()
		token.TotalNanos = total.Nanoseconds()
		token.TotalString = total.String()
		token.Total = total
		token.TimeSecs = now.Unix()
		token.TimeNanos = now.UnixNano()
		token.TimeString = now.Format(config.timeFormat)
		token.Time = now

		token.I = i
		token.MatchText = tokens.CurrentMatchText()

		matchRegexDefined := matchRegex != nil
		matchConditionDefined := matchCondition != nil
		matchDefined := matchRegexDefined || matchConditionDefined
		printToken := !matchDefined || !config.buffer

		matches := matchDefined
		if matchRegexDefined {
			matches = matches && matchRegex.MatchString(token.MatchText)
		}
		if matchConditionDefined {
			result, _ := matchCondition.executeSilent(tokens)
			matches = matches && strings.TrimSpace(result) == "true"
		}

		resetStopwatch := !matchDefined || matches

		if printToken {
			print(tokens)
		}
		if matches {
			currentCopy := tokens.CopyCurrent()
			currentCopy.Token().Start = nil // Prevent nested .start.start.start blow-up
			token.Start = currentCopy
			if config.buffer {
				tokens.FlushBuffer()
			}
		}
		if !printToken {
			tokens.AppendCurrentToBuffer()
		}
		if resetStopwatch {
			last = now
		}
		i++
	}

	if config.buffer {
		tokens.FlushBuffer()
	}

	if err := tokens.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
		os.Exit(1)
	}
}
