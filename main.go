package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var hookURL string

// createHook creates an organization hook to monitor Issues & PRs
func createHook(ctx context.Context, org string, client *github.Client) {
	log.Println("Creating Org Hook...")
	// hookName := viper.GetString("hook_name")
	hookEvents := []string{"pull_request", "issues"}
	hookConfig := map[string]interface{}{
		"url":          hookURL,
		"content_type": "json",
	}
	hookOptions := &github.Hook{
		Events: hookEvents,
		Config: hookConfig,
	}
	hook, rsp, err := client.Organizations.CreateHook(ctx, org, hookOptions)
	if rsp.StatusCode == 404 {
		log.Println("Unauthorized...Increase Token Scope")
		return
	}
	if err != nil {
		log.Fatal("Unable to create Org Hook:", err)
	}
	log.Println("hook:", hook)
}

func hookExists(h github.Hook) bool {
	log.Println("checking if", hookURL, "already created in", h)
	if hookURL == h.Config["url"].(string) {
		return true
	}
	return false
}

func processPullRequestEvent(c *gin.Context) github.PullRequestEvent {
	var json github.PullRequestEvent
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return github.PullRequestEvent{}
	}
	return json
}

func main() {
	viper.SetEnvPrefix("prj") // will be uppercased automatically
	viper.AutomaticEnv()
	hookURL = viper.GetString("hook_url")
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
		log.Println("project:", *p.Name)
	}
	//****************************************************************************
	// list all org hooks
	//****************************************************************************
	hooks, _, err := client.Organizations.ListHooks(ctx, org, nil)
	if err != nil {
		log.Fatal("Unable to List Hooks in Org:", org, err)
	} else if len(hooks) > 0 {
		for _, h := range hooks {
			log.Println("hook:", *h)
			if hookExists(*h) {
				log.Println("Hook already created, skipping...")
			}
		}
	} else {
		createHook(ctx, org, client)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	r.POST("/webhook", func(c *gin.Context) {
		prEvent := processPullRequestEvent(c)

		// payload, err := github.ValidatePayload(r, s.webhookSecretKey)
		// event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// switch event := event.(type) {
		// case *github.PullRequestEvent:
		// 	proccessPullRequestEvent(event)
		// }
		c.JSON(200, gin.H{
			"status": "ok",
			"action": prEvent.Action,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
