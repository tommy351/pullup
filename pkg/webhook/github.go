package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v24/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/pointer"
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
			err := s.applyResourceSet(r.Context(), &v1alpha1.ResourceSet{
				TypeMeta: k8s.GVKToTypeMeta(k8s.Kind("ResourceSet")),
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: hook.Namespace,
					Labels: map[string]string{
						k8s.LabelWebhookName:       hook.Name,
						k8s.LabelPullRequestNumber: strconv.Itoa(event.GetNumber()),
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         hook.APIVersion,
							Kind:               hook.Kind,
							Name:               hook.Name,
							UID:                hook.UID,
							Controller:         pointer.BoolPtr(true),
							BlockOwnerDeletion: pointer.BoolPtr(true),
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
			err := s.Client.PullupV1alpha1().ResourceSets(hook.Namespace).Delete(name, &metav1.DeleteOptions{})

			if err != nil && !errors.IsNotFound(err) {
				return xerrors.Errorf("failed to delete resource set %s: %w", name, err)
			}

			logger.Debug().
				Dict("resourceSet", zerolog.Dict().
					Str("name", name).
					Str("namespace", hook.Namespace)).
				Msg("Deleted resource set")
		}
	}

	return NoContent(w)
}

func (s *Server) applyResourceSet(ctx context.Context, rs *v1alpha1.ResourceSet) error {
	client := s.Client.PullupV1alpha1().ResourceSets(s.Namespace)
	logger := zerolog.Ctx(ctx).With().
		Dict("resourceSet", zerolog.Dict().
			Str("name", rs.Name).
			Str("namespace", rs.Namespace)).
		Logger()

	if _, err := client.Create(rs); err == nil {
		logger.Debug().Msg("Created resource set")
		return nil
	} else if !errors.IsAlreadyExists(err) {
		return xerrors.Errorf("failed to create resource set: %w", err)
	}

	patch, err := json.Marshal([]k8s.JSONPatch{
		{
			Op:    "replace",
			Path:  "/spec",
			Value: rs.Spec,
		},
	})

	if err != nil {
		return xerrors.Errorf("failed to marshal resource set spec: %w", err)
	}

	if _, err := client.Patch(rs.Name, types.JSONPatchType, patch); err != nil {
		return xerrors.Errorf("failed to patch resource set: %w", err)
	}

	logger.Debug().Msg("Updated resource set")
	return nil
}

func parseGithubWebhook(r *http.Request, options *v1alpha1.GitHubOptions) (interface{}, error) {
	payload, err := github.ValidatePayload(r, []byte(options.Secret))

	if err != nil {
		return nil, err
	}

	return github.ParseWebHook(github.WebHookType(r), payload)
}
