// Copyright 2015 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"strings"
	"time"
)

var (
	// DefaultWebhookConfig defines default values for Webhook configurations.
	DefaultWebhookConfig = WebhookConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: true,
		},
	}

	// DefaultEmailConfig defines default values for Email configurations.
	DefaultEmailConfig = EmailConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: false,
		},
		HTML: `{{ template "email.default.html" . }}`,
	}

	// DefaultEmailSubject defines the default Subject header of an Email.
	DefaultEmailSubject = `{{ template "email.default.subject" . }}`

	// DefaultPagerdutyConfig defines default values for PagerDuty configurations.
	DefaultPagerdutyConfig = PagerdutyConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: true,
		},
		Description: `{{ template "pagerduty.default.description" .}}`,
		Client:      `{{ template "pagerduty.default.client" . }}`,
		ClientURL:   `{{ template "pagerduty.default.clientURL" . }}`,
		Details: map[string]string{
			"firing":       `{{ template "pagerduty.default.instances" .Alerts.Firing }}`,
			"resolved":     `{{ template "pagerduty.default.instances" .Alerts.Resolved }}`,
			"num_firing":   `{{ .Alerts.Firing | len }}`,
			"num_resolved": `{{ .Alerts.Resolved | len }}`,
		},
	}

	// DefaultSlackConfig defines default values for Slack configurations.
	DefaultSlackConfig = SlackConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: false,
		},
		Color:     `{{ template "slack.default.color" }}`,
		Username:  `{{ template "slack.default.username" . }}`,
		Title:     `{{ template "slack.default.title" . }}`,
		TitleLink: `{{ template "slack.default.titlelink" . }}`,
		IconEmoji: `{{ template "slack.default.iconemoji" . }}`,
		IconURL:   `{{ template "slack.default.iconurl" . }}`,
		Pretext:   `{{ template "slack.default.pretext" . }}`,
		Text:      `{{ template "slack.default.text" . }}`,
		Fallback:  `{{ template "slack.default.fallback" . }}`,
	}

	// DefaultHipchatConfig defines default values for Hipchat configurations.
	DefaultHipchatConfig = HipchatConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: false,
		},
		Color:         `{{ if eq .Status "firing" }}red{{ else }}green{{ end }}`,
		From:          `{{ template "hipchat.default.from" . }}`,
		Notify:        false,
		Message:       `{{ template "hipchat.default.message" . }}`,
		MessageFormat: `text`,
	}

	// DefaultOpsGenieConfig defines default values for OpsGenie configurations.
	DefaultOpsGenieConfig = OpsGenieConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: true,
		},
		Message:     `{{ template "opsgenie.default.message" . }}`,
		Description: `{{ template "opsgenie.default.description" . }}`,
		Source:      `{{ template "opsgenie.default.source" . }}`,
		// TODO: Add a details field with all the alerts.
	}

	// DefaultVictorOpsConfig defines default values for VictorOps configurations.
	DefaultVictorOpsConfig = VictorOpsConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: true,
		},
		MessageType:  `CRITICAL`,
		StateMessage: `{{ template "victorops.default.message" . }}`,
		From:         `{{ template "victorops.default.from" . }}`,
	}

	// DefaultPushoverConfig defines default values for Pushover configurations.
	DefaultPushoverConfig = PushoverConfig{
		NotifierConfig: NotifierConfig{
			VSendResolved: true,
		},
		Title:    `{{ template "pushover.default.title" . }}`,
		Message:  `{{ template "pushover.default.message" . }}`,
		URL:      `{{ template "pushover.default.url" . }}`,
		Priority: `{{ if eq .Status "firing" }}2{{ else }}0{{ end }}`, // emergency (firing) or normal
		Retry:    duration(1 * time.Minute),
		Expire:   duration(1 * time.Hour),
	}
)

// NotifierConfig contains base options common across all notifier configurations.
type NotifierConfig struct {
	VSendResolved bool `yaml:"send_resolved" json:"send_resolved"`
}

func (nc *NotifierConfig) SendResolved() bool {
	return nc.VSendResolved
}

// EmailConfig configures notifications via mail.
type EmailConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	// Email address to notify.
	To           string            `yaml:"to" json:"to"`
	From         string            `yaml:"from" json:"from"`
	Smarthost    string            `yaml:"smarthost,omitempty" json:"smarthost,omitempty"`
	AuthUsername string            `yaml:"auth_username" json:"auth_username"`
	AuthPassword Secret            `yaml:"auth_password" json:"auth_password"`
	AuthSecret   Secret            `yaml:"auth_secret" json:"auth_secret"`
	AuthIdentity string            `yaml:"auth_identity" json:"auth_identity"`
	Headers      map[string]string `yaml:"headers" json:"headers"`
	HTML         string            `yaml:"html" json:"html"`
	RequireTLS   *bool             `yaml:"require_tls,omitempty" json:"require_tls,omitempty"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *EmailConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultEmailConfig
	type plain EmailConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.To == "" {
		return fmt.Errorf("missing to address in email config")
	}
	// Header names are case-insensitive, check for collisions.
	normalizedHeaders := map[string]string{}
	for h, v := range c.Headers {
		normalized := strings.Title(h)
		if _, ok := normalizedHeaders[normalized]; ok {
			return fmt.Errorf("duplicate header %q in email config", normalized)
		}
		normalizedHeaders[normalized] = v
	}
	c.Headers = normalizedHeaders

	return checkOverflow(c.XXX, "email config")
}

// PagerdutyConfig configures notifications via PagerDuty.
type PagerdutyConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	ServiceKey  Secret            `yaml:"service_key" json:"service_key"`
	URL         string            `yaml:"url" json:"url"`
	Client      string            `yaml:"client" json:"client"`
	ClientURL   string            `yaml:"client_url" json:"client_url"`
	Description string            `yaml:"description" json:"description"`
	Details     map[string]string `yaml:"details" json:"details"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *PagerdutyConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultPagerdutyConfig
	type plain PagerdutyConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.ServiceKey == "" {
		return fmt.Errorf("missing service key in PagerDuty config")
	}
	return checkOverflow(c.XXX, "pagerduty config")
}

// SlackConfig configures notifications via Slack.
type SlackConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	APIURL Secret `yaml:"api_url" json:"api_url"`

	// Slack channel override, (like #other-channel or @username).
	Channel  string `yaml:"channel" json:"channel"`
	Username string `yaml:"username" json:"username"`
	Color    string `yaml:"color" json:"color"`

	Title     string `yaml:"title" json:"title"`
	TitleLink string `yaml:"title_link" json:"title_link"`
	Pretext   string `yaml:"pretext" json:"pretext"`
	Text      string `yaml:"text" json:"text"`
	Fallback  string `yaml:"fallback" json:"fallback"`
	IconEmoji string `yaml:"icon_emoji" json:"icon_emoji"`
	IconURL   string `yaml:"icon_url" json:"icon_url"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *SlackConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultSlackConfig
	type plain SlackConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return checkOverflow(c.XXX, "slack config")
}

// HipchatConfig configures notifications via Hipchat.
type HipchatConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	APIURL        string `yaml:"api_url" json:"api_url"`
	AuthToken     Secret `yaml:"auth_token" json:"auth_token"`
	RoomID        string `yaml:"room_id" json:"room_id"`
	From          string `yaml:"from" json:"from"`
	Notify        bool   `yaml:"notify" json:"notify"`
	Message       string `yaml:"message" json:"message"`
	MessageFormat string `yaml:"message_format" json:"message_format"`
	Color         string `yaml:"color" json:"color"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *HipchatConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultHipchatConfig
	type plain HipchatConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.RoomID == "" {
		return fmt.Errorf("missing room id in Hipchat config")
	}

	return checkOverflow(c.XXX, "hipchat config")
}

// WebhookConfig configures notifications via a generic webhook.
type WebhookConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	// URL to send POST request to.
	URL string `yaml:"url" json:"url"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *WebhookConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultWebhookConfig
	type plain WebhookConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.URL == "" {
		return fmt.Errorf("missing URL in webhook config")
	}
	return checkOverflow(c.XXX, "webhook config")
}

// OpsGenieConfig configures notifications via OpsGenie.
type OpsGenieConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	APIKey      Secret            `yaml:"api_key" json:"api_key"`
	APIHost     string            `yaml:"api_host" json:"api_host"`
	Message     string            `yaml:"message" json:"message"`
	Description string            `yaml:"description" json:"description"`
	Source      string            `yaml:"source" json:"source"`
	Details     map[string]string `yaml:"details" json:"details"`
	Teams       string            `yaml:"teams" json:"teams"`
	Tags        string            `yaml:"tags" json:"tags"`
	Note        string            `yaml:"note" json:"note"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *OpsGenieConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultOpsGenieConfig
	type plain OpsGenieConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.APIKey == "" {
		return fmt.Errorf("missing API key in OpsGenie config")
	}
	return checkOverflow(c.XXX, "opsgenie config")
}

// VictorOpsConfig configures notifications via VictorOps.
type VictorOpsConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	APIKey       Secret `yaml:"api_key" json:"api_key"`
	APIURL       string `yaml:"api_url" json:"api_url"`
	RoutingKey   string `yaml:"routing_key" json:"routing_key"`
	MessageType  string `yaml:"message_type" json:"message_type"`
	StateMessage string `yaml:"message" json:"message"`
	From         string `yaml:"from" json:"from"`

	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *VictorOpsConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultVictorOpsConfig
	type plain VictorOpsConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.APIKey == "" {
		return fmt.Errorf("missing API key in VictorOps config")
	}
	if c.RoutingKey == "" {
		return fmt.Errorf("missing Routing key in VictorOps config")
	}
	return checkOverflow(c.XXX, "victorops config")
}

type duration time.Duration

func (d *duration) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	if err == nil {
		*d = duration(parsed)
	}
	return err
}

func (d duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

type PushoverConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	UserKey  Secret   `yaml:"user_key" json:"user_key"`
	Token    Secret   `yaml:"token" json:"token"`
	Title    string   `yaml:"title" json:"title"`
	Message  string   `yaml:"message" json:"message"`
	URL      string   `yaml:"url" json:"url"`
	Priority string   `yaml:"priority" json:"priority"`
	Retry    duration `yaml:"retry" json:"retry"`
	Expire   duration `yaml:"expire" json:"expire"`

	// Catches all undefined fields and must be empty after parsing.
	XXX map[string]interface{} `yaml:",inline" json:"-"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *PushoverConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultPushoverConfig
	type plain PushoverConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.UserKey == "" {
		return fmt.Errorf("missing user key in Pushover config")
	}
	if c.Token == "" {
		return fmt.Errorf("missing token in Pushover config")
	}
	return checkOverflow(c.XXX, "pushover config")
}
