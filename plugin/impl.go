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

	// Create list of actions
	actions := []OpenURIAction{
		{
			Name: "Open repository",
			Targets: []OpenURITarget{
				{
					OS:  "default",
					URI: p.pipeline.Repo.Link,
				},
			},
		},
	}

	// Create list of card facts
	facts := []MessageCardSectionFact{
		{
			Name:  "Build Number",
			Value: fmt.Sprintf("%d", p.pipeline.Build.Number),
		},
		{
			Name:  "Git Author",
			Value: fmt.Sprintf("%s (%s)", p.pipeline.Commit.Author, p.pipeline.Commit.AuthorEmail),
		},
		{
			Name:  "Commit Message",
			Value: p.pipeline.Commit.Message,
		},
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

	// If commit link is not null add commit link fact to card
	if p.pipeline.Commit.Link != "" {
		actions = append(actions, OpenURIAction{
			Name: "Open commit diff",
			Targets: []OpenURITarget{
				{
					OS:  "default",
					URI: p.pipeline.Commit.Link,
				},
			},
		})
	} else if commitLink, present := os.LookupEnv("DRONE_COMMIT_LINK"); present && commitLink != "" {
		actions = append(actions, OpenURIAction{
			Name: "Open commit diff",
			Targets: []OpenURITarget{
				{
					OS:  "default",
					URI: commitLink,
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
				ActivityImage:    "https://github.com/uchugroup/drone-teams/raw/master/drone.png",
				ActivityTitle:    fmt.Sprintf("%s (%s)", p.pipeline.Repo.Slug, p.pipeline.Build.Branch),
				ActivitySubtitle: strings.ToUpper(p.settings.Status),
				// ActivityText:     fmt.Sprintf("Start time: %s)", p.pipeline.Build.Started.String()),
				Markdown:        false,
				Facts:           facts,
				PotentialAction: actions,
			},
		},
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
