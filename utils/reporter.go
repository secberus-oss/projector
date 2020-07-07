package utils

import github "github.com/google/go-github/v32/github"

// Reporter stores all reports and metadata
type Reporter struct {
	Reports []Report `json:"Reports"`
	GH      *GH
}

// Report shows all stats based on a project
type Report struct {
	ProjectBoard   string                 `json:"ProjectBoard"`
	PRsClosed      string                 `json:"PRsClosed"`
	IssuesClosed   string                 `json:"IssuesClosed"`
	LabelBreakdown map[string]interface{} `json:"LabelBreakdown"`
}

// NewReporter creates a new instance of Reporter
func NewReporter() *Reporter {
	r := Reporter{}
	return &r
}

// GenerateReports calls necessary functions to complete a report
func (r *Reporter) GenerateReports(projects []*github.Project) {
	for _, p := range projects {
		report := Report{
			ProjectBoard: *p.Name,
		}
		r.Reports = append(r.Reports, report)
	}
}
