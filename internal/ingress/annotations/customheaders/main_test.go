/*
Copyright 2023 The Kubernetes Authors.

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

package customheaders

import (
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-backend",
					Port: networking.ServiceBackendPort{
						Number: 80,
					},
				},
			},
		},
	}
}

type mockBackend struct {
	resolver.Mock
}

// GetDefaultBackend returns the backend that must be used as default
func (m mockBackend) GetDefaultBackend() defaults.Backend {
	return defaults.Backend{
		AllowedResponseHeaders: []string{"Content-Type", "Access-Control-Max-Age"},
	}
}

func TestCustomHeadersParseInvalidAnnotations(t *testing.T) {
	ing := buildIngress()
	configMapResolver := mockBackend{}
	configMapResolver.ConfigMaps = map[string]*api.ConfigMap{}

	_, err := NewParser(configMapResolver).Parse(ing)
	if err != nil {
		t.Errorf("expected error parsing ingress with custom-response-headers")
	}

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("custom-headers")] = "custom-headers-configmap"
	ing.SetAnnotations(data)
	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err == nil {
		t.Errorf("expected error parsing ingress with custom-response-headers")
	}
	if i != nil {
		t.Errorf("expected %v but got %v", nil, i)
	}
}

func TestCustomHeadersParseAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("custom-headers")] = "custom-headers-configmap"
	ing.SetAnnotations(data)

	configMapResolver := mockBackend{}
	configMapResolver.ConfigMaps = map[string]*api.ConfigMap{}

	configMapResolver.ConfigMaps["custom-headers-configmap"] = &api.ConfigMap{Data: map[string]string{"Content-Type": "application/json", "Access-Control-Max-Age": "600"}}

	i, err := NewParser(configMapResolver).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with custom-response-headers: %s", err)
	}
	val, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a *Config type")
	}

	expectedResponseHeaders := map[string]string{}
	expectedResponseHeaders["Content-Type"] = "application/json"
	expectedResponseHeaders["Access-Control-Max-Age"] = "600"

	c := &Config{expectedResponseHeaders}

	if !reflect.DeepEqual(c, val) {
		t.Errorf("expected %v but got %v", c, val)
	}
}
