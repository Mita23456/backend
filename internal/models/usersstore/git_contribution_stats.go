// Code generated by github.com/firstcontributions/matro. DO NOT EDIT.

package usersstore

type GitContributionStats struct {
	Issues       int64 `bson:"issues,omitempty"`
	PullRequests int64 `bson:"pull_requests,omitempty"`
}

func NewGitContributionStats() *GitContributionStats {
	return &GitContributionStats{}
}

type GitContributionStatsFilters struct {
	Ids []string
}
