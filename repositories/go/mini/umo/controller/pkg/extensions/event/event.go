package event

import corev1 "k8s.io/api/core/v1"

type Event struct {
}

type Node struct {
	Name   string
	OldPod corev1.Pod
	NewPod corev1.Pod
}
type Listener interface {
	OnEvent(event Event)
	GetName() string
}

type WebhookListener struct {
}

func NewWebhookListener() *WebhookListener {
	return &WebhookListener{}
}

func (w *WebhookListener) OnEvent(event Event) {
}

func (w *WebhookListener) GetName() string {
	return "webhook"
}
