package service

import (
	"context"
	"fmt"
	"github.com/aemengo/gswt/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/google/go-github/v35/github"
)

type Service struct {
	ctx          context.Context
	client       *github.Client
	pr           *github.PullRequest
	workflowRuns []*github.WorkflowRun
	logger       *log.Logger
	homeDir      string
	org          string
	repo         string
	prNum        int
	fetchChan    chan bool
}

func New(ctx context.Context, client *github.Client, logger *log.Logger, homeDir string, org string, repo string, prNum int) (*Service, error) {
	pr, _, err := client.PullRequests.Get(ctx, org, repo, prNum)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		ctx:       ctx,
		client:    client,
		logger:    logger,
		homeDir:   homeDir,
		org:       org,
		repo:      repo,
		prNum:     prNum,
		pr:        pr,
		fetchChan: make(chan bool, 1),
	}
	go svc.pullAllWorkflowRuns()
	return svc, nil
}

func (s *Service) Logs(checkRun *github.CheckRun) (string, error) {
	path := s.logPath(checkRun)
	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}

	s.waitForWorkflows()

	workFlowId, ok := s.pullWorkflowID(checkRun)
	if !ok {
		return "", fmt.Errorf("unable to find workflow for '%s'", checkRun.GetName())
	}

	jobs, _, err := s.client.Actions.ListWorkflowJobs(s.ctx, s.org, s.repo, workFlowId, &github.ListWorkflowJobsOptions{
		Filter: "latest",
	})
	if err != nil {
		return "", err
	}

	jobId, ok := s.pullWorkflowJobID(jobs, checkRun)
	if !ok {
		return "", fmt.Errorf("unable to find workflow job for '%s'", checkRun.GetName())
	}

	link, _, err := s.client.Actions.GetWorkflowJobLogs(s.ctx, s.org, s.repo, jobId, true)
	if err != nil {
		return "", err
	}

	err = download(link.String(), path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func download(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func (s *Service) waitForWorkflows() {
	if len(s.workflowRuns) > 0 {
		return
	}

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-timeout:
			return
		case <-s.fetchChan:
			return
		}
	}
}

func (s *Service) pullWorkflowJobID(jobs *github.Jobs, checkRun *github.CheckRun) (int64, bool) {
	for _, job := range jobs.Jobs {
		checkRunID := strconv.FormatInt(checkRun.GetID(), 10)

		if path.Base(job.GetCheckRunURL()) == checkRunID {
			return job.GetID(), true
		}
	}

	return 0, false
}

func (s *Service) pullWorkflowID(checkRun *github.CheckRun) (int64, bool) {
	for _, run := range s.workflowRuns {
		checkSuiteID := strconv.FormatInt(checkRun.GetCheckSuite().GetID(), 10)
		if path.Base(run.GetCheckSuiteURL()) == checkSuiteID {
			return run.GetID(), true
		}
	}

	return 0, false
}

func (s *Service) pullAllWorkflowRuns() {
	var (
		result []*github.WorkflowRun
		page   = 1
	)

	for {
		runs, _, err := s.client.Actions.ListRepositoryWorkflowRuns(s.ctx, s.org, s.repo, &github.ListWorkflowRunsOptions{
			Actor:       s.pr.GetUser().GetLogin(),
			Branch:      s.pr.GetHead().GetRef(),
			Event:       "pull_request",
			Status:      "completed",
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		})
		if err != nil {
			s.logger.Println(err)
			s.fetchChan <- true
			return
		}

		if runs.GetTotalCount() == 0 {
			s.workflowRuns = result
			s.fetchChan <- true
			return
		}

		result = append(result, runs.WorkflowRuns...)
		page = page + 1
	}
}

func (s *Service) logPath(checkRun *github.CheckRun) string {
	filename := fmt.Sprintf("%d.log", checkRun.GetID())
	return filepath.Join(utils.LogsDir(s.homeDir), filename)
}

func (s *Service) Commits() ([]*github.RepositoryCommit, error) {
	commits, _, err := s.client.PullRequests.ListCommits(s.ctx, s.org, s.repo, s.prNum, nil)
	if err != nil {
		return nil, err
	}

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Commit.Committer.Date.After(*commits[j].Commit.Committer.Date)
	})

	return commits, nil
}

func (s *Service) CheckRuns(ref ...string) (*github.ListCheckRunsResults, error) {
	var r = s.pr.GetHead().GetSHA()
	if len(ref) > 0 {
		r = ref[0]
	}

	checkRuns, _, err := s.client.Checks.ListCheckRunsForRef(s.ctx, s.org, s.repo, r, nil)
	if err != nil {
		return nil, err
	}

	return checkRuns, nil
}
