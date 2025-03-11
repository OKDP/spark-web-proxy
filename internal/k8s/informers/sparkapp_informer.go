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
	"fmt"
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
		log.Info("Running kubernetes informer on all namespaces")
	} else {
		log.Info("Running kubernetes informer on the following namespaces: %s", namespaces)
	}

	for _, ns := range namespaces {
		go i.WatchNamespaceSparkApps(clientset, ns)
	}
}

func (i SparkAppInformer) WatchNamespaceSparkApps(clientset *kubernetes.Clientset, namespace string) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("Spark app informer on namespace '%s' failed, restarting... \n %+v", namespace, r)
					time.Sleep(5 * time.Second)
				}
			}()

			factory := informers.NewSharedInformerFactoryWithOptions(clientset, 5*time.Minute,
				informers.WithNamespace(namespace),
				informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
					opts.LabelSelector = "spark-role=driver"
				}),
			)

			podInformer := factory.Core().V1().Pods().Informer()

			// Register event handlers
			podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    i.sparkAppAddedOrUpdated,
				UpdateFunc: func(oldObj, newObj interface{}) { i.sparkAppAddedOrUpdated(newObj) },
				DeleteFunc: i.sparkAppDeleted,
			})

			stopCh := make(chan struct{})
			defer close(stopCh)

			log.Info("Starting Spark app informer for namespace: %s", namespace)
			factory.Start(stopCh)
			factory.WaitForCacheSync(stopCh)
			<-stopCh
		}()
	}
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
		InternalUrl: fmt.Sprintf("http://%s:%d", pod.Status.PodIP, i.ui.Port),
		Namespace:   pod.Namespace,
		Status:      string(pod.Status.Phase),
	}

	model.AddOrUpdateSparkApp(appID, sparkJob)
	log.Info("Spark app : %s/%s (%s) -> %s", sparkJob.Namespace, appID, sparkJob.Status, sparkJob.InternalUrl)
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
