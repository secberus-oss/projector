package main

import (
	"context"
	"log"

	"github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// createHook creates an organization hook to monitor Issues & PRs
func createHook(ctx context.Context, org string, client *github.Client) {
	log.Println("Creating Org Hook...")
	// hookName := viper.GetString("hook_name")
	hookEvents := []string{"pull_request", "issues"}
	hookURL := viper.GetString("hook_url")
	hookConfig := map[string]interface{}{
		"url":          hookURL,
		"content_type": "json",
	}
	hookOptions := &github.Hook{
		Events: hookEvents,
		Config: hookConfig,
	}
	hook, rsp, err := client.Organizations.CreateHook(ctx, org, hookOptions)
	if err != nil {

	}
	if rsp.StatusCode == 404 {
		log.Fatal("Unauthorized..Increase Token Scope")
	} else {
		log.Fatal("Unable to create Org Hook:", err)
	}
	log.Println("hook:", hook.ID)
}

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
	// list all org hooks
	//****************************************************************************
	hooks, rsp, err := client.Organizations.ListHooks(ctx, org, nil)
	if err != nil {
		// check if hooks exist
		if rsp.StatusCode == 404 {
			log.Println("No Org Hooks Found")
			createHook(ctx, org, client)
		} else {
			log.Fatal("Unable to List Hooks in Org:", org, err)
		}
	}
	// list hooks
	for _, h := range hooks {
		log.Println("hook:", *h.ID)
	}
}
