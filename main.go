package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/ruggi/quando"
	"github.com/ruggi/quando/rules/en"
	"github.com/satori/uuid"
)

const (
	subtitleFormat    = "Mon, Jan _2 2006 at 15:04"
	applescriptFormat = "2006-01-02 15:04:05"
	defaultDuration   = time.Hour
	updateJobName     = "checkForUpdate"
	repo              = "ruggi/calfred"
)

var (
	wf            *aw.Workflow
	iconAvailable = &aw.Icon{Value: "update-available.png"}

	flags struct {
		checkUpdates bool
	}
)

func init() {
	flag.BoolVar(&flags.checkUpdates, "check", false, "check for a new version")
	wf = aw.New(update.GitHub(repo))
}

func parseQuery(p *quando.Parser, s string) (time.Time, time.Time, string, error) {
	r, err := p.Parse(s)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	dur := r.Duration
	if dur == 0 {
		dur = defaultDuration
	}
	return r.Time, r.Time.Add(dur), r.Text, nil
}

func run() {
	wf.Args()
	flag.Parse()

	if flags.checkUpdates {
		wf.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")

		cmd := exec.Command(os.Args[0], "-check")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	if wf.UpdateAvailable() {
		wf.Configure(aw.SuppressUIDs(true))
		wf.NewItem("Workflow update available").
			Subtitle("↩ to install").
			Autocomplete("workflow:update").
			Valid(false).
			Icon(iconAvailable)
	}

	query := strings.TrimSpace(strings.Join(wf.Args(), " "))
	parser := quando.NewParser(quando.WithRules(en.Rules...))
	start, end, what, err := parseQuery(parser, query)
	if err != nil {
		wf.NewItem(query)
	} else {
		wf.NewItem(what).
			Subtitle(start.Format(subtitleFormat)).
			Arg("event").
			Var("what", what).
			Var("when_start", start.Format(applescriptFormat)).
			Var("when_end", end.Format(applescriptFormat)).
			Valid(true).
			Icon(aw.IconWorkflow).
			UID(uuid.NewV4().String())
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
