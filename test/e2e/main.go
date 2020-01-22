package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v29/github"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/pointer"
)

// nolint: gochecknoglobals
var (
	webhookHost = os.Getenv("WEBHOOK_SERVICE_NAME")
	webhookName = os.Getenv("WEBHOOK_NAME")
	backoff     = wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    10,
	}
	logger *zap.Logger
	event  *github.PullRequestEvent
)

func main() {
	var err error
	logger, err = zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))

	if err != nil {
		panic(err)
	}

	event = newPullRequestEvent()

	if err := waitUntilWebhookReady(); err != nil {
		logger.Fatal("Webhook is not ready", zap.Error(err))
	}

	if err := triggerWebhook(); err != nil {
		logger.Fatal("Failed to trigger the webhook", zap.Error(err))
	}

	if err := validateService(); err != nil {
		logger.Fatal("Failed to validate the service", zap.Error(err))
	}
}

func waitUntilWebhookReady() error {
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		res, err := http.Get(fmt.Sprintf("http://%s/healthz", webhookHost))

		if err != nil {
			logger.Warn("Webhook is not ready yet", zap.Error(err))
			return false, nil
		}

		// nolint: errcheck
		defer res.Body.Close()

		return res.StatusCode == http.StatusOK, nil
	})

	if err != nil {
		return fmt.Errorf("webhook is not ready: %w", err)
	}

	logger.Info("Webhook is ready")
	return nil
}

func newPullRequestEvent() *github.PullRequestEvent {
	prNumber := 46
	repoOwner := "foo"
	repoName := "bar"
	repoFullName := fmt.Sprintf("%s/%s", repoOwner, repoName)

	return &github.PullRequestEvent{
		Action: pointer.StringPtr("opened"),
		Number: &prNumber,
		Repo: &github.Repository{
			Name:     &repoName,
			Owner:    &github.User{Login: &repoOwner},
			FullName: &repoFullName,
		},
		PullRequest: &github.PullRequest{
			Number: &prNumber,
			Base: &github.PullRequestBranch{
				Ref: pointer.StringPtr("base"),
				SHA: pointer.StringPtr("b436f6eb3356504235c0c9a8e74605c820d8d9cc"),
			},
			Head: &github.PullRequestBranch{
				Ref: pointer.StringPtr("test"),
				SHA: pointer.StringPtr("0ce4cf0450de14c6555c563fa9d36be67e69aa2f"),
			},
		},
	}
}

func triggerWebhook() error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(event); err != nil {
		return fmt.Errorf("failed to encode the body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/webhooks/github", webhookHost), &buf)

	if err != nil {
		return fmt.Errorf("failed to build the request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Delivery", uuid.Must(uuid.NewRandom()).String())
	req.Header.Set("X-GitHub-Event", "pull_request")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to send the request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("http response error: %d", res.StatusCode)
	}

	logger.Info("Webhook is triggered")
	return nil
}

func validateService() error {
	name := fmt.Sprintf("%s-%d", webhookName, event.GetNumber())

	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		res, err := http.Get(fmt.Sprintf("http://%s/test", name))

		if err != nil {
			logger.Warn("Service is not ready yet", zap.Error(err))
			return false, nil
		}

		// nolint: errcheck
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			if v := res.Header.Get("X-Resource-Name"); v != name {
				return false, fmt.Errorf("expected header X-Resource-Name to be %s, got %q", name, v)
			}

			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("invalid service: %w", err)
	}

	logger.Info("Service is valid")
	return nil
}
