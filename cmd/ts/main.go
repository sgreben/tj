package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"text/template"
	"time"
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
	Total       string        `json:"total,omitempty"`
	Text        string        `json:"text,omitempty"`
	Previous    string        `json:"previous,omitempty"`
	Start       string        `json:"start,omitempty"`
}

type configuration struct {
	timeFormat string // -timeformat="..."
	template   string // -template="..."
	plain      bool   // -plain
	start      string // -start="..."
	version    string
	previous   bool
}

var config = configuration{}

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

type printerFunc func(line *line) error

var printer printerFunc

func jsonPrinter() printerFunc {
	enc := json.NewEncoder(os.Stdout)
	return func(line *line) error {
		return enc.Encode(line)
	}
}

func templatePrinter(t string) printerFunc {
	template := template.Must(template.New("").Parse(t))
	newline := []byte("\n")
	return func(line *line) error {
		err := template.Execute(os.Stdout, line)
		os.Stdout.Write(newline)
		return err
	}
}

func timeFormatsHelp() string {
	help := "either a go time format string or one of the predefined format names (https://golang.org/pkg/time/#pkg-constants)\n"
	buf := bytes.NewBuffer([]byte(help))
	for name, format := range timeFormats {
		fmt.Fprintln(buf, "\t", name, "-", format)
	}
	return buf.String()
}

func init() {
	flag.StringVar(&config.template, "template", "", "go template (https://golang.org/pkg/text/template)")
	flag.StringVar(&config.timeFormat, "timeformat", "RFC3339", timeFormatsHelp())
	flag.BoolVar(&config.plain, "plain", false, "-template='{{.Time}} +{{.DeltaNanos}} {{.Text}}'")
	flag.StringVar(&config.start, "start", "", "a regex pattern. if given, only lines matching it (re)start the stopwatch")
	flag.BoolVar(&config.previous, "previous", false, "include previous line")
	flag.Parse()
	if knownFormat, ok := timeFormats[config.timeFormat]; ok {
		config.timeFormat = knownFormat
	}
	if config.plain {
		config.template = "{{.Time}} +{{.DeltaNanos}} {{.Text}}"
	}
	if config.template != "" {
		printer = templatePrinter(config.template)
	} else {
		printer = jsonPrinter()
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	now := time.Now()
	line := line{}
	last := now
	first := now
	previous := ""
	i := uint64(0)
	var start *regexp.Regexp
	if config.start != "" {
		start = regexp.MustCompile(config.start)
	}
	for scanner.Scan() {
		now = time.Now()
		delta := now.Sub(last)
		total := now.Sub(first)
		line.DeltaSecs = delta.Seconds()
		line.DeltaNanos = delta.Nanoseconds()
		line.Delta = delta
		line.DeltaString = delta.String()
		line.TotalSecs = total.Seconds()
		line.TotalNanos = total.Nanoseconds()
		line.Total = total.String()
		line.TimeSecs = now.Unix()
		line.TimeNanos = now.UnixNano()
		line.Time = now
		line.TimeString = now.Format(config.timeFormat)
		line.Text = scanner.Text()
		line.I = i
		if config.previous {
			line.Previous = previous
			previous = line.Text
		}
		if err := printer(&line); err != nil {
			fmt.Fprintln(os.Stderr, "output error:", err)
		}
		if start != nil && start.MatchString(line.Text) {
			last = now
			line.Start = line.Text
		}
		i++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
		os.Exit(1)
	}
}
