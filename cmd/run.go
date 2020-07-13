package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pcman312/flaker/cmd/results"
	"github.com/pcman312/flaker/cmd/types"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var (
	durationFlag = &cli.DurationFlag{
		Name:    "d",
		Aliases: []string{"duration"},
		Usage: "How long to run the commands for. This is the minimum run time as this will wait " +
			"until currently running commands finish before returning",
		Required:    true,
		DefaultText: "",
	}
	parallelFlag = &cli.IntFlag{
		Name:        "p",
		Aliases:     []string{"parallel"},
		Usage:       "Number of concurrent runs of the command",
		DefaultText: "1",
		Value:       1,
	}
	refreshFlag = &cli.DurationFlag{
		Name:        "r",
		Aliases:     []string{"refresh"},
		Usage:       "How frequently to refresh the output. This takes a number followed by a unit such as '1m', '30s', '1h'",
		DefaultText: "1s",
		Value:       1 * time.Second,
	}
	resultsFileFlag = &cli.StringFlag{
		Name:    "o",
		Aliases: []string{"output_file"},
		Usage:   "File to write results to in JSON format",
	}
	rootCmdFlag = &cli.StringSliceFlag{
		Name:        "root_command",
		Usage:       "Specifies the underlying root command to use to enable piping and redirection",
		Value:       cli.NewStringSlice("bash", "-c"),
		DefaultText: "bash -c",
	}
	stopOnFailure = &cli.BoolFlag{
		Name:  "stop_on_failure",
		Usage: "Stops executions if any executions fail",
	}
)

func Run(args []string) (code int) {
	app := cli.App{
		Name:      "flaker",
		Usage:     "Repeatedly run a command (usually a test) to check for flakiness",
		UsageText: "flaker [options] -- [command]",
		Version:   "0.0.3",
		Commands:  nil,
		Flags: []cli.Flag{
			durationFlag,
			parallelFlag,
			refreshFlag,
			resultsFileFlag,
			rootCmdFlag,
			stopOnFailure,
		},
		Action: run,
	}
	err := app.Run(args)
	if err != nil {
		fmt.Printf("%s: %s\n", color.RedString("ERROR"), err)
		return 1
	}
	return 0
}

func run(flagCtx *cli.Context) error {
	args := flagCtx.Args().Slice()

	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	rootCmd := flagCtx.StringSlice(rootCmdFlag.Name)
	if rootCmd[0] == "" {
		rootCmd = []string{"bash", "-c"}
	}

	shCmd := shellCommand{
		name: rootCmd[0],
		args: append(rootCmd[1:], strings.Join(args, " ")),
	}

	duration := flagCtx.Duration(durationFlag.Name)
	numRoutines := flagCtx.Int(parallelFlag.Name)
	refreshRate := flagCtx.Duration(refreshFlag.Name)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	startables := []starter{}

	resultsChan := make(chan types.Results, numRoutines)

	// Set up processing
	runners := []runner{}
	for i := 0; i < numRoutines; i++ {
		runner, err := newRunner(shCmd, resultsChan)
		if err != nil {
			return fmt.Errorf("unable to create runner: %w", err)
		}
		runners = append(runners, runner)
		startables = append(startables, runner)
	}

	var stopFunc func()
	failed := false

	if flagCtx.Bool(stopOnFailure.Name) {
		stopFunc = func() {
			cancel()
			stop(runners...)
			failed = true
		}
	}

	stats := newRunStats()

	var resultsWriter io.Writer
	resultsFolder := flagCtx.String(resultsFileFlag.Name)
	if resultsFolder != "" {
		path, err := filepath.Abs(resultsFolder)
		if err != nil {
			path = resultsFolder
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", path, err)
		}
		resultsWriter = f
	}

	rl, err := results.NewListener(
		results.ResultsChan(resultsChan),
		results.Stats(stats),
		results.Writer(resultsWriter),
		results.StopOnFailure(stopFunc),
	)
	if err != nil {
		return fmt.Errorf("unable to create results listener: %w", err)
	}
	startables = append(startables, rl)

	reporter, err := newReporter(stats, refreshRate)
	if err != nil {
		return fmt.Errorf("unable to create reporter: %w", err)
	}
	startables = append(startables, reporter)

	start(startables...)

	<-ctx.Done()

	stop(runners...)

	rl.Close()
	reporter.Close()

	if failed {
		return fmt.Errorf("command failed")
	}
	return nil
}

type starter interface {
	Start()
}

func start(startable ...starter) {
	for _, s := range startable {
		s.Start()
	}
}

func stop(runners ...runner) {
	wg := &sync.WaitGroup{}
	for _, c := range runners {
		wg.Add(1)
		go func() {
			c.Close()
			wg.Done()
		}()
	}
	wg.Wait()
}
