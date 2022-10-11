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
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/test/e2e/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = SIGDescribe("Keystone Containers [Feature:KeystoneContainers]", func() {
	f := framework.NewDefaultFramework("keystone-container-test")

	ginkgo.Context("When creating a job with two containers", func() {

		ginkgo.It("should complete the job once the keystone container exits successfully [Keystone]", func() {
			jobClient := f.ClientSet.BatchV1().Jobs(f.Namespace.Name)
			podClient := f.PodClient()
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-job",
				},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyOnFailure,
							Containers: []v1.Container{
								// the main container should exit before the sidecar with 0 exit code
								{
									Name:    "main-container",
									Image:   imageutils.GetE2EImage(imageutils.BusyBox),
									Command: []string{"sh", "-c", "sleep 1 && exit 0"},
								},
								{
									Name:    "sidecar-container",
									Image:   imageutils.GetE2EImage(imageutils.BusyBox),
									Command: []string{"sh", "-c", "sleep infinity"},
								},
							},
						},
					},
				},
			}

			j, err := jobClient.Create(context.TODO(), job, metav1.CreateOptions{})
			framework.ExpectNoError(err, "error while creating the job")

			p, err := podClient.List(context.TODO(), metav1.ListOptions{})
			framework.ExpectNotEqual(len(p.Items), 0)

			// it is expected that the pod succeeds and the job should have a completed
			// status eventually even if the sidecar container has not terminated in the pod
			gomega.Eventually(func() bool {
				j, err = jobClient.Get(context.TODO(), j.Name, metav1.GetOptions{})
				framework.ExpectNoError(err, "error while getting job")
				framework.Logf("Job : %+v", j)
				for _, c := range j.Status.Conditions {
					if c.Type == batchv1.JobComplete && c.Status == v1.ConditionTrue {
						return true
					}
					time.Sleep(5 * time.Second)
				}
				return false
			}, time.Minute).Should(gomega.BeTrue())

		})

	})
})
