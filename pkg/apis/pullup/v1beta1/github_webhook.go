package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type GitHubWebhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status GitHubWebhookStatus `json:"status,omitempty"`
	Spec   GitHubWebhookSpec   `json:"spec,omitempty"`
}

func (in *GitHubWebhook) GetSpec() WebhookSpec {
	return in.Spec.WebhookSpec
}

// +kubebuilder:object:root=true

type GitHubWebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GitHubWebhook `json:"items"`
}

type GitHubWebhookSpec struct {
	WebhookSpec `json:",inline"`

	Repositories []GitHubRepository `json:"repositories"`
}

type GitHubRepository struct {
	Name        string                        `json:"name"`
	Push        *GitHubPushEventFilter        `json:"push,omitempty"`
	PullRequest *GitHubPullRequestEventFilter `json:"pullRequest,omitempty"`
}

type GitHubPushEventFilter struct {
	Branches *WebhookFilter `json:"branches,omitempty"`
	Tags     *WebhookFilter `json:"tags,omitempty"`
}

type GitHubPullRequestEventFilter struct {
	Branches *WebhookFilter               `json:"branches,omitempty"`
	Types    []GitHubPullRequestEventType `json:"types,omitempty"`
}

// +kubebuilder:validation:Enum=assigned;unassigned;labeled;unlabeled;opened;edited;closed;reopened;synchronize;ready_for_review;locked;unlocked;review_requested;review_request_removed
type GitHubPullRequestEventType string

type GitHubWebhookStatus struct {
	WebhookStatus `json:",inline"`
}
