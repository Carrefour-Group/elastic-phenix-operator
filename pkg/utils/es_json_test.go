/*
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

package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsValidUpdateProperties(t *testing.T) {
	scenarios := []struct {
		oldProperties string
		newProperties string
		validUpdate   bool
	}{
		{oldProperties: `{"properties":{}}`, newProperties: `{"properties":{}}`, validUpdate: true},
		{oldProperties: `{"properties":{}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"}}}`, validUpdate: true},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"}}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"}, "cityCode":{"type":"keyword"}}}`, validUpdate: true},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"}}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"}}}`, validUpdate: true},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"}}}`, newProperties: `{"properties":{"cityCode":{"type":"keyword"}}}`, validUpdate: false},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line1":{"type":"keyword"}}}}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"}}}`, validUpdate: false},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line1":{"type":"keyword"}}}}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line1":{"type":"keyword"},"line2":{"type":"keyword"}}}}}`, validUpdate: true},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line1":{"type":"keyword"}}}}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line2":{"type":"keyword"}}}}}`, validUpdate: false},
		{oldProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line1":{"type":"keyword"},"line2":{"type":"keyword"}}}}}`, newProperties: `{"properties":{"cityName":{"type":"keyword"},"cityAddress":{"type":"object","properties":{"line1":{"type":"keyword"}}}}}`, validUpdate: false},
		{oldProperties: `{"properties":{"cityAddress":{"type":"object","properties":{"line1":{"type":"object","properties":{"road":{"type":"keyword"}}}}}}}`, newProperties: `{"properties":{"cityAddress":{"type":"object","properties":{"line1":{"type":"object","properties":{}}}}}}`, validUpdate: false},
		{oldProperties: `{"properties":{"cityAddress":{"type":"object","properties":{"line1":{"type":"object","properties":{"road":{"type":"keyword"}}}}}}}`, newProperties: `{"properties":{"cityAddress":{"type":"object","properties":{"line1":{"type":"object","properties":{"number":{"type":"keyword"}}}}}}}`, validUpdate: false},
		{oldProperties: `{"properties":{"cityAddress":{"type":"object","properties":{"line1":{"type":"object","properties":{"road":{"type":"keyword"}}}}}}}`, newProperties: `{"properties":{"cityAddress":{"type":"object","properties":{"line1":{"type":"object","properties":{"road":{"type":"keyword"},"number":{"type":"keyword"}}}}}}}`, validUpdate: true},
		{oldProperties: `{"propertie`, newProperties: `{"propertie`, validUpdate: false},
	}

	for _, s := range scenarios {
		got := IsValidUpdateProperties(s.oldProperties, s.newProperties)
		assert.Equal(t, s.validUpdate, got, fmt.Sprintf("oldProperties: %v, newProperties: %v", s.oldProperties, s.newProperties))
	}
}

func TestGetElasticsearchVersion(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		jsonBody  string
		esVersion int
		error     bool
	}{
		{jsonBody: `{"name":"adf8","cluster_name":"es","cluster_uuid":"wLyZGB","version":{"number":"8.4.2"}}`, esVersion: 8, error: false},
		{jsonBody: `{"name":"adf8","cluster_name":"es","cluster_uuid":"wLyZGB","version":{"number":"7.9.2"}}`, esVersion: 7, error: false},
		{jsonBody: `{"name":"adf8","cluster_name":"es","cluster_uuid":"wLyZGB","version":{"number":"6.7.1"}}`, esVersion: 6, error: false},
		{jsonBody: `{"name":"adf8","cluster_name":"es","cluster_uuid":"wLyZGB","version":{"number":"5.4.1"}}`, esVersion: -1, error: true},
		{jsonBody: `{"name":"adf8","cluster_name":"es","cluster_uuid":"wLyZGB"}`, esVersion: -1, error: true},
		{jsonBody: `{"name":"adf8","clust`, esVersion: -1, error: true},
	}

	for _, s := range scenarios {
		esVersion, err := GetElasticsearchVersion(s.jsonBody)
		if s.error {
			assert.NotNil(err)
			assert.Equal(-1, esVersion)
		} else {
			assert.Nil(err)
			assert.Equal(s.esVersion, esVersion)
		}
	}
}

func TestEsModel_AddSettings(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		model  string
		output string
		error  bool
	}{
		{model: `{}`, output: `{"settings":{"number_of_replicas":2,"number_of_shards":3}}`, error: false},
		{model: `{"settings":{"number_of_replicas":4,"number_of_shards":1}}`, output: `{"settings":{"number_of_replicas":2,"number_of_shards":3}}`, error: false},
		{model: `{"settings":{"override":true}}`, output: `{"settings":{"override":true,"number_of_replicas":2,"number_of_shards":3}}`, error: false},
		{model: `{"settings":`, error: true},
	}

	for _, s := range scenarios {
		got, err := (&EsModel{Model: s.model}).AddSettings(2, 3)
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.JSONEq(got, s.output)
		}
	}
}

func TestEsModel_GetNumberOfShards(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		model        string
		expectShards int32
		error        bool
	}{
		{model: `{}`, error: true},
		{model: `{"settings":{"number_of_shards": 3}}`, expectShards: int32(3), error: false},
		{model: `{"settings":{"number_of_shards": "3"}}`, expectShards: int32(3), error: false},
		{model: `{"settings":{}}`, error: true},
		{model: `{"settings":{"num`, error: true},
	}

	for _, s := range scenarios {
		got, err := (&EsModel{Model: s.model}).GetNumberOfShards()
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.Equal(*got, s.expectShards)
		}
	}
}

func TestEsModel_GetNumberOfReplicas(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		model          string
		expectReplicas int32
		error          bool
	}{
		{model: `{}`, error: true},
		{model: `{"settings":{"number_of_replicas": 3}}`, expectReplicas: int32(3), error: false},
		{model: `{"settings":{"number_of_replicas": "3"}}`, expectReplicas: int32(3), error: false},
		{model: `{"settings":{}}`, error: true},
		{model: `{"settings":{"num`, error: true},
	}

	for _, s := range scenarios {
		got, err := (&EsModel{Model: s.model}).GetNumberOfReplicas()
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.Equal(*got, s.expectReplicas)
		}
	}
}

func TestEsModel_GetProperties(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		model            string
		expectProperties string
	}{
		{model: `{}`, expectProperties: ""},
		{model: `{"mappings":{"dynamic": false,"properties":{"cityCode":{"type":"keyword"}}}}`, expectProperties: `{"properties":{"cityCode":{"type":"keyword"}}}`},
		{model: `{"mappings":{"_doc": {"dynamic": false,"properties":{"cityCode":{"type":"keyword"}}}}}`, expectProperties: `{"properties":{"cityCode":{"type":"keyword"}}}`},
		{model: `{"mappings":{"properties": {}}}`, expectProperties: `{"properties":{}}`},
		{model: `{"mappings":{}}`, expectProperties: ""},
		{model: `{"mappi`, expectProperties: ""},
	}

	for _, s := range scenarios {
		got := (&EsModel{Model: s.model}).GetProperties()
		if s.expectProperties == "" {
			assert.Nil(got)
		} else {
			assert.NotNil(got)
			assert.True(CompareJson(s.expectProperties, *got))
		}
	}
}

func TestEsModel_IsMappingWithType(t *testing.T) {
	yes := true
	no := false
	scenarios := []struct {
		model      string
		isWithType *bool
	}{
		{model: `{}`, isWithType: nil},
		{model: `{"mappings":{}}`, isWithType: nil},
		{model: `{"settings":{"num`, isWithType: nil},
		{model: `{"mappings":{"_doc": {"properties":{"description":{"type":"keyword"}}}}}`, isWithType: &yes},
		{model: `{"mappings":{"properties":{"description":{"type":"keyword"}}}}`, isWithType: &no},
	}

	for _, s := range scenarios {
		got := (&EsModel{Model: s.model}).IsMappingWithType()
		assert.Equal(t, s.isWithType, got)
	}
}

func TestEsModel_IsValid(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		esType string
		model  string
		valid  bool
		error  bool
	}{
		{esType: "Index", model: `{}`, valid: true, error: false},
		{esType: "Index", model: `{"aliases":{}, "mappings":{}, "settings":{}}`, valid: true, error: false},
		{esType: "Index", model: `{"mappings":{}}`, valid: true, error: false},
		{esType: "Index", model: `{"settings":{}}`, valid: true, error: false},
		{esType: "Index", model: `{"aliases":{}}`, valid: true, error: false},
		{esType: "Index", model: `{"alia`, error: true},
		{esType: "Template", model: `{}`, error: true},
		{esType: "Template", model: `{"aliases":{}, "mappings":{}, "settings":{},"index_patterns":{}}`, valid: true, error: false},
		{esType: "Template", model: `{"index_patterns":{}}`, valid: true, error: false},
		{esType: "Template", model: `{"mappings":{}}`, valid: false, error: true},
		{esType: "Template", model: `{"mappings":{},"index_patterns":{}}`, valid: true, error: false},
		{esType: "Template", model: `{"settings":{}}`, valid: false, error: true},
		{esType: "Template", model: `{"settings":{},"index_patterns":{}}`, valid: true, error: false},
		{esType: "Template", model: `{"aliases":{}}`, valid: false, error: true},
		{esType: "Template", model: `{"aliases":{},"index_patterns":{}}`, valid: true, error: false},
		{esType: "Template", model: `{"alia`, valid: false, error: true},
	}

	for _, s := range scenarios {
		got, err := (&EsModel{Model: s.model}).IsValid(s.esType)
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			var isValid string
			if s.valid {
				isValid = "should be"
			} else {
				isValid = "should not be"
			}
			assert.Equal(got, s.valid, fmt.Sprintf("model '%v' of type '%v' %v valid", s.model, s.esType, isValid))
		}
	}
}

func TestEsSettings_GetNumberOfShards(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		settings     string
		index        string
		expectShards int32
		error        bool
	}{
		{settings: `{}`, error: true},
		{settings: `{"product":{"settings":{"index": {"number_of_shards": "3"}}}}`, index: "product", expectShards: int32(3), error: false},
		{settings: `{"product":{"settings":{"index": {}}}}`, index: "product", error: true},
		{settings: `{"product":{"settings":{"index": {"number_of_shards": "3"}}}}`, index: "store", error: true},
		{settings: `{"product":{"settings":{"in`, error: true},
	}

	for _, s := range scenarios {
		got, err := (&EsSettings{Settings: s.settings}).GetNumberOfShards(s.index)
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.Equal(*got, s.expectShards)
		}
	}
}

func TestEsSettings_GetNumberOfReplicas(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		settings       string
		index          string
		expectReplicas int32
		error          bool
	}{
		{settings: `{}`, error: true},
		{settings: `{"product":{"settings":{"index": {"number_of_replicas": "3"}}}}`, index: "product", expectReplicas: int32(3), error: false},
		{settings: `{"product":{"settings":{"index": {}}}}`, index: "product", error: true},
		{settings: `{"product":{"settings":{"index": {"number_of_replicas": "3"}}}}`, index: "store", error: true},
		{settings: `{"product":{"settings":{"in`, error: true},
	}

	for _, s := range scenarios {
		got, err := (&EsSettings{Settings: s.settings}).GetNumberOfReplicas(s.index)
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.Equal(*got, s.expectReplicas)
		}
	}
}

func TestEsMappings_GetProperties(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		mappings         string
		index            string
		expectProperties string
	}{
		{mappings: `{}`, expectProperties: ""},
		{mappings: `{"city":{"mappings":{"properties":{"cityCode":{"type":"keyword"}}}}}`, index: "city", expectProperties: `{"properties":{"cityCode":{"type":"keyword"}}}`},
		{mappings: `{"city":{"mappings":{"properties": {}}}}`, index: "city", expectProperties: `{"properties":{}}`},
		{mappings: `{"city":{"mappings":{"properties": {}}}}`, index: "product", expectProperties: ""},
		{mappings: `{"city":{"mappings":{}}}`, index: "city", expectProperties: ""},
		{mappings: `{"city":{"mappi`, index: "city", expectProperties: ""},
	}

	for _, s := range scenarios {
		got := (&EsMappings{Mappings: s.mappings}).GetProperties(s.index)
		if s.expectProperties == "" {
			assert.Nil(got)
		} else {
			assert.NotNil(got)
			assert.True(CompareJson(s.expectProperties, *got))
		}
	}
}
