// Copyright (c) 2020, the Drone Plugins project authors.
// Please see the AUTHORS file for details. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file.

package main

import (
	"github.com/uchugroup/drone-teams/plugin"
	"github.com/urfave/cli/v2"
)

// settingsFlags has the cli.Flags for the plugin.Settings.
func settingsFlags(settings *plugin.Settings) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "webhook",
			Usage:       "MS teams connector webhook endpoint",
			EnvVars:     []string{"PLUGIN_WEBHOOK"},
			Destination: &settings.Webhook,
		},
		&cli.StringFlag{
			Name:        "status",
			Usage:       "Overwrite the status value",
			EnvVars:     []string{"PLUGIN_STATUS"},
			Destination: &settings.Status,
		},
		&cli.StringSliceFlag{
			Name:        "facts",
			Usage:       "Add custom facts to the card",
			EnvVars:     []string{"PLUGIN_FACTS"},
			Destination: &settings.CustomFacts,
		},
		&cli.BoolFlag{
			Name:        "logs_on_error",
			Usage:       "Display logs on error",
			EnvVars:     []string{"PLUGIN_LOGS_ON_ERROR"},
			Destination: &settings.Logs.OnError,
		},
		&cli.StringFlag{
			Name:        "logs_auth_token",
			Usage:       "Auth token to read the logs",
			EnvVars:     []string{"PLUGIN_LOGS_AUTH_TOKEN"},
			Destination: &settings.Logs.AuthToken,
		},
	}
}
