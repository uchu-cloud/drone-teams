// Copyright (c) 2020, the Drone Plugins project authors.
// Please see the AUTHORS file for details. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file.

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Settings for the plugin.
type Settings struct {
	Webhook     string
	Status      string
	CustomFacts cli.StringSlice
	Logs        Logs
}

type Logs struct {
	OnError   bool
	AuthToken string
}

type BuildInfo struct {
	Number int
	Status string
	Stages []BuildStage
}

type BuildStage struct {
	Number   int
	Name     string
	Status   string
	ExitCode int
	Steps    []BuildStep
}

type BuildStep struct {
	Number   int
	Name     string
	Status   string
	ExitCode int
}

type BuildLog struct {
	Proc string
	Pos  int
	Out  string
}

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	// Verify the webhook endpoint
	if p.settings.Webhook == "" {
		// If webhook is undefined, check if the ${DRONE_BRANCH}_teams_webhook env var is defined.
		branchWebhook := fmt.Sprintf("%s_teams_webhook", os.Getenv("DRONE_BRANCH"))
		if os.Getenv(branchWebhook) == "" {
			return fmt.Errorf("no webhook endpoint provided")
		}
		// Set webhook setting to ${DRONE_BRANCH}_teams_webhook
		p.settings.Webhook = os.Getenv(branchWebhook)
	}

	// If the plugin status setting is defined, use that as the build status
	if p.settings.Status == "" {
		p.settings.Status = p.pipeline.Build.Status
	}

	return nil
}

// Execute provides the implementation of the plugin.
func (p *Plugin) Execute() error {

	// Default card color is green
	themeColor := "96FF33"

	// Create list of card facts
	facts := []MessageCardSectionFact{
		{
			Name:  "Build Number",
			Value: fmt.Sprintf("%d", p.pipeline.Build.Number),
		},
		{
			Name: "Git Author",
			Value: fmt.Sprintf("%s \"%s\" (%s)",
				p.pipeline.Commit.Author.Name,
				p.pipeline.Commit.Author.Email,
				p.pipeline.Commit.Author.Username,
			),
		},
	}

	// Check for commit message
	if len(p.pipeline.Commit.Message.String()) > 0 {
		facts = append(facts, MessageCardSectionFact{
			Name:  "Commit Message",
			Value: p.pipeline.Commit.Message.String(),
		})
	}

	// Add custom facts supplied by the user
	for _, fact := range p.settings.CustomFacts.Value() {

		factKV := strings.Split(fact, ":")

		if len(factKV) < 2 {
			continue
		}

		card := MessageCardSectionFact{
			Name:  factKV[0],
			Value: factKV[1],
		}

		facts = append(facts, card)
	}

	// Create list of actions
	actions := []OpenURIAction{
		{
			Type: "OpenUri",
			Name: "Open repository",
			Targets: []OpenURITarget{
				{
					OS:  "default",
					URI: p.pipeline.Repo.Link,
				},
			},
		},
	}

	// If commit link is not null add commit link fact to card
	// Only load this button for Push and Pull Requests, otherwise won't make sense
	switch p.pipeline.Build.Event {
	case "push", "pull_request":
		link := ""
		if p.pipeline.Commit.Link != "" {
			link = p.pipeline.Commit.Link
		} else if commitLink, present := os.LookupEnv("DRONE_COMMIT_LINK"); present && commitLink != "" {
			link = commitLink
		}

		if len(link) > 0 {
			actions = append(actions, OpenURIAction{
				Type: "OpenUri",
				Name: "Open commit diff",
				Targets: []OpenURITarget{
					{
						OS:  "default",
						URI: link,
					},
				},
			})
		}
	case "tag":
		actions = append(actions, OpenURIAction{
			Type: "OpenUri",
			Name: "Open tag list",
			Targets: []OpenURITarget{
				{
					OS:  "default",
					URI: fmt.Sprintf("%s/tags", p.pipeline.Repo.Link),
				},
			},
		})
	}

	// If build has failed, change color to red and add failed step fact
	if p.settings.Status == "failure" {
		themeColor = "FF5733"

		facts = append(facts, MessageCardSectionFact{
			Name:  "Failed Build Steps",
			Value: strings.Join(p.pipeline.Build.FailedSteps, " "),
		})

		if p.settings.Logs.OnError && len(p.settings.Logs.AuthToken) > 0 {
			logs, err := p.assembleLogs()
			if err == nil && logs != nil && len(logs) > 0 {
				facts = append(facts, logs...)
			}
		}

		actions = append(actions, OpenURIAction{
			Type: "OpenUri",
			Name: "Open build pipeline",
			Targets: []OpenURITarget{
				{
					OS: "default",
					URI: fmt.Sprintf("%s://%s/%s/%d",
						p.pipeline.System.Proto,
						p.pipeline.System.Host,
						p.pipeline.Repo.Slug,
						p.pipeline.Build.Number,
					),
				},
			},
		})

		// If the plugin status setting is defined and is "building", set the color to blue
	} else if p.settings.Status == "building" {
		themeColor = "002BFF"
	}

	// Create rich message card body
	card := MessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: themeColor,
		Summary:    p.pipeline.Repo.Slug,
		Sections: []MessageCardSection{
			{
				Markdown:         false,
				Facts:            facts,
				ActivityImage:    "https://github.com/uchugroup/drone-teams/raw/master/drone.png",
				ActivityTitle:    fmt.Sprintf("%s (%s%s)", p.pipeline.Repo.Slug, p.pipeline.Build.Branch, p.pipeline.Build.Tag),
				ActivitySubtitle: strings.ToUpper(p.settings.Status),
				ActivityText: fmt.Sprintf("%s %s %s (build time %s)",
					p.pipeline.Build.Event,
					p.pipeline.Build.DeployTo,
					p.pipeline.Commit.Ref,
					time.Since(p.pipeline.Build.Created).Round(time.Second).String(),
				),
			},
		},
		PotentialAction: actions,
	}

	log.Info("Generated card: ", card)

	// MS teams webhook post
	jsonValue, _ := json.Marshal(card)
	_, err := http.Post(p.settings.Webhook, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error("Failed to send request to teams webhook")
		return err
	}

	return nil
}

func (p *Plugin) assembleLogs() ([]MessageCardSectionFact, error) {

	logs := make([]MessageCardSectionFact, 0)

	if len(p.pipeline.Build.FailedSteps) == 0 {
		return logs, nil
	}

	// Get build info
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/api/repos/%s/%s/builds/%d",
		p.pipeline.System.Host,
		p.pipeline.Repo.Owner,
		p.pipeline.Repo.Name,
		p.pipeline.Build.Number,
	), nil)
	if err != nil {
		log.Error("Failed to create build info request")
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.settings.Logs.AuthToken))

	resp, err := p.network.Client.Do(req)

	if err != nil {
		log.Errorf("Failed to get build info for %s", req.URL.String())
		return nil, err
	} else if resp.StatusCode >= 400 {
		log.Errorf("Failed to get build info for %s with status %s", req.URL.String(), resp.Status)
		return nil, fmt.Errorf("server error %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read build info")
		return nil, err
	}

	buildInfo := new(BuildInfo)
	if err = json.Unmarshal(data, buildInfo); err != nil {
		log.Error("Failed to parse build info")
		return nil, err
	}

	if buildInfo.Status == "success" {
		return logs, nil
	}

	// Loop all stages
	for _, buildStage := range buildInfo.Stages {

		if buildStage.Status == "success" {
			continue
		}

		for _, buildStep := range buildStage.Steps {

			if buildStep.ExitCode == 0 {
				continue
			}

			// Get logs
			req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/api/repos/%s/%s/builds/%d/logs/%d/%d",
				p.pipeline.System.Host,
				p.pipeline.Repo.Owner,
				p.pipeline.Repo.Name,
				p.pipeline.Build.Number,
				buildStage.Number,
				buildStage.Number,
			), nil)
			if err != nil {
				log.Error("Failed to create build logs request")
				return nil, err
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.settings.Logs.AuthToken))

			resp, err = p.network.Client.Do(req)

			if err != nil {
				log.Errorf("Failed to get log for %s/%s build %d stage %s step %s",
					p.pipeline.Repo.Owner,
					p.pipeline.Repo.Name,
					p.pipeline.Build.Number,
					buildStage.Name,
					buildStage.Name,
				)
				return nil, err
			} else if resp.StatusCode >= 400 {
				log.Errorf("Failed to get log for %s/%s build %d stage %s step %s",
					p.pipeline.Repo.Owner,
					p.pipeline.Repo.Name,
					p.pipeline.Build.Number,
					buildStage.Name,
					buildStage.Name,
				)
				return nil, fmt.Errorf("server error %s", resp.Status)
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("Failed to read log for %s/%s build %d stage %s step %s",
					p.pipeline.Repo.Owner,
					p.pipeline.Repo.Name,
					p.pipeline.Build.Number,
					buildStage.Name,
					buildStage.Name,
				)
				return nil, err
			}

			var buildLogs []BuildLog
			if err = json.Unmarshal(data, &buildLogs); err != nil {
				log.Errorf("Failed to parse log for %s/%s build %d stage %s step %s",
					p.pipeline.Repo.Owner,
					p.pipeline.Repo.Name,
					p.pipeline.Build.Number,
					buildStage.Name,
					buildStage.Name,
				)
				return nil, err
			}

			// Compile logs
			logValue := make([]string, 0)
			for _, buildLog := range buildLogs {
				logValue = append(logValue, fmt.Sprintf("Command #%d: %s\nResult: %s",
					buildLog.Pos,
					buildLog.Proc,
					buildLog.Out,
				))
			}

			var log MessageCardSectionFact
			log.Name = fmt.Sprintf("Log for %s/%s", buildStage.Name, buildStep.Name)
			log.Value = strings.Join(logValue, "\n")

			logs = append(logs, log)
		}
	}

	return logs, nil
}
