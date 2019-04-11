package webhook

import (
	"net/http"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) Webhook(w http.ResponseWriter, r *http.Request) error {
	name := Params(r)["name"]
	hook, err := s.Client.PullupV1alpha1().Webhooks(s.Namespace).Get(name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			return String(w, http.StatusNotFound, "Webhook not found")
		}

		return xerrors.Errorf("failed to get webhook: %w", err)
	}

	hook.SetGroupVersionKind(v1alpha1.Kind("Webhook"))

	if hook.Spec.GitHub != nil {
		return s.webhookGithub(w, r, hook)
	}

	return String(w, http.StatusBadRequest, "Unsupported webhook type")
}
