package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v32/github"
	"github.com/secberus-oss/projector/utils"
	"github.com/spf13/viper"
)

var projectID int64
var columnID int64

func createPullRequestProjectCard(ctx context.Context, client *github.Client, id int64, columnID int64) {
	projectCardOptions := &github.ProjectCardOptions{
		ContentID:   id,
		ContentType: "PullRequest",
	}
	card, _, err := client.Projects.CreateProjectCard(ctx, columnID, projectCardOptions)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(card)
}

func getCardColumnID(ctx context.Context, client *github.Client, columnName string) int64 {
	columns, _, err := client.Projects.ListProjectColumns(ctx, projectID, &github.ListOptions{})
	if err != nil {
		log.Fatal("Unable to List columns in project ", err)
	}
	for _, c := range columns {
		log.Println("columns:", *c.Name)
		if *c.Name == columnName {
			return *c.ID
		}
	}
	log.Fatal("Unable to find column: ", columnName, "in project")
	return 0
}

func proccessPullRequestEvent(ctx context.Context, client *github.Client, e *github.PullRequestEvent) {
	log.Println("Received PR Event!", *e.Action)
	if *e.Action == "opened" {
		columnID := getCardColumnID(ctx, client, viper.GetString("default_column"))
		createPullRequestProjectCard(ctx, client, *e.PullRequest.ID, columnID)
	}
}

func proccessIssuesEvent(e *github.IssuesEvent) {
	log.Print("Received Issues Event! ")
}

func main() {
	viper.SetEnvPrefix("prj") // will be uppercased automatically
	viper.AutomaticEnv()

	gh := utils.NewGH()
	gh.ListRepos()
	gh.ListProjects()
	gh.ListHooks()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	r.POST("/webhook", func(c *gin.Context) {
		//prEvent := proccessPullRequestEvent(c)
		//log.Println(prEvent.Label)
		var empty []byte
		payload, err := github.ValidatePayload(c.Request, empty)
		event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
		if err != nil {
			log.Println(err)
		}
		switch event := event.(type) {
		case *github.PullRequestEvent:
			proccessPullRequestEvent(ctx, client, event)
		case *github.IssuesEvent:
			proccessIssuesEvent(event)
		}
		c.JSON(200, gin.H{
			"status": "ok",
			"action": "test",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
