// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package endpoints

import (
	"context"
	"testing"

	logrtest "github.com/go-logr/logr/testr"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/hashicorp/consul-k8s/control-plane/connect-inject/constants"
)

func TestCreateServiceRegistrationsNodeLabelMetadataMultiport(t *testing.T) {
	t.Parallel()

	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{
		Name:      nodeName,
		Namespace: "default",
		Labels:    map[string]string{"rack": "rack-1"},
	}}
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}
	controller := &Controller{
		Client:  fake.NewClientBuilder().WithObjects(node, namespace).Build(),
		Context: context.Background(),
		Log:     logrtest.New(t),
	}
	pod := *createServicePod("pod1", "1.2.3.4", true, true)
	pod.Annotations[constants.AnnotationPort] = "60606,9314"
	pod.Annotations[constants.AnnotationService] = "udpprober-server,udpprober-server-metrics"
	pod.Annotations[constants.AnnotationMeta+"rack"] = "$NODE_LABEL:rack"
	pod.Annotations[constants.AnnotationMeta+"missing"] = "$NODE_LABEL:missing"
	pod.Annotations[constants.AnnotationMeta+"malformed"] = "$NODE_LABEL:"

	for _, serviceName := range []string{"udpprober-server", "udpprober-server-metrics"} {
		endpoint := corev1.Endpoints{ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: "default"}}
		service, proxy, err := controller.createServiceRegistrations(pod, pod.Status.PodIP, endpoint, api.HealthPassing)
		require.NoError(t, err)
		require.Equal(t, "rack-1", service.Service.Meta["rack"])
		require.Equal(t, "rack-1", proxy.Service.Meta["rack"])
		require.NotContains(t, service.Service.Meta, "missing")
		require.NotContains(t, service.Service.Meta, "malformed")
	}
}
