package github

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/google/go-github/v24/github"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"golang.org/x/xerrors"
)

const defaultResourceName = "{{ .Name }}-{{ .Number }}"

type nameTemplateData struct {
	Name      string
	RepoOwner string
	RepoName  string
	Number    int
}

func getResourceName(event *github.PullRequestEvent, hook *v1alpha1.Webhook) (string, error) {
	resourceName := hook.Spec.ResourceName

	if resourceName == "" {
		resourceName = defaultResourceName
	}

	tmpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(resourceName)

	if err != nil {
		return "", xerrors.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	data := &nameTemplateData{
		Name:      hook.Name,
		RepoOwner: event.Repo.GetOwner().GetLogin(),
		RepoName:  event.Repo.GetName(),
		Number:    event.GetNumber(),
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", xerrors.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
