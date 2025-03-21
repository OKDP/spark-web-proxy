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

package utils

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func GetSparkUIPort(pod *corev1.Pod) int32 {
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if strings.Contains(strings.ToLower(port.Name), "ui") {
				return port.ContainerPort
			}
		}
	}
	return 4040
}

func GetSparkAppID(pod *corev1.Pod) string {
	for _, container := range pod.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "SPARK_APPLICATION_ID" {
				return envVar.Value
			}
		}
	}
	return "-1"
}
