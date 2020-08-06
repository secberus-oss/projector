package utils

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	github "github.com/google/go-github/v32/github"
)

// Reporter stores all reports and metadata
type Reporter struct {
	Reports []Report `json:"Reports"`
	GH      *GH
}

// Report shows all stats based on a project
type Report struct {
	ProjectBoard string `json:"ProjectBoard"`
	IssuesClosed int    `json:"IssuesClosed"`
	//PRsClosed    int           `json:"PRsClosed"`
	LabelCounts  []*LabelCount `json:"LabelCounts"`
	ProjectCards []*Card       `json:"ProjectCards"`
}

// Card extends github cards to store type
type Card struct {
	github.ProjectCard
	ContentType    string          `json:"ContentType"`
	Repo           string          `json:"Repo"`
	Number         int             `json:"Number"`
	Name           string          `json:"Name"`
	PullRequestURL *string         `json:"PullRequestURL,omitempty"`
	Labels         []*github.Label `json:"Labels,omitempty"`
}

// LabelCount keeps  track of all issues with labels
type LabelCount struct {
	Name  *string `json:"Name,omitempty"`
	Count int     `json:"Count,omitempty"`
}

// NewReporter creates a new instance of Reporter
func NewReporter() *Reporter {
	r := Reporter{
		GH: NewGH(),
	}
	return &r
}

// GenerateReports calls necessary functions to complete a report
func (r *Reporter) GenerateReports(projects []*github.Project) {
	log.Println("Reports Generating...")
	var wg sync.WaitGroup

	for _, p := range projects {
		wg.Add(1)
		go r.GenerateReport(p, &wg)
	}
	wg.Wait()
}

// GenerateReport calls necessary functions to complete a report
func (r *Reporter) GenerateReport(project *github.Project, wg *sync.WaitGroup) {
	cardsDone := r.GetProjectCardsFromColumn(*project.ID, "Done")
	cardsWithMetadata := r.GetContentTypes(cardsDone)
	report := Report{
		ProjectBoard: *project.Name,
		IssuesClosed: len(cardsDone),
		LabelCounts:  r.GetLabelCount(cardsWithMetadata),
		ProjectCards: cardsWithMetadata,
	}
	log.Println("Processing report for", *project.Name)
	r.Reports = append(r.Reports, report)
	defer wg.Done()
}

// GetProjectCardsFromColumn Gets all the cards for a Project
func (r *Reporter) GetProjectCardsFromColumn(projID int64, column string) []*github.ProjectCard {
	cols := r.GH.ListProjectColumns(projID)
	var cards []*github.ProjectCard
	if colID, ok := r.GH.GetCardColumnIDByName(cols, column); ok {
		//the value exists
		cards = r.GH.ListProjectCards(colID)
	} else {
		log.Print("Unable to get Column ID")
		return nil
	}
	return cards
}

// GetContentTypes figures out if PR or Issue
func (r *Reporter) GetContentTypes(cards []*github.ProjectCard) []*Card {
	cardsWithType := []*Card{}
	for _, c := range cards {
		c.Creator = &github.User{Name: c.Creator.Name}
		var card = Card{ProjectCard: *c}
		if c.ContentURL != nil {
			var card = Card{ProjectCard: *c}
			repo, contentType, number, err := r.StripContentURL(*c.ContentURL)
			if err != nil {
				log.Printf("Error Parsing URL")
			}
			card.Number = *number
			card.Repo = *repo
			if issue, ok := r.isIssue(&card); ok {
				card.ContentType = *contentType
				card.Name = *issue.Title
				card.Labels = issue.Labels
				if issue.PullRequestLinks != nil {
					card.PullRequestURL = issue.PullRequestLinks.URL
				}
			}
			cardsWithType = append(cardsWithType, &card)
		} else {
			cardsWithType = append(cardsWithType, &card)
		}
	}
	return cardsWithType
}


func (r *Reporter) isIssue(card *Card) (*github.Issue, bool) {
	i, _ := r.GH.GetIssue(card.Repo, card.Number)
	if i != nil && *i.State == "closed" {
		return i, true
	}
	return nil, false
}

// Get All Project Labels, then for each label found increase count

// GetLabelCount does all the summation of labels
func (r *Reporter) GetLabelCount(cards []*Card) []*LabelCount {
	labelCounts := []*LabelCount{}
	for _, c := range cards {
		for _, l := range c.Labels {
			lc := &LabelCount{
				Name:  l.Name,
				Count: 1,
			}
			labelCounts = r.AppendIfMissing(labelCounts, lc)
		}
	}
	return labelCounts
}

// AppendIfMissing counts labels
func (r *Reporter) AppendIfMissing(slice []*LabelCount, i *LabelCount) []*LabelCount {
	for _, ele := range slice {
		if *ele.Name == *i.Name {
			ele.Count++
			return slice
		}
	}
	return append(slice, i)
}

// StripContentURL strips an api url for data
func (r *Reporter) StripContentURL(s string) (*string, *string, *int, error) {
	stripedContent := strings.Replace(s, "https://api.github.com/repos/"+r.GH.org+"/", "", -1)
	sc := strings.Split(stripedContent, "/")
	repo := sc[0]
	contentType := sc[1]
	number, err := strconv.Atoi(sc[2])
	if err != nil {
		return nil, nil, nil, errors.New("Issue/PR Number Conversion Error")
	}
	return &repo, &contentType, &number, nil
}
