// Copyright (c) 2020, the Drone Plugins project authors.
// Please see the AUTHORS file for details. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file.

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
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
					time.Since(p.pipeline.Build.Created).Truncate(time.Second).String(),
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
