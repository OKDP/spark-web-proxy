/*
 *    Copyright 2025 okdp.io
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package informers

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/okdp/spark-history-web-proxy/internal/config"
	log "github.com/okdp/spark-history-web-proxy/internal/logging"
	"github.com/okdp/spark-history-web-proxy/internal/model"
)

type SparkAppInformer struct {
	namespaces []string
	ui         config.UI
}

func NewSparkAppInformer(config *config.ApplicationConfig) *SparkAppInformer {
	return &SparkAppInformer{
		namespaces: config.Spark.JobNamespaces,
		ui:         config.Spark.UI,
	}
}

func (i SparkAppInformer) WatchSparkApps(clientset *kubernetes.Clientset) {
	namespaces := i.namespaces
	if len(namespaces) == 0 {
		namespaces = []string{metav1.NamespaceAll}
	}

	for _, ns := range namespaces {
		go i.WatchNamespaceSparkApps(clientset, ns)
	}
}

func (i SparkAppInformer) WatchNamespaceSparkApps(clientset *kubernetes.Clientset, namespace string) {

	log.Info("Running spark app informer on the following namespaces: %s", func() string {
		if namespace == metav1.NamespaceAll {
			return "all"
		}
		return namespace
	}())

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 5*time.Minute,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
			opts.LabelSelector = "spark-role=driver"
		}),
	)

	podInformer := factory.Core().V1().Pods().Informer()

	// Register event handlers
	registration, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    i.sparkAppAddedOrUpdated,
		UpdateFunc: func(_, newObj interface{}) { i.sparkAppAddedOrUpdated(newObj) },
		DeleteFunc: i.sparkAppDeleted,
	})

	if err != nil {
		log.Error("Failed to add spark app event handler: %v", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	factory.Start(ctx.Done())

	<-ctx.Done()

	log.Info("Received shutdown signal. Stopping Spark app informer...")
	_ = podInformer.RemoveEventHandler(registration)
	log.Info("Spark app informer successfully stopped.")
}

func (i SparkAppInformer) sparkAppAddedOrUpdated(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	appID := getSparkAppID(pod)
	if appID == "" {
		return
	}

	sparkJob := model.SparkApp{
		InternalURL: fmt.Sprintf("http://%s:%d", pod.Status.PodIP, i.ui.Port),
		Namespace:   pod.Namespace,
		Status:      string(pod.Status.Phase),
	}

	model.AddOrUpdateSparkApp(appID, sparkJob)
	log.Info("Spark app updated: %s/%s (%s) -> %s", sparkJob.Namespace, appID, sparkJob.Status, sparkJob.InternalURL)
}

func (i SparkAppInformer) sparkAppDeleted(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	appID := getSparkAppID(pod)
	if appID == "" {
		return
	}

	model.DeleteSparkApp(appID)
	log.Info("Removed Spark App '%s' on namespace '%s' (appID: %s)", pod.Name, pod.Namespace, appID)
}

func getSparkAppID(pod *corev1.Pod) string {
	for _, container := range pod.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "SPARK_APPLICATION_ID" {
				return envVar.Value
			}
		}
	}
	return ""
}
