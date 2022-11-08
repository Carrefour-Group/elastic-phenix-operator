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
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

const (
	es7Scheme  = "http"
	es7Host    = "elastic7"
	es7Port    = "9200"
	es7Version = 7
)

func buildElasticsearch(t *testing.T) *Elasticsearch7 {
	assert := assert.New(t)

	esConfig, err1 := (&EsConfig{}).FromURI(fmt.Sprintf("%v://%v:%v", es7Scheme, es7Host, es7Port))
	assert.Nil(err1)
	assert.Equal(esConfig, &EsConfig{Scheme: es7Scheme, Host: es7Host, Port: es7Port, Version: es7Version})
	log := zap.New(zap.UseDevMode(true))

	elasticsearch := &Elasticsearch7{}
	err2 := elasticsearch.NewClient(esConfig, log)
	assert.Nil(err2)
	assert.NotNil(elasticsearch)

	return elasticsearch
}

func deleteAll(elasticsearch *Elasticsearch7) {
	ctx := context.Background()
	elasticsearch.DeleteIndex(ctx, "k8s_epo_*")
	elasticsearch.DeleteTemplate(ctx, "k8s_epo_*")
}

func TestEsVersion_ES7(t *testing.T) {
	assert := assert.New(t)

	rawUrl := fmt.Sprintf("%v://%v:%v", es7Scheme, es7Host, es7Port)

	esVersion, err := EsVersion(rawUrl)
	assert.Nil(err)
	assert.Equal(7, esVersion)

	esVersion2, err2 := EsVersion("not-url")
	assert.NotNil(err2)
	assert.Equal(-1, esVersion2)
}

func TestElasticsearch7_PingES(t *testing.T) {
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	err3 := elasticsearch.PingES(context.Background())
	assert.Nil(err3)
}

func TestElasticsearch7_CreateOrUpdateIndex(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_create_index"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}, "mappings":{"properties":{"description":{"type":"keyword"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	assert.True(*elasticsearch.existsIndex(ctx, indexName))
	replicas, shards := elasticsearch.getNumberOfReplicasAndShards(ctx, indexName)
	assert.Equal(int32(3), *replicas)
	assert.Equal(int32(5), *shards)
	properties, err := elasticsearch.getProperties(ctx, indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties, `{"properties" :{"description":{"type":"keyword"}}}`))
}

func TestElasticsearch7_CreateOrUpdateIndex_WithType(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_create_index_with_type"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}, "mappings":{"_doc": {"properties":{"description":{"type":"keyword"}}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	assert.True(*elasticsearch.existsIndex(ctx, indexName))
}

func TestElasticsearch7_CreateOrUpdateIndex_WithoutType(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_create_index_without_type"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}, "mappings":{"properties":{"description":{"type":"keyword"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	assert.True(*elasticsearch.existsIndex(ctx, indexName))
}

func TestElasticsearch7_UpdateIndexMapping(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_update_settings"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"mappings":{"properties":{"description":{"type":"keyword"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	properties, err := elasticsearch.getProperties(ctx, indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties, `{"properties" :{"description":{"type":"keyword"}}}`))

	//add new field in properties
	status2, err := elasticsearch.updateIndexProperties(ctx, indexName, `{"mappings":{"properties":{"description":{"type":"keyword"}, "newField":{"type":"text"}}}}`)
	assert.Nil(err)
	assert.Equal("200", status2.HttpCodeStatus)
	assert.Equal(StatusCreated, status2.Status)
	properties2, err := elasticsearch.getProperties(ctx, indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties2, `{"properties" :{"description":{"type":"keyword"}, "newField":{"type":"text"}}}`))

	//change field type in properties => bad request
	status3, err := elasticsearch.updateIndexProperties(ctx, indexName, `{"mappings":{"properties":{"description":{"type":"text"}, "newField":{"type":"text"}}}}`)
	assert.NotNil(err)
	assert.Equal("400", status3.HttpCodeStatus)
	assert.Equal(StatusError, status3.Status)
	properties3, err := elasticsearch.getProperties(ctx, indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties3, `{"properties" :{"description":{"type":"keyword"}, "newField":{"type":"text"}}}`))

	//update with less properties
	status4, err := elasticsearch.updateIndexProperties(ctx, indexName, `{"mappings":{"properties":{"newField":{"type":"text"}}}}`)
	assert.NotNil(err)
	assert.Equal(StatusError, status4.Status)
	properties4, err := elasticsearch.getProperties(ctx, indexName)
	assert.Nil(err)
	assert.True(CompareJson(*properties4, `{"properties" :{"description":{"type":"keyword"}, "newField":{"type":"text"}}}`))
}

func TestElasticsearch7_UpdateIndexSettings(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_update_settings"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.Equal(StatusCreated, status.Status)
	replicas, shards := elasticsearch.getNumberOfReplicasAndShards(ctx, indexName)
	assert.Equal(int32(3), *replicas)
	assert.Equal(int32(5), *shards)

	//update replicas number
	status2, err := elasticsearch.updateIndexSettings(ctx, indexName, `{"settings":{"number_of_replicas": "1", "number_of_shards": "5"}}`)
	assert.Nil(err)
	assert.Equal("200", status2.HttpCodeStatus)
	assert.Equal(StatusCreated, status2.Status)
	replicas2, shards2 := elasticsearch.getNumberOfReplicasAndShards(ctx, indexName)
	assert.Equal(int32(1), *replicas2)
	assert.Equal(int32(5), *shards2)

	//cannot update shards number
	status3, err := elasticsearch.updateIndexSettings(ctx, indexName, `{"settings":{"number_of_replicas": "1", "number_of_shards": "7"}}`)
	assert.NotNil(err)
	assert.Equal(StatusError, status3.Status)
	replicas3, shards3 := elasticsearch.getNumberOfReplicasAndShards(ctx, indexName)
	assert.Equal(int32(1), *replicas3)
	assert.Equal(int32(5), *shards3)
}

func TestElasticsearch7_UpdateIndexReplicas(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_update_index_setting"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"settings":{"number_of_replicas": "3", "number_of_shards": "5"}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsIndex(ctx, indexName))
	replicas, shards := elasticsearch.getNumberOfReplicasAndShards(ctx, indexName)
	assert.Equal(int32(3), *replicas)
	assert.Equal(int32(5), *shards)

	status2, _, err2 := elasticsearch.updateIndexReplicas(ctx, indexName, 4)
	assert.Nil(err2)
	assert.Equal(200, status2)
	replicas2, shards2 := elasticsearch.getNumberOfReplicasAndShards(ctx, indexName)
	assert.Equal(int32(4), *replicas2)
	assert.Equal(int32(5), *shards2)
}

func TestElasticsearch7_DeleteIndex(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	indexName := "k8s_epo_test_delete_index"

	status, err := elasticsearch.CreateOrUpdateIndex(ctx, indexName, `{"settings":{"number_of_replicas": "3"}}`)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsIndex(ctx, indexName))

	err2 := elasticsearch.DeleteIndex(ctx, indexName)
	assert.Nil(err2)
	assert.False(*elasticsearch.existsIndex(ctx, indexName))
}

func TestElasticsearch7_CreateOrUpdateTemplate(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	templateName := "k8s_epo_test_create_template"

	status, err := elasticsearch.CreateOrUpdateTemplate(ctx, templateName, `{"index_patterns": ["k8s_epo_test_*"], "mappings": {"properties": {}}}`, nil)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsTemplate(ctx, templateName))
}

func TestElasticsearch7_CreateOrUpdateTemplate_WithType(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	templateName := "k8s_epo_test_create_template_with_type"

	status, err := elasticsearch.CreateOrUpdateTemplate(ctx, templateName, `{"index_patterns": ["k8s_epo_test_*"], "mappings": {"my-type": {"properties": {}}}}`, nil)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsTemplate(ctx, templateName))
}

func TestElasticsearch7_DeleteTemplate(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	elasticsearch := buildElasticsearch(t)
	defer deleteAll(elasticsearch)

	templateName := "k8s_epo_test_delete_template"

	status, err := elasticsearch.CreateOrUpdateTemplate(ctx, templateName, `{"index_patterns": ["k8s_epo_test_*"]}`, nil)
	assert.Nil(err)
	assert.Equal("200", status.HttpCodeStatus)
	assert.True(*elasticsearch.existsTemplate(ctx, templateName))

	err2 := elasticsearch.DeleteTemplate(ctx, templateName)
	assert.Nil(err2)
	assert.False(*elasticsearch.existsTemplate(ctx, templateName))
}
