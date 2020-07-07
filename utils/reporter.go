package utils

import (
	"log"
	"strconv"

	github "github.com/google/go-github/v32/github"
)

// Reporter stores all reports and metadata
type Reporter struct {
	Reports []Report `json:"Reports"`
	GH      *GH
}

// Report shows all stats based on a project
type Report struct {
	ProjectBoard   string                 `json:"ProjectBoard"`
	TotalClosed    string                 `json:"TotalClosed"`
	LabelBreakdown map[string]interface{} `json:"LabelBreakdown"`
	ProjectCards   []*github.ProjectCard  `json:"ProjectCards"`
	Issues         []*github.Issue        `json:"Issues"`
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
	for _, p := range projects {
		cardsDone := r.GetProjectCardsFromColumn(*p.ID, "Done")
		report := Report{
			ProjectBoard: *p.Name,
			ProjectCards: cardsDone,
			TotalClosed:  strconv.Itoa(len(cardsDone)),
		}
		log.Println("Processing report for", *p.Name)
		r.Reports = append(r.Reports, report)
	}
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
