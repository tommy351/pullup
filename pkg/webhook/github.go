package webhook

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v24/github"
	"github.com/rs/zerolog/hlog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) webhookGithub(w http.ResponseWriter, r *http.Request, hook *v1alpha1.Webhook) error {
	logger := hlog.FromRequest(r)
	payload, err := parseGithubWebhook(r, hook.Spec.GitHub)

	if err != nil {
		logger.Warn().Err(err).Msg("Invalid webhook")
		return String(w, http.StatusBadRequest, "Invalid webhook")
	}

	if event, ok := payload.(*github.PullRequestEvent); ok {
		name := fmt.Sprintf("%s-%d", hook.Name, event.GetNumber())

		switch event.GetAction() {
		case "opened", "reopened", "synchronize":
			err := s.Client.ApplyResourceSet(r.Context(), &v1alpha1.ResourceSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       "ResourceSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: hook.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         hook.APIVersion,
							Kind:               hook.Kind,
							Name:               hook.Name,
							UID:                hook.UID,
							Controller:         k8s.BoolP(true),
							BlockOwnerDeletion: k8s.BoolP(true),
						},
					},
				},
				Spec: v1alpha1.ResourceSetSpec{
					Resources: hook.Spec.Resources,
					Number:    event.GetNumber(),
					Base: &v1alpha1.Commit{
						Ref: event.PullRequest.Base.Ref,
						SHA: event.PullRequest.Base.SHA,
					},
					Head: &v1alpha1.Commit{
						Ref: event.PullRequest.Head.Ref,
						SHA: event.PullRequest.Head.SHA,
					},
					Merge: &v1alpha1.Commit{
						SHA: event.PullRequest.MergeCommitSHA,
					},
				},
			})

			if err != nil {
				return xerrors.Errorf("failed to apply resource set %s: %w", name, err)
			}

		case "closed":
			if err := s.Client.DeleteResourceSet(r.Context(), name); err != nil {
				if !k8s.IsNotFoundError(err) {
					return xerrors.Errorf("failed to delete resource set %s: %w", name, err)
				}
			}
		}
	}

	return NoContent(w)
}

func parseGithubWebhook(r *http.Request, options *v1alpha1.GitHubOptions) (interface{}, error) {
	payload, err := github.ValidatePayload(r, []byte(options.Secret))

	if err != nil {
		return nil, err
	}

	return github.ParseWebHook(github.WebHookType(r), payload)
}
