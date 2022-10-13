/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node

import (
	"fmt"
	"github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/test/e2e/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"
	admissionapi "k8s.io/pod-security-admission/api"
)

var _ = SIGDescribe("Keystone Containers [Feature:KeystoneContainers]", func() {
	f := framework.NewDefaultFramework("keystone-container-test")
	f.NamespacePodSecurityEnforceLevel = admissionapi.LevelPrivileged
	var podClient *framework.PodClient
	ginkgo.BeforeEach(func() {
		podClient = f.PodClient()
	})

	ginkgo.Context("When creating a pod with two containers", func() {

		ginkgo.It("should delete the pod once the keystone container exits successfully [Keystone]", func() {
			keystone := "Keystone"
			podName := fmt.Sprintf("keystone-test-pod-%s", uuid.NewUUID())
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: podName,
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyOnFailure,
					Containers: []v1.Container{
						// the main container should exit before the sidecar with 0 exit code
						{
							Name:    "main-container",
							Image:   imageutils.GetE2EImage(imageutils.BusyBox),
							Command: []string{"sh", "-c", "sleep 10 && exit 0"},
							Lifecycle: &v1.Lifecycle{
								Type: &keystone,
							},
						},
						{
							Name:    "sidecar-container",
							Image:   imageutils.GetE2EImage(imageutils.BusyBox),
							Command: []string{"sh", "-c", "sleep 3600"},
						},
					},
				},
			}

			// create the pod and wait for it to be in running state
			podClient.CreateSync(pod)

			// the pod should succeed when the main container exits
			podClient.WaitForSuccess(podName, framework.PodStartTimeout)

		})

	})
})
