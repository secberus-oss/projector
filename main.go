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
	viper.BindEnv("github_token")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("github_token")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	repos, _, err := client.Repositories.List(ctx, "", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range repos {
		log.Printf(*r.Name)
	}
}
