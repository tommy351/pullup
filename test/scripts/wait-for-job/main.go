package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	namespace := flag.String("namespace", "default", "kubernetes namespace")
	jobName := flag.String("job", "", "job name")
	kubeconfig := flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "path to the kubeconfig file")

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	if err := printJobLogs(client, *namespace, *jobName); err != nil {
		panic(err)
	}
}

func getPodOfTestJob(client *kubernetes.Clientset, namespace, name string) (*corev1.Pod, error) {
	list, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "job-name=" + name,
	})

	if err != nil {
		return nil, xerrors.Errorf("failed to list pods: %w", err)
	}

	var runningPods []corev1.Pod

	for _, pod := range list.Items {
		if pod.Status.Phase == corev1.PodRunning {
			runningPods = append(runningPods, pod)
		}
	}

	if len(runningPods) == 0 {
		return nil, nil
	}

	return &runningPods[0], nil
}

func printJobLogs(client *kubernetes.Clientset, namespace, name string) error {
	var pod *corev1.Pod
	backoff := wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    10,
	}

	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		var err error
		pod, err = getPodOfTestJob(client, namespace, name)

		if err != nil {
			return false, xerrors.Errorf("failed to get pod: %w", err)
		}

		return pod != nil, nil
	})

	if err != nil {
		return xerrors.Errorf("wait failed: %w", err)
	}

	fmt.Printf("Watching logs of pod %q\n", pod.Name)

	req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow: true,
	})

	reader, err := req.Stream()

	if err != nil {
		return xerrors.Errorf("failed to open the stream: %w", err)
	}

	defer reader.Close()

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return xerrors.Errorf("failed to write logs: %w", err)
	}

	return nil
}
