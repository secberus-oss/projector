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
	if !strings.Contains(reflect.TypeOf(e).String(), rule.Content) {
		log.Println("Content Type Condition Check Failed", rule)
		return false
	}
	if *e.PullRequest.State != rule.State {
		log.Println("State Condition Check Failed")
		return false
	}
	if e.Label == nil || rule.Label != *e.Label.Name {
		log.Println("Label Condition Check Failed", rule)
		return false
	}
	log.Println("All Condition Checks Passed!!")
	return true
}

// MatchesIssueConditions make sure the rule has all its conditions met
func (r *RulesProcessor) MatchesIssueConditions(rule LabelRule, e *github.IssuesEvent) bool {
	if !strings.Contains(reflect.TypeOf(e).String(), rule.Content) {
		log.Println("Content Type Condition Check Failed", rule)
		return false
	}
	if *e.Issue.State != rule.State {
		log.Println("State Condition Check Failed", rule)
		return false
	}
	if e.Label == nil || rule.Label != *e.Label.Name {
		log.Println("Label Condition Check Failed", rule)
		return false
	}
	log.Println("All Condition Checks Passed!!", rule)
	return true
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
				projID := r.gh.GetProjectID(rule.Project)
				columns := r.gh.ListProjectColumns(*projID)
				if colID, ok := r.gh.GetCardColumnIDByName(columns, rule.Column); ok {
					//the value exists
					r.gh.CreateProjectCard(rule.Content, *e.PullRequest.ID, colID)
				} else {
					log.Print("Unable to get Column ID")
				}
			}
		}
	case *github.IssuesEvent:
		log.Print("received an Issue to process label rules")
		if *e.Action != "labeled" && *e.Action != "unlabeled" {
			log.Println("Ignoring issue action", *e.Action)
			return
		}
		for _, rule := range r.LabelRules {
			if r.MatchesIssueConditions(rule, e) {
				log.Println("Found a Rule that Matches an Event Condition")
				projID := r.gh.GetProjectID(rule.Project)
				columns := r.gh.ListProjectColumns(*projID)

				if colID, ok := r.gh.GetCardColumnIDByName(columns, rule.Column); ok {
					//the value exists
					if *e.Action == "labeled" {
						r.gh.CreateProjectCard(rule.Content, *e.Issue.ID, colID)
					} else {
						r.gh.DeleteProjectIssueCard(rule.Content, *e.Issue, *e.Repo.Name, rule.Project)
					}
				} else {
					log.Print("Unable to get Column ID")
				}
			}
		}
	}
}
