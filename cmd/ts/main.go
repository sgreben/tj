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
	TotalString string        `json:"total,omitempty"`
	Total       time.Duration `json:"-"`
	Text        string        `json:"text,omitempty"`
	Start       string        `json:"start,omitempty"`
}

type configuration struct {
	timeFormat string // -timeformat="..."
	template   string // -template="..."
	plain      bool   // -plain
	start      string // -start="..."
	version    string
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

var start *regexp.Regexp

func init() {
	flag.StringVar(&config.template, "template", "", "go template (https://golang.org/pkg/text/template)")
	flag.StringVar(&config.timeFormat, "timeformat", "RFC3339", timeFormatsHelp())
	flag.BoolVar(&config.plain, "plain", false, "-template='{{.TimeString}} +{{.DeltaNanos}} {{.Text}}'")
	flag.StringVar(&config.start, "start", "", "a regex pattern. if given, only lines matching it (re)start the stopwatch")
	flag.Parse()
	if knownFormat, ok := timeFormats[config.timeFormat]; ok {
		config.timeFormat = knownFormat
	}
	if config.plain {
		config.template = "{{.TimeString}} +{{.DeltaNanos}} {{.Text}}"
	}
	if config.template != "" {
		printer = templatePrinter(config.template)
	} else {
		printer = jsonPrinter()
	}
	if config.start != "" {
		start = regexp.MustCompile(config.start)
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	line := line{Time: time.Now()}
	first := line.Time
	last := line.Time
	i := uint64(0)
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
		if err := printer(&line); err != nil {
			fmt.Fprintln(os.Stderr, "output error:", err)
		}
		if start != nil {
			if start.MatchString(line.Text) {
				last = now
				line.Start = line.Text
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
