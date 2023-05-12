package main

// Monitors Pods created by StatefulSets, and if Pods are in `CrashLoopBackOff` status,
// and they have different image tag - kill them so StatefulSet would recreate the Pod with new image.

import (
	"context"
	"errors"

	pocketk8s "github.com/pokt-network/pocket/shared/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	appstypedv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coretypedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Loop through existing pods and set up a watch for new pods so we don't hit Kubernetes API all the time
func initDeleteCrashedPods(client *kubernetes.Clientset) {
	stsClient := client.AppsV1().StatefulSets(pocketk8s.CurrentNamespace)
	podClient := client.CoreV1().Pods(pocketk8s.CurrentNamespace)

	// Loop through all existing Pods and delete the ones that are in CrashLoopBackOff status
	podList, _ := podClient.List(context.TODO(), metav1.ListOptions{})
	for _, pod := range podList.Items {
		deleteCrashedPods(&pod, stsClient, podClient)
	}

	// Set up a watch for new Pods
	w, _ := podClient.Watch(context.TODO(), metav1.ListOptions{})
	for event := range w.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			deleteCrashedPods(pod, stsClient, podClient)
		}
	}
}

func deleteCrashedPods(pod *corev1.Pod, stsClient appstypedv1.StatefulSetInterface, podClient coretypedv1.PodInterface) error {
	// If annotation is present, we monitor the Pod
	if containerToMonitor, ok := pod.Annotations["cluster-manager-delete-on-crash-container"]; ok {
		for _, podContainer := range pod.Spec.Containers {
			if podContainer.Name == containerToMonitor {
				for _, containerStatus := range pod.Status.ContainerStatuses {

					// Only proceed if container is in CrashLoopBackOff status
					if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
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
						for _, stsContainer := range sts.Spec.Template.Spec.Containers {
							if stsContainer.Name == containerToMonitor {

								// Loop through all containers in the Pod and find the one we monitor
								// If images are different, delete the Pod
								if stsContainer.Image != podContainer.Image {
									deletePolicy := metav1.DeletePropagationForeground

									if err := podClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
										PropagationPolicy: &deletePolicy,
									}); err != nil {
										return err
									} else {
										logger.Info().Str("pod", pod.Name).Msg("deleted crashed pod")
									}
								} else {
									logger.Info().Str("pod", pod.Name).Msg("pod crashed, but image is the same, not deleting")
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}
