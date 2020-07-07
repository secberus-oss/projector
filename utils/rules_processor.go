package utils

import (
	"log"
	"reflect"
	"strings"

	github "github.com/google/go-github/v32/github"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// RulesProcessor encapsulates rules and processes them
type RulesProcessor struct {
	gh         *GH
	rc         *viper.Viper
	LabelRules []LabelRule
}

// LabelRule defines rules based on labels
type LabelRule struct {
	Name        string
	Description string
	Column      string
	Label       string
	State       string
	Content     string
	Project     string
}

// NewRulesProcessor creates new metadata object of Rules
func NewRulesProcessor() *RulesProcessor {
	r := RulesProcessor{
		rc: viper.New(),
		gh: NewGH(),
	}
	r.LoadRulesConfig()
	return &r
}

// LoadRulesConfig so we can process all the rules
func (r *RulesProcessor) LoadRulesConfig() {
	log.Println("Loading rules config...")
	r.rc.SetConfigName(".prj") // name of config file (without extension)
	r.rc.SetConfigType("yaml")
	r.rc.AddConfigPath(".")
	r.rc.AddConfigPath("$HOME/.prj")
	err := r.rc.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		log.Println("Error reading config file:", err)
	} else {
		log.Print("Loaded Rules Config")
	}
}

// MatchesPRRuleConditions make sure the rule has all its conditions met
func (r *RulesProcessor) MatchesPRRuleConditions(rule LabelRule, e *github.PullRequestEvent) bool {
	if typeCheck := strings.Contains(reflect.TypeOf(e).String(), rule.Content); typeCheck != true {
		log.Println("Content Type Condition Check Failed")
		return false
	}
	if *e.PullRequest.State != rule.State {
		log.Println("State Condition Check Failed")
		return false
	}
	if r.HasLabels(rule, e.PullRequest.Labels) != true {
		log.Println("Label Condition Check Failed")
		return false
	}
	log.Println("All Condition Checks Passed!!")
	return true
}

// MatchesIssueConditions make sure the rule has all its conditions met
func (r *RulesProcessor) MatchesIssueConditions(rule LabelRule, e *github.IssuesEvent) bool {
	if typeCheck := strings.Contains(reflect.TypeOf(e).String(), rule.Content); typeCheck != true {
		return false
	}
	if *e.Issue.State != rule.State {
		return false
	}
	if r.HasLabels(rule, e.Issue.Labels) != true {
		return false
	}
	return true
}

// HasLabels check
func (r *RulesProcessor) HasLabels(rule LabelRule, labels []*github.Label) bool {
	for _, l := range labels {
		if rule.Label == *l.Name {
			return true
		}
	}
	return false
}

// ProcessLabelRules so we can automate the things
func (r *RulesProcessor) ProcessLabelRules(e interface{}) {
	labelRules := r.rc.Get("LabelRules")
	lRMap := labelRules.([]interface{})
	mapstructure.Decode(lRMap, &r.LabelRules)
	log.Print("Found Rules", r.LabelRules)
	switch e := e.(type) {
	case *github.PullRequestEvent:
		log.Print("received a PR to process label rules")
		for _, rule := range r.LabelRules {
			if r.MatchesPRRuleConditions(rule, e) {
				log.Print("Found a Rule that Matches an Event Condition")
				projID := r.gh.GetProjectID(r.gh.ListProjects(), rule.Project)
				columns := r.gh.ListProjectColumns(*projID)
				if colID, ok := r.gh.GetCardColumnIDByName(columns, rule.Column); ok {
					//the value exists
					r.gh.CreatetProjectCard(rule.Content, *e.PullRequest.ID, colID)
				} else {
					log.Print("Unable to get Column ID")
				}
			}
		}

	case *github.IssuesEvent:
		log.Print("received an Issue to process label rules", e)
	}
}