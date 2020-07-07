package utils

import (
	"context"
	"log"

	github "github.com/google/go-github/v32/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// GH encapsulates github client & metadata
type GH struct {
	c                  *github.Client
	org                string
	ctx                context.Context
	DefaultProjectName string
	DefaultProjectID   int64
	hookURL            string
	Secret             []byte
	defaultColumnID    int64
	defaultColumnName  string
	repos              []*github.Repository
	projects           []*github.Project
	defaultColumns     []*github.ProjectColumn
}

// NewGH creates a new instance of GH
func NewGH() *GH {
	gh := GH{
		c:                  initClient(),
		org:                viper.GetString("org_name"),
		hookURL:            viper.GetString("hook_url"),
		Secret:             []byte(viper.GetString("hook_secret")),
		defaultColumnName:  viper.GetString("default_column"),
		DefaultProjectName: viper.GetString("default_project"),
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

// ListRepos shows all the repos in an org
func (g *GH) ListRepos() {
	ctx := context.Background()
	repos, rsp, err := g.c.Repositories.ListByOrg(ctx, g.org, nil)
	if err != nil {
		log.Println("Unable to List Repos in Org...", rsp, err)
	}
	log.Println("Listing Repos..")
	g.repos = repos
}

// ListProjects shows all the projects in an org
func (g *GH) ListProjects() []*github.Project {
	ctx := context.Background()
	projectOptions := &github.ProjectListOptions{State: "open"}
	projects, rsp, err := g.c.Organizations.ListProjects(ctx, g.org, projectOptions)
	if err != nil {
		log.Println("Unable to List Projects in Org", rsp, g.org, err)
	}
	return projects
}

// GetProjectID gets the id of project to be added on all PRs/Issues by default
func (g *GH) GetProjectID(projects []*github.Project, name string) *int64 {
	for _, p := range projects {
		log.Println("Found Project ID:", *p.ID, "For Project:", *p.Name)
		if *p.Name == name {
			log.Println("Found Project ID:", *p.ID, "For Project:", *p.Name)
			return p.ID
		}
	}
	log.Println("Couldn't Find Project ID for:", name)
	return nil
}

// GetDefaultProjectColumns sets data for default project columns
func (g *GH) GetDefaultProjectColumns() {
	g.defaultColumns = g.ListProjectColumns(g.DefaultProjectID)
}

// GetDefaultColumnID returns the id of the default column for new PRs/Issues
func (g *GH) GetDefaultColumnID() {
	if v, ok := g.GetCardColumnIDByName(g.defaultColumns, g.defaultColumnName); ok {
		//the value exists
		g.defaultColumnID = v
	}
}

// ListHooks gets all of the hooks in an org
func (g *GH) ListHooks() []*github.Hook {
	ctx := context.Background()
	hooks, rsp, err := g.c.Organizations.ListHooks(ctx, g.org, nil)
	if err != nil {
		log.Println("Unable to List Hooks in Org", g.org, rsp, err)
		return nil
	}
	log.Println("Listing Hooks...", hooks)
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
	if rsp.StatusCode == 404 {
		log.Println("Unauthorized to Create Hook in Org")
	}
	if err != nil {
		log.Println("Unable to Create Hook in Org", rsp, err)
	}
	log.Println("Created Hook: ", hook)
	return hook
}

// ListProjectColumns gets all the columns of a project
func (g *GH) ListProjectColumns(prjID int64) []*github.ProjectColumn {
	ctx := context.Background()
	columns, _, err := g.c.Projects.ListProjectColumns(ctx, prjID, &github.ListOptions{})
	if err != nil {
		log.Fatal("Unable to List columns in project ", err)
		return nil
	}
	return columns
}

// GetCardColumnIDByName returns the ID of a column given a name
func (g *GH) GetCardColumnIDByName(columns []*github.ProjectColumn, columnName string) (int64, bool) {
	for _, c := range columns {
		log.Println("columns:", *c.Name)
		if *c.Name == columnName {
			log.Println("Column ID Found:", *c.ID)
			return *c.ID, true
		}
	}
	return 0, false
}

// CreatetProjectCard adds the Project to an Issue or PR
func (g *GH) CreatetProjectCard(contentType string, id int64, columnID int64) {
	log.Println("Creating project card...")
	ctx := context.Background()
	projectCardOptions := &github.ProjectCardOptions{
		ContentID:   id,
		ContentType: contentType,
	}
	card, rsp, err := g.c.Projects.CreateProjectCard(ctx, columnID, projectCardOptions)
	if err != nil {
		log.Println("projectCardOptions:", projectCardOptions)
		log.Println("Problem Creating Project Card", rsp, err)
	} else {
		log.Println("Created Project Card", card)
	}
}

// ProccessPullRequestEvent takes a PR event and performs actions on it
func (g *GH) ProccessPullRequestEvent(e *github.PullRequestEvent) {
	log.Println("Received PR Event! Action: ", *e.Action)
	if *e.Action == "opened" && *e.PullRequest.State == "open" {
		log.Println("Processing Opened PR Event...")
		log.Println("PR ID:", *e.PullRequest.ID)
		log.Println("Project Column Name:", g.defaultColumnName, "Column ID: ", g.defaultColumnID, "Proj ID:", g.DefaultProjectID)
		g.CreatetProjectCard("PullRequest", *e.PullRequest.ID, g.defaultColumnID)
	}
}

// ProccessIssuesEvent takes a PR event and performs actions on it
func (g *GH) ProccessIssuesEvent(e *github.IssuesEvent) {
	log.Print("Received Issues Event! ")
	if *e.Action == "opened" {
		g.CreatetProjectCard("Issue", *e.Issue.ID, g.defaultColumnID)
	}
}
