package service

import (
	"context"
	"github.com/google/go-github/v35/github"
)

type Service struct {
	ctx    context.Context
	client *github.Client
	org    string
	repo   string
	pr     int
}

func New(ctx context.Context, client *github.Client, org string, repo string, pr int) *Service {
	return &Service{
		ctx:    ctx,
		client: client,
		org:    org,
		repo:   repo,
		pr:     pr,
	}
}

func (s *Service) Commits() ([]*github.RepositoryCommit, error) {
	commits, _, err := s.client.PullRequests.ListCommits(s.ctx, s.org, s.repo, s.pr, nil)
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (s *Service) CheckRuns(ref ...string) (*github.ListCheckRunsResults, error) {
	var r string

	if len(ref) > 0 {
		r = ref[0]
	} else {
		pr, _, err := s.client.PullRequests.Get(s.ctx, s.org, s.repo, s.pr)
		if err != nil {
			return nil, err
		}

		r = *pr.Head.SHA
	}

	checkRuns, _, err := s.client.Checks.ListCheckRunsForRef(s.ctx, s.org, s.repo, r, nil)
	if err != nil {
		return nil, err
	}

	return checkRuns, nil
}
