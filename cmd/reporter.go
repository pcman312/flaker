package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/k0kubun/go-ansi"
	"github.com/vrecan/life"
)

type statusGetter interface {
	status() status
}

// reporter reports results of the flaker run both at run time as well as the final results
type reporter struct {
	*life.Life

	status   statusGetter
	tickRate time.Duration
}

func newReporter(status statusGetter, tickRate time.Duration) (reporter, error) {
	if status == nil {
		return reporter{}, fmt.Errorf("status is nil")
	}

	r := reporter{
		Life:     life.NewLife(),
		status:   status,
		tickRate: tickRate,
	}
	r.SetRun(r.run)
	return r, nil
}

func (r reporter) run() {
	start := time.Now()
	ticker := time.NewTicker(r.tickRate)
LOOP:
	for {
		select {
		case <-r.Done:
			break LOOP
		case <-ticker.C:
			r.report(start)
		}
	}
	r.reportFinal(start)
}

const (
	dateTimeFormat = "2006-01-02 15:04:05"
)

func (r reporter) report(start time.Time) {
	status := r.status.status()
	now := time.Now()
	duration := now.Sub(start)

	clearLine()
	percentSuccessful := "<no runs>"
	if status.runs > 0 {
		percentSuccessful = fmt.Sprintf("%5.2f%% success", float64(status.successful)/float64(status.runs)*100.0)
	}
	fmt.Printf("%s | %6s | Runs: %d (%s:%s) (%s)",
		now.Format(dateTimeFormat),
		fmtDuration(duration.Truncate(time.Millisecond)),
		status.runs,
		color.GreenString("%d", status.successful),
		color.RedString("%d", status.failed),
		percentSuccessful,
	)
}

func fmtDuration(dur time.Duration) string {
	h := dur / time.Hour
	dur -= h * time.Hour

	m := dur / time.Minute
	dur -= m * time.Minute

	s := dur / time.Second
	dur -= s * time.Second

	ms := dur / time.Millisecond

	if h == 0 {
		return fmt.Sprintf("%2dm%2ds.%03d", m, s, ms)
	}
	return fmt.Sprintf("%dh%2dm%2ds.%03d", h, m, s, ms)
}

func (r reporter) reportFinal(start time.Time) {
	r.report(start)
	fmt.Println()
}

const (
	clearToEndOfLine        = 0
	clearToBeginaningOfLine = 1
	clearEntireLine         = 2
)

func clearLine() {
	ansi.EraseInLine(clearEntireLine)
	ansi.CursorHorizontalAbsolute(0)
}
