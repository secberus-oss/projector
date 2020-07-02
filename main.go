package main

import (
	"log"

	"github.com/gin-gonic/gin"
	github "github.com/google/go-github/v32/github"
	"github.com/secberus-oss/projector/utils"
	"github.com/spf13/viper"
)

// PRJ stores projector metadata
type PRJ struct {
	status int
}

// NewPRJ creates a new instance of PRJ
func NewPRJ() *PRJ {
	prj := PRJ{}
	return &prj
}

// CheckHealth used to see if the service is up
func (prj *PRJ) CheckHealth() int {
	return 200
}

func main() {

	prj := NewPRJ()
	viper.SetEnvPrefix("prj") // will be uppercased automatically
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	gh := utils.NewGH()
	gh.ListRepos()
	gh.ListProjects()
	if !gh.HookExists(gh.ListHooks()) {
		gh.CreateHook()
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(prj.CheckHealth(), gin.H{
			"status": "ok",
		})
	})
	r.POST("/webhook", func(c *gin.Context) {
		//prEvent := proccessPullRequestEvent(c)
		//log.Println(prEvent.Label)
		payload, err := github.ValidatePayload(c.Request, gh.Secret)
		event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
		if err != nil {
			log.Println(err)
		}
		switch event := event.(type) {
		case *github.PullRequestEvent:
			gh.ProccessPullRequestEvent(event)
		case *github.IssuesEvent:
			gh.ProccessIssuesEvent(event)
		}
		c.JSON(200, gin.H{
			"status": "ok",
			"action": "test",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
