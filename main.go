package main

import (
	"context"
	"log"

	"github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func main() {
	viper.SetEnvPrefix("prj") // will be uppercased automatically
	viper.AutomaticEnv()
	org := viper.GetString("org_name")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("github_token")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	//****************************************************************************
	// list all repositories
	//****************************************************************************
	repos, _, err := client.Repositories.ListByOrg(ctx, org, nil)
	if err != nil {
		log.Fatal("Unable to List Repos in Org:", org, err)
	}
	for _, r := range repos {
		log.Println("repo:", *r.Name)
	}
	//****************************************************************************
	// list all projects
	//****************************************************************************
	projectOptions := &github.ProjectListOptions{State: "open"}
	projects, _, err := client.Organizations.ListProjects(ctx, org, projectOptions)
	if err != nil {
		log.Fatal("Unable to List Projects in Org:", org, err)
	}
	for _, p := range projects {
		log.Println("repo:", *p.Name)
	}
	//****************************************************************************
}
