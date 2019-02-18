package api

import (
	"net/http"

	"github.com/google/go-github/v24/github"
	"github.com/labstack/echo/v4"
	"github.com/tommy351/pullup/pkg/config"
)

func (s *Server) GitHubWebhook(ctx echo.Context) error {
	req := ctx.Request()
	payload, err := github.ValidatePayload(req, []byte(s.Config.GitHub.Secret))

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

func (s *Server) findRepoConfig(name string) *config.RepoConfig {
	for _, repo := range s.Config.Repositories {
		if repo.Name == name {
			return &repo
		}
	}

	return nil
}

func (s *Server) handlePullRequestEvent(event *github.PullRequestEvent, repo *config.RepoConfig) error {
	switch event.GetAction() {
	case "opened", "reopened":
		return s.createResources(repo)
	case "synchronize":
		return s.updateResources(repo)
	case "closed":
		return s.destroyResources(repo)
	}

	return nil
}

func (s *Server) createResources(repo *config.RepoConfig) error {
	return nil
}

func (s *Server) updateResources(repo *config.RepoConfig) error {
	return nil
}

func (s *Server) destroyResources(repo *config.RepoConfig) error {
	return nil
}
