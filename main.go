package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/mackerel-plugin-log-incr-rate/followparser"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Version by Makefile
var Version string

type cmdOpts struct {
	LogFile     string `long:"log-file" description:"path to log file calculate lines increased" required:"true"`
	BaseLogFile string `long:"base-log-file" description:"path to base log file count lines" required:"true"`
	KeyPrefix   string `long:"key-prefix" description:"Metric key prefix" required:"true"`
	Version     bool   `short:"v" long:"version" description:"Show version"`
}

type simpleCounter struct {
	total    float64
	duration float64
}

func (lc *simpleCounter) Parse(b []byte) error {
	lc.total = lc.total + 1
	return nil
}

func (lc *simpleCounter) Display(duration float64) {
	lc.duration = duration
}

func (lc *simpleCounter) GetTotal() float64 {
	return lc.total
}

func (lc *simpleCounter) GetDuration() float64 {
	return lc.duration
}

func getStats(opts cmdOpts, logger *zap.Logger) error {
	logCounter := &simpleCounter{}
	baseLogCounter := &simpleCounter{}
	var g errgroup.Group

	g.Go(func() error {
		return followparser.Parse("incr-rate-log", opts.LogFile, logCounter, logger)
	})

	g.Go(func() error {
		return followparser.Parse("incr-rate-base", opts.BaseLogFile, baseLogCounter, logger)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	now := uint64(time.Now().Unix())

	if logCounter.GetDuration() > 0 {
		fmt.Printf("log-incr-rate.%s_count.log\t%f\t%d\n",
			opts.KeyPrefix,
			logCounter.GetTotal()/logCounter.GetDuration(),
			now)
	}
	if baseLogCounter.GetDuration() > 0 {
		fmt.Printf("log-incr-rate.%s_count.base\t%f\t%d\n",
			opts.KeyPrefix,
			baseLogCounter.GetTotal()/baseLogCounter.GetDuration(),
			now)
	}

	if logCounter.GetDuration() > 0 && baseLogCounter.GetDuration() > 0 && baseLogCounter.GetTotal() > 0 {
		fmt.Printf("log-incr-rate.%s_rate.log\t%f\t%d\n",
			opts.KeyPrefix,
			(logCounter.GetTotal()/logCounter.GetDuration())/(baseLogCounter.GetTotal()/baseLogCounter.GetDuration()),
			now)
	}

	return nil
}

func printVersion() {
	fmt.Printf(`%s %s
Compiler: %s %s
`,
		os.Args[0],
		Version,
		runtime.Compiler,
		runtime.Version())
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := cmdOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		return 1
	}
	if opts.Version {
		printVersion()
		return 0
	}

	logger, _ := zap.NewProduction()
	err = getStats(opts, logger)
	if err != nil {
		logger.Error("getStats", zap.Error(err))
		return 1
	}
	return 0
}
