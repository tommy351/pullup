package webhook

import (
	"net/http"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (s *Server) Webhook(w http.ResponseWriter, r *http.Request) error {
	hook := new(v1alpha1.Webhook)

	err := s.Client.Get(r.Context(), types.NamespacedName{
		Namespace: s.Namespace,
		Name:      Params(r)["name"],
	}, hook)

	if err != nil {
		if errors.IsNotFound(err) {
			return String(w, http.StatusNotFound, "Webhook not found")
		}

		return xerrors.Errorf("failed to get webhook: %w", err)
	}

	logger := log.FromContext(r.Context()).WithValues("webhook", hook)
	r = r.WithContext(log.NewContext(r.Context(), logger))
	hook.SetGroupVersionKind(k8s.Kind("Webhook"))

	if hook.Spec.GitHub != nil {
		return s.webhookGithub(w, r, hook)
	}

	return String(w, http.StatusBadRequest, "Unsupported webhook type")
}
