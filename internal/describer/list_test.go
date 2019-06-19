/*
Copyright (c) 2019 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package describer

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	configFake "github.com/heptio/developer-dash/internal/config/fake"
	printerfake "github.com/heptio/developer-dash/internal/modules/overview/printer/fake"
	"github.com/heptio/developer-dash/internal/testutil"
	"github.com/heptio/developer-dash/pkg/store"
	"github.com/heptio/developer-dash/pkg/plugin"
	"github.com/heptio/developer-dash/pkg/view/component"
)

func TestListDescriber(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	thePath := "/"

	pod := testutil.CreatePod("pod")
	pod.CreationTimestamp = metav1.Time{
		Time: time.Unix(1547472896, 0),
	}

	key, err := store.KeyFromObject(pod)
	require.NoError(t, err)

	ctx := context.Background()
	namespace := "default"

	dashConfig := configFake.NewMockDash(controller)
	pluginManager := plugin.NewManager(nil)
	dashConfig.EXPECT().PluginManager().Return(pluginManager)

	podListTable := createPodTable(*pod)

	objectPrinter := printerfake.NewMockPrinter(controller)
	podList := &corev1.PodList{Items: []corev1.Pod{*pod}}
	objectPrinter.EXPECT().Print(gomock.Any(), podList, pluginManager).Return(podListTable, nil)

	options := Options{
		Dash:    dashConfig,
		Printer: objectPrinter,
		LoadObjects: func(ctx context.Context, namespace string, fields map[string]string, objectStoreKeys []store.Key) ([]*unstructured.Unstructured, error) {
			return testutil.ToUnstructuredList(t, pod), nil
		},
	}

	d := NewList(thePath, "list", key, podListType, podObjectType, false)
	cResponse, err := d.Describe(ctx, "/path", namespace, options)
	require.NoError(t, err)

	list := component.NewList("list", nil)
	list.Add(podListTable)
	expected := component.ContentResponse{
		Components: []component.Component{list},
	}

	assert.Equal(t, expected, cResponse)
}