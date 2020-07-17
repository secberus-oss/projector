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
	log.Print("Running Reports...")
	r := utils.NewReporter()
	r.GenerateReports(p.gh.Projects)
	return r.Reports
}

// loadConfig to get github things
func (p *PRJ) loadConfig() {
	p.gh.DefaultProjectID = *p.gh.GetProjectID(p.gh.DefaultProjectName)
	p.gh.GetDefaultProjectColumns()
	p.gh.GetDefaultColumnID()
	if !p.gh.HookExists(p.gh.ListHooks()) {
		p.gh.CreateHook()
	}
}

func main() {
	prj := NewPRJ()
	prj.loadConfig()
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(prj.CheckHealth(), gin.H{
			"status": "ok",
		})
	})
	r.POST("/webhook", func(c *gin.Context) {
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
		//log.Println(string(reports))
		c.JSON(200, prj.RunReports())
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
