package main

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"runtime"
	"strconv"
)

const (
	GCRegexpGo14 = `gc\d+\(\d+\): ([\d.]+\+?)+ us, \d+ -> (?P<Heap1>\d+) MB, \d+ \(\d+-\d+\) objects,( \d+ goroutines,)? \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`
	GCRegexpGo15 = `gc \d+ @[\d.]+s \d+%: ([\d.]+\+?)+ ms clock, ([\d.+\/]+?)+ ms cpu, \d+->\d+->\d+ MB, (?P<Heap1>\d+) MB goal, \d+ P`
	SCVGRegexp   = `scvg\d+: inuse: (?P<inuse>\d+), idle: (?P<idle>\d+), sys: (?P<sys>\d+), released: (?P<released>\d+), consumed: (?P<consumed>\d+) \(MB\)`
)

var (
	gcrego14 = regexp.MustCompile(GCRegexpGo14)
	gcrego15 = regexp.MustCompile(GCRegexpGo15)
	scvgre   = regexp.MustCompile(SCVGRegexp)
)

type Parser struct {
	reader      io.Reader
	GcChan      chan *gctrace
	ScvgChan    chan *scvgtrace
	NoMatchChan chan string

	gcRegexp   *regexp.Regexp
	scvgRegexp *regexp.Regexp
}

func NewParser(r io.Reader) *Parser {
	gcChan := make(chan *gctrace, 1)
	scvgChan := make(chan *scvgtrace, 1)
	noMatchChan := make(chan string, 1)

	return &Parser{
		reader:      r,
		GcChan:      gcChan,
		ScvgChan:    scvgChan,
		NoMatchChan: noMatchChan,
	}
}

func (p *Parser) Run() {
	sc := bufio.NewScanner(p.reader)

	if p.gcRegexp == nil {
		// Set regexp based on Golang version
		if runtime.Version() == "go1.5" {
			p.gcRegexp = gcrego15
		} else {
			p.gcRegexp = gcrego14
		}
	}

	for sc.Scan() {
		line := sc.Text()

		if result := p.gcRegexp.FindStringSubmatch(line); result != nil {
			p.GcChan <- parseGCTrace(p.gcRegexp, result)
			continue
		}

		if result := scvgre.FindStringSubmatch(line); result != nil {
			p.ScvgChan <- parseSCVGTrace(result)
			continue
		}

		p.NoMatchChan <- line
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseGCTrace(gcre *regexp.Regexp, matches []string) *gctrace {
	matchMap := getMatchMap(gcre, matches)

	return &gctrace{
		Heap1: matchMap["Heap1"],
	}
}

func parseSCVGTrace(matches []string) *scvgtrace {
	matchMap := getMatchMap(scvgre, matches)

	return &scvgtrace{
		inuse:    matchMap["inuse"],
		idle:     matchMap["idle"],
		sys:      matchMap["sys"],
		released: matchMap["released"],
		consumed: matchMap["consumed"],
	}
}

// Transform our matches in a readable hash map.
//
// The resulting hash map will be something like { "Heap1": 123 }
func getMatchMap(re *regexp.Regexp, matches []string) map[string]int64 {
	matchingNames := re.SubexpNames()
	matchMap := map[string]int64{}
	for i, value := range matches {
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			// Happens on first element of range and any matching parenthesis
			// that includes non-parseable string
			//
			// For example a matching array would contain:
			// [ "scvg1: inuse:3 ..." "3" ]
			continue
		}
		matchMap[matchingNames[i]] = intVal
	}
	return matchMap
}
