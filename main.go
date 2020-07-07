package main

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	github "github.com/google/go-github/v32/github"
	"github.com/secberus-oss/projector/utils"
	"github.com/spf13/viper"
)

// PRJ stores projector metadata
type PRJ struct {
	status        int
	gh            *utils.GH
	RuleProcessor *utils.RulesProcessor
}

// NewPRJ creates a new instance of PRJ
func NewPRJ() *PRJ {
	viper.SetEnvPrefix("prj") // will be uppercased automatically
	viper.AutomaticEnv()
	prj := PRJ{
		gh:            utils.NewGH(),
		RuleProcessor: utils.NewRulesProcessor(),
	}
	return &prj
}

// CheckHealth used to see if the service is up
func (p *PRJ) CheckHealth() int {
	return 200
}

// RunReports used to run reports
func (p *PRJ) RunReports() []utils.Report {
	r := utils.NewReporter()
	r.GenerateReports(p.gh.Projects)
	return r.Reports
}

// loadConfig to get github things
func (p *PRJ) loadConfig() {
	p.gh.ListRepos()
	projects := p.gh.ListProjects()
	p.gh.DefaultProjectID = *p.gh.GetProjectID(projects, p.gh.DefaultProjectName)
	p.gh.GetDefaultProjectColumns()

	if !p.gh.HookExists(p.gh.ListHooks()) {
		p.gh.CreateHook()
	}
}

func main() {
	prj := NewPRJ()
	prj.loadConfig()
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(prj.CheckHealth(), gin.H{
			"status": "ok",
		})
	})
	r.POST("/webhook", func(c *gin.Context) {
		//log.Println(prEvent.Label)
		payload, _ := github.ValidatePayload(c.Request, prj.gh.Secret)
		event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)

		if err != nil {
			log.Println(err)
		}
		prj.RuleProcessor.ProcessLabelRules(event)
		switch event := event.(type) {
		case *github.PullRequestEvent:
			prj.gh.ProccessPullRequestEvent(event)
		case *github.IssuesEvent:
			prj.gh.ProccessIssuesEvent(event)
		}
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	r.GET("/reports", func(c *gin.Context) {
		reports, err := json.Marshal(prj.RunReports())
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{
				"error": "Failed to Generate Report",
			})
		}
		c.JSON(200, reports)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
