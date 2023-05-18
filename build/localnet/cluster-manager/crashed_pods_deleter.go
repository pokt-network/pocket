package main

// Monitors Pods created by StatefulSets, and if the Pods are in a `CrashLoopBackOff` status,
// and they have a different image tag - kill them. StatefulSet would then recreate the Pod with a new image.

import (
	"context"
	"errors"
	"strings"

	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	appstypedv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coretypedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Loop through existing pods and set up a watch for new Pods so we don't hit Kubernetes API all the time
func initCrashedPodsDeleter(client *kubernetes.Clientset) {
	stsClient := client.AppsV1().StatefulSets(pocketk8s.CurrentNamespace)
	podClient := client.CoreV1().Pods(pocketk8s.CurrentNamespace)

	// Loop through all existing Pods and delete the ones that are in CrashLoopBackOff status
	podList, err := podClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("error listing pods on init")
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		if err := deleteCrashedPods(pod, stsClient, podClient); err != nil {
			logger.Error().Err(err).Msg("error deleting crashed pod on init")
		}
	}

	// Set up a watch for new Pods
	w, err := podClient.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("error setting up watch for new pods")
	}
	for event := range w.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				logger.Error().Msg("error casting pod on watch")
				continue
			}

			if err := deleteCrashedPods(pod, stsClient, podClient); err != nil {
				logger.Error().Err(err).Msg("error deleting crashed pod on watch")
			}
		}
	}
}

func containerErroneousStatus(status *corev1.ContainerStatus) bool {
	return status.State.Waiting != nil &&
		(strings.HasPrefix(status.State.Waiting.Reason, "Err") ||
			strings.HasSuffix(status.State.Waiting.Reason, "BackOff"))
}

func deleteCrashedPods(
	pod *corev1.Pod,
	stsClient appstypedv1.StatefulSetInterface,
	podClient coretypedv1.PodInterface,
) error {
	// If annotation is present, we monitor the Pod
	containerToMonitor, ok := pod.Annotations["cluster-manager-delete-on-crash-container"]
	if !ok {
		return nil
	}

	for ci := range pod.Spec.Containers {
		podContainer := &pod.Spec.Containers[ci]

		// Only proceed if container is the one we monitor
		if podContainer.Name != containerToMonitor {
			continue
		}

		for si := range pod.Status.ContainerStatuses {
			containerStatus := &pod.Status.ContainerStatuses[si]

			// Only proceed if container is in some sort of Err status
			if containerErroneousStatus(containerStatus) {
				// Get StatefulSet that created the Pod
				var stsName string
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "StatefulSet" {
						stsName = ownerRef.Name
						break
					}
				}

				if stsName == "" {
					return errors.New("no StatefulSet found for this pod")
				}

				sts, err := stsClient.Get(context.TODO(), stsName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				// Loop through all containers in the StatefulSet and find the one we monitor
				for sci := range sts.Spec.Template.Spec.Containers {
					stsContainer := &sts.Spec.Template.Spec.Containers[sci]
					if stsContainer.Name != containerToMonitor {
						continue
					}

					// If images are different, delete the Pod
					if stsContainer.Image != podContainer.Image {
						deletePolicy := metav1.DeletePropagationForeground

						if err := podClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
							PropagationPolicy: &deletePolicy,
						}); err != nil {
							return err
						}

						logger.Info().Str("pod", pod.Name).Msg("deleted crashed pod")
					} else {
						logger.Info().Str("pod", pod.Name).Msg("pod crashed, but image is the same, not deleting")
					}

				}
			}
		}
	}

	return nil
}
