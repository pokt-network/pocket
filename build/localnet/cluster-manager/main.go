package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watcher, err := clientset.CoreV1().Services("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan() {
		service, ok := event.Object.(*v1.Service)
		if !ok {
			continue
		}

		if !isValidator(service) {
			continue
		}

		switch event.Type {
		case watch.Added:
			fmt.Printf("Validator %s added\n", service.Name)
		case watch.Modified:
			fmt.Printf("Validator %s modified\n", service.Name)
		case watch.Deleted:
			fmt.Printf("Validator %s deleted\n", service.Name)
		}
	}
}

func isValidator(service *v1.Service) bool {
	return service.Labels["v1-purpose"] == "validator"
}
