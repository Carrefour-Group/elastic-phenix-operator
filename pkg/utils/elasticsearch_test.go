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
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

const (
	esScheme = "http"
	esHost   = "elastic"
	esPort   = "9200"
)

func buildElasticsearch(t *testing.T) *Elasticsearch {
	assert := assert.New(t)

	esConfig, err1 := (&EsConfig{}).FromURI(fmt.Sprintf("%v://%v:%v", esScheme, esHost, esPort))
	assert.Nil(err1)
	assert.Equal(esConfig, &EsConfig{Scheme: esScheme, Host: esHost, Port: esPort})
	log := zap.New(zap.UseDevMode(true))

	elasticsearch, err2 := (&Elasticsearch{}).NewClient(esConfig, log)
	assert.Nil(err2)
	assert.NotNil(elasticsearch)

	return elasticsearch
}

func deleteAll(elasticsearch *Elasticsearch) {
	elasticsearch.DeleteIndex("k8s_epo_*")
	elasticsearch.DeleteTemplate("k8s_epo_*")
}

func TestEsConfig_String(t *testing.T) {
	esConfig := EsConfig{Scheme: "https", Host: "myhost", Port: "9201", Username: "me", Password: "you"}
	got := esConfig.String()
	expect := "https://myhost:9201"
	if got != expect {
		t.Errorf("EsConfig String %v does not match expect %v", got, expect)
	}
}

func TestEsConfig_FromURI(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		uri            string
		expectEsConfig *EsConfig
		error          bool
	}{
		{uri: "", error: true},
		{uri: "test", error: true},
		{uri: "myhost:9200", error: true},
		{uri: "http://myhost:9200", expectEsConfig: &EsConfig{Scheme: "http", Host: "myhost", Port: "9200"}, error: false},
		{uri: "http://myusername:mypass@myhost:9200", expectEsConfig: &EsConfig{Scheme: "http", Host: "myhost",
			Port: "9200", Username: "myusername", Password: "mypass"}, error: false},
	}

	for _, s := range scenarios {
		got, err := (&EsConfig{}).FromURI(s.uri)

		if s.error {
			assert.NotNil(err)
			assert.Nil(got)
		} else {
			assert.Equal(s.expectEsConfig, got, fmt.Sprintf("expected EsConfig %v does not equal to got EsConfig %v", s.expectEsConfig, got))
		}
	}
}

func TestBuildEsStatus(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		statusCode     int
		message        string
		expectEsStatus EsStatus
	}{
		{statusCode: 200, message: "OK", expectEsStatus: EsStatus{Status: StatusCreated, HttpCodeStatus: "200", Message: "OK"}},
		{statusCode: 404, message: "KO", expectEsStatus: EsStatus{Status: StatusError, HttpCodeStatus: "404", Message: "KO"}},
		{statusCode: 500, message: "KO", expectEsStatus: EsStatus{Status: StatusRetry, HttpCodeStatus: "500", Message: "KO"}},
	}

	for _, s := range scenarios {
		got := BuildEsStatus(s.statusCode, s.message)

		assert.NotNil(got)
		assert.Equal(s.expectEsStatus, *got)
	}
}

func TestElasticsearch_PingES(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	err3 := elasticsearch.PingES()
	assert.Nil(err3)
}

func TestElasticsearch_CreateOrUpdateIndex(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_create_index"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}, "mappings":{"properties":{"description":{"type":"keyword"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	assert.True(*elasticsearch.existsIndex(indexName))
	replicas, shards := elasticsearch.getNumberOfReplicasAndShards(indexName)
	assert.Equal(int32(3), *replicas)
	assert.Equal(int32(5), *shards)
	properties, err := elasticsearch.getProperties(indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties, `{"properties" :{"description":{"type":"keyword"}}}`))
}

func TestElasticsearch_CreateOrUpdateIndex_WithType(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_create_index_with_type"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}, "mappings":{"_doc": {"properties":{"description":{"type":"keyword"}}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	assert.True(*elasticsearch.existsIndex(indexName))
}

func TestElasticsearch_CreateOrUpdateIndex_WithoutType(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_create_index_without_type"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}, "mappings":{"properties":{"description":{"type":"keyword"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	assert.True(*elasticsearch.existsIndex(indexName))
}

func TestElasticsearch_UpdateIndexMapping(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_update_settings"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"mappings":{"properties":{"description":{"type":"keyword"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	properties, err := elasticsearch.getProperties(indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties, `{"properties" :{"description":{"type":"keyword"}}}`))

	//add new field in properties
	status2, err := elasticsearch.updateIndexProperties(indexName, `{"mappings":{"properties":{"description":{"type":"keyword"}, "newField":{"type":"text"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status2.HttpCodeStatus)
	assert.Equal(StatusCreated, status2.Status)
	properties2, err := elasticsearch.getProperties(indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties2, `{"properties" :{"description":{"type":"keyword"}, "newField":{"type":"text"}}}`))

	//change field type in properties => bad request
	status3, err := elasticsearch.updateIndexProperties(indexName, `{"mappings":{"properties":{"description":{"type":"text"}, "newField":{"type":"text"}}}}`)
	assert.NotNil(err)
	assert.Equal("400", status3.HttpCodeStatus)
	assert.Equal(StatusError, status3.Status)
	properties3, err := elasticsearch.getProperties(indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties3, `{"properties" :{"description":{"type":"keyword"}, "newField":{"type":"text"}}}`))

	//update with less properties
	status4, err := elasticsearch.updateIndexProperties(indexName, `{"mappings":{"properties":{"newField":{"type":"text"}}}}`)
	assert.NotNil(err)
	assert.Equal(StatusError, status4.Status)
	properties4, err := elasticsearch.getProperties(indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties4, `{"properties" :{"description":{"type":"keyword"}, "newField":{"type":"text"}}}`))
}

func TestElasticsearch_UpdateIndexSettings(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_update_settings"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	replicas, shards := elasticsearch.getNumberOfReplicasAndShards(indexName)
	assert.Equal(int32(3), *replicas)
	assert.Equal(int32(5), *shards)

	//update replicas number
	status2, err := elasticsearch.updateIndexSettings(indexName, `{"settings":{"number_of_replicas": "1", "number_of_shards": "5"}}`)
	assert.Nil(err)
	assert.Equal("200", status2.HttpCodeStatus)
	assert.Equal(StatusCreated, status2.Status)
	replicas2, shards2 := elasticsearch.getNumberOfReplicasAndShards(indexName)
	assert.Equal(int32(1), *replicas2)
	assert.Equal(int32(5), *shards2)

	//cannot update shards number
	status3, err := elasticsearch.updateIndexSettings(indexName, `{"settings":{"number_of_replicas": "1", "number_of_shards": "7"}}`)
	assert.NotNil(err)
	assert.Equal(StatusError, status3.Status)
	replicas3, shards3 := elasticsearch.getNumberOfReplicasAndShards(indexName)
	assert.Equal(int32(1), *replicas3)
	assert.Equal(int32(5), *shards3)
}

func TestElasticsearch_UpdateIndexReplicas(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_update_index_setting"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsIndex(indexName))
	replicas, shards := elasticsearch.getNumberOfReplicasAndShards(indexName)
	assert.Equal(int32(3), *replicas)
	assert.Equal(int32(5), *shards)

	status2, _, err2 := elasticsearch.updateIndexReplicas(indexName, 4)
	assert.Nil(err2)
	assert.Equal(200, status2)
	replicas2, shards2 := elasticsearch.getNumberOfReplicasAndShards(indexName)
	assert.Equal(int32(4), *replicas2)
	assert.Equal(int32(5), *shards2)
}

func TestElasticsearch_DeleteIndex(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_delete_index"

	status, err := elasticsearch.CreateOrUpdateIndex(indexName, `{"settings":{"number_of_replicas": "3"}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsIndex(indexName))

	err2 := elasticsearch.DeleteIndex(indexName)
	assert.Nil(err2)
	assert.False(*elasticsearch.existsIndex(indexName))
}

func TestElasticsearch_CreateOrUpdateTemplate(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	templateName := "k8s_epo_test_create_template"

	status, err := elasticsearch.CreateOrUpdateTemplate(templateName, `{"index_patterns": ["k8s_epo_test_*"]}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsTemplate(templateName))
}

func TestElasticsearch_DeleteTemplate(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	templateName := "k8s_epo_test_delete_template"

	status, err := elasticsearch.CreateOrUpdateTemplate(templateName, `{"index_patterns": ["k8s_epo_test_*"]}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsTemplate(templateName))

	err2 := elasticsearch.DeleteTemplate(templateName)
	assert.Nil(err2)
	assert.False(*elasticsearch.existsTemplate(templateName))
}
