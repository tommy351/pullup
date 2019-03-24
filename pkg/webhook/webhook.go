package webhook

import (
	"net/http"

	"github.com/tommy351/pullup/pkg/k8s"
	"golang.org/x/xerrors"
)

func (s *Server) Webhook(w http.ResponseWriter, r *http.Request) error {
	name := Params(r)["name"]
	hook, err := s.Client.GetWebhook(r.Context(), name)

	if err != nil {
		if k8s.IsNotFoundError(err) {
			return String(w, http.StatusNotFound, "Webhook not found")
		}

		return xerrors.Errorf("failed to get webhook: %w", err)
	}

	if hook.Spec.GitHub != nil {
		return s.webhookGithub(w, r, hook)
	}

	return String(w, http.StatusBadRequest, "Unsupported webhook type")
}
