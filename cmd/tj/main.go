package main

import (
	"bufio"
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

type line struct {
	I           uint64        `json:"-"` // line number
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
	Text        string        `json:"text,omitempty"`
	StartText   string        `json:"startText,omitempty"`
	JSONText    string        `json:"jsonText,omitempty"`
	Object      interface{}   `json:"object,omitempty"`
	StartObject interface{}   `json:"startObject,omitempty"`
}

type configuration struct {
	timeFormat   string        // -timeformat="..."
	template     string        // -template="..."
	start        string        // -start="..."
	readJSON     bool          // -readjson
	jsonTemplate string        // -jsontemplate="..."
	colorScale   string        // -scale
	fast         time.Duration // -scale-fast
	slow         time.Duration // -scale-slow
	version      string
}

type printerFunc func(line *line) error

var (
	config       configuration
	printer      printerFunc
	start        *regexp.Regexp
	jsonTemplate *template.Template
	scale        color.Scale
)

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
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
}

var templates = map[string]string{
	"Time":      "{{.TimeString}} {{.Text}}",
	"TimeDelta": "{{.TimeString}} +{{.DeltaNanos}} {{.Text}}",
	"Delta":     "{{.DeltaNanos}} {{.Text}}",
	"ColorText": "{{color .}}{{.Text}}{{reset}}",
	"Color":     "{{color .}}█{{reset}} {{.Text}}",
}

var colorScales = map[string]string{
	"GreenToRed":       "#0F0 -> #F00",
	"BlueToRed":        "#00F -> #F00",
	"CyanToRed":        "#0FF -> #F00",
	"WhiteToRed":       "#FFF -> #F00",
	"WhiteToPurple":    "#FFF -> #F700FF",
	"BlackToRed":       "#000 -> #F00",
	"BlackToPurple":    "#000 -> #F700FF",
	"WhiteToBlueToRed": "#FFF -> #00F -> #F00",
}

var templateFuncs = template.FuncMap{
	"color": foregroundColor,
	"reset": func() string { return color.Reset },
}

func foregroundColor(line *line) string {
	c := float64(line.DeltaNanos-int64(config.fast)) / float64(config.slow-config.fast)
	return color.Foreground(scale(c))
}

func jsonPrinter() printerFunc {
	enc := json.NewEncoder(os.Stdout)
	return func(line *line) error {
		return enc.Encode(line)
	}
}

func templatePrinter(t string) printerFunc {
	template := template.Must(template.New("-template").Funcs(templateFuncs).Option("missingkey=zero").Parse(t))
	newline := []byte("\n")
	return func(line *line) error {
		err := template.Execute(os.Stdout, line)
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

func init() {
	flag.StringVar(&config.template, "template", "", templatesHelp())
	flag.StringVar(&config.timeFormat, "timeformat", "RFC3339", timeFormatsHelp())
	flag.StringVar(&config.start, "start", "", "a regex pattern. if given, only lines matching it (re)start the stopwatch")
	flag.BoolVar(&config.readJSON, "readjson", false, "parse each stdin line as JSON")
	flag.StringVar(&config.jsonTemplate, "jsontemplate", "", "go template, used to extract text from json input. implies -readjson")
	flag.StringVar(&config.colorScale, "scale", "BlueToRed", colorScalesHelp())
	flag.DurationVar(&config.fast, "scale-fast", 100*time.Millisecond, "the lower bound for the color scale")
	flag.DurationVar(&config.slow, "scale-slow", 2*time.Second, "the upper bound for the color scale")
	flag.Parse()
	if knownFormat, ok := timeFormats[config.timeFormat]; ok {
		config.timeFormat = knownFormat
	}
	if knownTemplate, ok := templates[config.template]; ok {
		config.template = knownTemplate
	}
	if knownScale, ok := colorScales[config.colorScale]; ok {
		config.colorScale = knownScale
	}
	if config.colorScale != "" {
		scale = color.NewScale(color.Parse(config.colorScale))
	}
	if config.template != "" {
		printer = templatePrinter(config.template)
	} else {
		printer = jsonPrinter()
	}
	if config.start != "" {
		start = regexp.MustCompile(config.start)
	}
	if config.jsonTemplate != "" {
		config.readJSON = true
		jsonTemplate = template.Must(template.New("-jsontemplate").Option("missingkey=zero").Parse(config.jsonTemplate))
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	line := line{Time: time.Now()}
	first := line.Time
	last := line.Time
	i := uint64(0)
	b := bytes.NewBuffer(nil)
	for scanner.Scan() {
		now := time.Now()
		delta := now.Sub(last)
		total := now.Sub(first)
		line.DeltaSecs = delta.Seconds()
		line.DeltaNanos = delta.Nanoseconds()
		line.DeltaString = delta.String()
		line.Delta = delta
		line.TotalSecs = total.Seconds()
		line.TotalNanos = total.Nanoseconds()
		line.TotalString = total.String()
		line.Total = total
		line.TimeSecs = now.Unix()
		line.TimeNanos = now.UnixNano()
		line.TimeString = now.Format(config.timeFormat)
		line.Time = now
		line.Text = scanner.Text()
		line.I = i
		match := line.Text
		if config.readJSON {
			line.Object = new(interface{})
			if err := json.Unmarshal([]byte(line.Text), &line.Object); err != nil {
				fmt.Fprintln(os.Stderr, "JSON parse error:", err)
			}
			if jsonTemplate != nil {
				b.Reset()
				if err := jsonTemplate.Execute(b, line.Object); err != nil {
					fmt.Fprintln(os.Stderr, "template error:", err)
				}
				line.JSONText = b.String()
				match = line.JSONText
			}
		}
		if err := printer(&line); err != nil {
			fmt.Fprintln(os.Stderr, "output error:", err)
		}
		if start != nil {
			if start.MatchString(match) {
				last = now
				line.StartText = line.Text
				line.StartObject = line.Object
			}
		} else {
			last = now
		}
		i++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
		os.Exit(1)
	}
}
