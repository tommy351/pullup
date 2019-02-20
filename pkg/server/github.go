package server

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/ansel1/merry"
	"github.com/google/go-github/v24/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/kubernetes"
)

func (s *Server) GitHubWebhook(w http.ResponseWriter, r *http.Request) error {
	logger := hlog.FromRequest(r)
	payload, err := s.validatePayload(r)

	if err != nil {
		logger.Warn().Err(err).Msg("Failed to validate payload")
		return ErrInvalidPayload
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)

	if err != nil {
		logger.Warn().Err(err).Msg("Failed to parse webhook")
		return ErrInvalidPayload
	}

	logger.Debug().Interface("payload", event).Msg("Received GitHub webhook")

	if event, ok := event.(*github.PullRequestEvent); ok {
		repo := s.findRepoConfig("github.com/" + *event.Repo.FullName)

		if repo == nil {
			return ErrRepositoryNotFound
		}

		if err := s.handlePullRequestEvent(r.Context(), event, repo); err != nil {
			return merry.Wrap(err)
		}
	}

	return String(w, http.StatusOK, "ok")
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
		repo := repo

		if repo.Name == name {
			return &repo
		}
	}

	return nil
}

func (s *Server) handlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent, repo *config.RepoConfig) error {
	logger := zerolog.Ctx(ctx).With().Str("repository", repo.Name).Logger()
	ctx = logger.WithContext(ctx)

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
		res := res

		if err := s.KubernetesClient.Apply(ctx, s.buildResource(event, &res)); err != nil {
			return merry.Wrap(err)
		}
	}

	return nil
}

func (s *Server) deleteResources(ctx context.Context, event *github.PullRequestEvent, repo *config.RepoConfig) error {
	for _, res := range repo.Resources {
		res := res

		if err := s.KubernetesClient.Delete(ctx, s.buildResource(event, &res)); err != nil {
			return merry.Wrap(err)
		}
	}

	return nil
}
