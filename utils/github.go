package utils

import (
	"context"
	"errors"
	"log"
	"os"

	github "github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// GH encapsulates github client & metadata
type GH struct {
	c                 *github.Client
	org               string
	ctx               context.Context
	defaultProjectID  int64
	hookURL           string
	Secret            []byte
	defaultColumnID   *int64
	defaultColumnName string
}

// NewGH creates a new instance of GH
func NewGH() *GH {
	gh := GH{
		c:                 initClient(),
		org:               viper.GetString("org_name"),
		hookURL:           viper.GetString("hook_url"),
		Secret:            []byte(viper.GetString("hook_secret")),
		defaultColumnName: viper.GetString("default_column"),
	}
	return &gh
}

func initClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("github_token")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// init shows all the repos in an org
func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}

// ListRepos shows all the repos in an org
func (g *GH) ListRepos() []*github.Repository {
	ctx := context.Background()
	repos, rsp, err := g.c.Repositories.ListByOrg(ctx, g.org, nil)
	ctxtLog := logrus.WithFields(logrus.Fields{
		"org":      g.org,
		"repos":    repos,
		"response": rsp,
	})
	if err != nil {
		ctxtLog = ctxtLog.WithFields(
			logrus.Fields{
				"error": err,
			})
		ctxtLog.Error("Unable to List Repos in Org")
		return nil
	}
	return repos
}

// ListProjects shows all the projects in an org
func (g *GH) ListProjects() []*github.Project {
	ctx := context.Background()
	projectOptions := &github.ProjectListOptions{State: "open"}
	projects, rsp, err := g.c.Organizations.ListProjects(ctx, g.org, projectOptions)
	ctxtLog := logrus.WithFields(logrus.Fields{
		"org":             g.org,
		"projects":        projects,
		"project_options": projectOptions,
		"response":        rsp,
	})
	if err != nil {
		ctxtLog = ctxtLog.WithFields(
			logrus.Fields{
				"error": err,
			})
		ctxtLog.Error("Unable to List Projects in Org")
	}
	return projects
}

// GetDefaultProjectID gets the id of project to be added on all PRs/Issues by default
func (g *GH) GetDefaultProjectID(projects []*github.Project) (*int64, error) {
	for _, p := range projects {
		log.Println("project:", *p.Name)
		if *p.Name == viper.GetString("default_project") {
			return p.ID, nil
		}
	}
	return nil, errors.New("Unable to get Project ID")
}

// ListHooks gets all of the hooks in an org
func (g *GH) ListHooks() []*github.Hook {
	ctx := context.Background()
	hooks, rsp, err := g.c.Organizations.ListHooks(ctx, g.org, nil)
	ctxtLog := logrus.WithFields(logrus.Fields{
		"org":      g.org,
		"hooks":    hooks,
		"response": rsp,
	})
	if err != nil {
		ctxtLog = ctxtLog.WithFields(
			logrus.Fields{
				"error": err,
			})
		ctxtLog.Error("Unable to List Hooks in Org")
		return nil
	}
	return hooks
}

// HookExists checks if a hook with provided URL already exists
func (g *GH) HookExists(hooks []*github.Hook) bool {
	for _, h := range hooks {
		if g.hookURL == h.Config["url"].(string) {
			return true
		}
	}
	return false
}

// CreateHook creates an organization hook to monitor Issues & PRs
func (g *GH) CreateHook() *github.Hook {
	ctx := context.Background()
	hookEvents := []string{"pull_request", "issues"}
	hookConfig := map[string]interface{}{
		"url":          g.hookURL,
		"content_type": "json",
	}
	hookOptions := &github.Hook{
		Events: hookEvents,
		Config: hookConfig,
	}
	hook, rsp, err := g.c.Organizations.CreateHook(ctx, g.org, hookOptions)
	ctxtLog := logrus.WithFields(logrus.Fields{
		"org":      g.org,
		"hook":     hook,
		"response": rsp,
	})
	if rsp.StatusCode == 404 {
		ctxtLog.Error("Unauthorized to Create Hook in Org")
	}
	if err != nil {
		ctxtLog = ctxtLog.WithFields(
			logrus.Fields{
				"error": err,
			})
		ctxtLog.Error("Unable to Create Hook in Org")
	}
	return hook
}

// ListProjectColumns gets all the columns of a project
func (g *GH) ListProjectColumns() []*github.ProjectColumn {
	ctx := context.Background()
	columns, _, err := g.c.Projects.ListProjectColumns(ctx, g.defaultProjectID, &github.ListOptions{})
	if err != nil {
		log.Fatal("Unable to List columns in project ", err)
		return nil
	}
	return columns
}

// GetCardColumnIDByName returns the ID of a column given a name
func (g *GH) GetCardColumnIDByName(columns []*github.ProjectColumn, columnName string) *int64 {
	for _, c := range columns {
		log.Println("columns:", *c.Name)
		if *c.Name == columnName {
			return c.ID
		}
	}
	log.Fatal("Unable to find column: ", columnName, "in project")
	return nil
}

// CreatetProjectCard adds the Project to an Issue or PR
func (g *GH) CreatetProjectCard(contentType string, id int64, columnID int64) {
	ctx := context.Background()
	projectCardOptions := &github.ProjectCardOptions{
		ContentID:   id,
		ContentType: contentType,
	}
	card, _, err := g.c.Projects.CreateProjectCard(ctx, columnID, projectCardOptions)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(card)
}

// ProccessPullRequestEvent takes a PR event and performs actions on it
func (g *GH) ProccessPullRequestEvent(e *github.PullRequestEvent) {
	log.Println("Received PR Event!", *e.Action)
	if *e.Action == "opened" {
		g.defaultColumnID = g.GetCardColumnIDByName(g.ListProjectColumns(), g.defaultColumnName)
		g.CreatetProjectCard("pull_request", *e.PullRequest.ID, *g.defaultColumnID)
	}
}

// ProccessIssuesEvent takes a PR event and performs actions on it
func (g *GH) ProccessIssuesEvent(e *github.IssuesEvent) {
	log.Print("Received Issues Event! ")
}
