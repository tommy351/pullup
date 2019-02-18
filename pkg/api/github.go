package api

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v24/github"
	"github.com/labstack/echo/v4"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/kubernetes"
)

func (s *Server) GitHubWebhook(ctx echo.Context) error {
	req := ctx.Request()
	payload, err := s.validatePayload(req)

	if err != nil {
		return err
	}

	event, err := github.ParseWebHook(github.WebHookType(req), payload)

	if err != nil {
		return err
	}

	switch event := event.(type) {
	case *github.PullRequestEvent:
		repo := s.findRepoConfig("github.com/" + *event.Repo.FullName)

		if repo == nil {
			return ctx.String(http.StatusNotFound, "Repository is not set in the config")
		}

		if err := s.handlePullRequestEvent(event, repo); err != nil {
			return err
		}
	}

	return ctx.String(http.StatusOK, "ok")
}

func (s *Server) validatePayload(req *http.Request) ([]byte, error) {
	secret := s.Config.GitHub.Secret

	if secret == "" {
		return ioutil.ReadAll(req.Body)
	}

	return github.ValidatePayload(req, []byte(secret))
}

func (s *Server) findRepoConfig(name string) *config.RepoConfig {
	for _, repo := range s.Config.Repositories {
		if repo.Name == name {
			return &repo
		}
	}

	return nil
}

func (s *Server) handlePullRequestEvent(event *github.PullRequestEvent, repo *config.RepoConfig) error {
	ctx := context.TODO()

	switch event.GetAction() {
	case "opened", "reopened", "synchronize":
		return s.applyResources(ctx, event, repo)
	case "closed":
		return s.deleteResources(ctx, event, repo)
	}

	return nil
}

func (s *Server) buildResource(event *github.PullRequestEvent, res *config.ResourceConfig) *kubernetes.Resource {
	return &kubernetes.Resource{
		ResourceConfig:    *res,
		PullRequestNumber: *event.Number,
		HeadCommitSHA:     *event.PullRequest.Head.SHA,
	}
}

func (s *Server) applyResources(ctx context.Context, event *github.PullRequestEvent, repo *config.RepoConfig) error {
	for _, res := range repo.Resources {
		if err := s.KubernetesClient.Apply(ctx, s.buildResource(event, &res)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) deleteResources(ctx context.Context, event *github.PullRequestEvent, repo *config.RepoConfig) error {
	for _, res := range repo.Resources {
		if err := s.KubernetesClient.Delete(ctx, s.buildResource(event, &res)); err != nil {
			return err
		}
	}

	return nil
}
