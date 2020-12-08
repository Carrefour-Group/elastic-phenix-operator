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
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-logr/logr"
)

const (
	StatusCreated                      = "Created"
	StatusRetry                        = "Retry"
	StatusError                        = "Error"
	Index                              = "Index"
	Template                           = "Template"
	ElasticMainFnTimeout time.Duration = 10 * time.Second
)

type EsConfig struct {
	Scheme   string
	Host     string
	Port     string
	Username string
	Password string
}

func (conf EsConfig) String() string {
	return fmt.Sprintf("%v://%v:%v", conf.Scheme, conf.Host, conf.Port)
}

func (conf *EsConfig) FromURI(rawurl string) (*EsConfig, error) {
	if u, err := url.Parse(rawurl); err != nil {
		return nil, err
	} else {
		var esConfig *EsConfig
		if u.User != nil {
			password, _ := u.User.Password()
			esConfig = &EsConfig{Scheme: u.Scheme, Host: u.Hostname(), Port: u.Port(), Username: u.User.Username(), Password: password}
		} else {
			esConfig = &EsConfig{Scheme: u.Scheme, Host: u.Hostname(), Port: u.Port()}
		}
		if esConfig.Host == "" {
			return nil, errors.New(fmt.Sprintf("Cannot retrieve host from rawurl %v", rawurl))
		} else {
			return esConfig, nil
		}
	}
}

type EsStatus struct {
	Status         string
	HttpCodeStatus string
	Message        string
}

func BuildEsStatus(statusCode int, message string) *EsStatus {
	var status string
	if is4xxStatusCode(statusCode) {
		status = StatusError
	} else if is2xxStatusCode(statusCode) {
		status = StatusCreated
	} else {
		status = StatusRetry
	}
	return &EsStatus{Status: status, HttpCodeStatus: strconv.Itoa(statusCode), Message: message}
}

type Elasticsearch struct {
	Config *EsConfig
	Client *elasticsearch.Client
	log    logr.Logger
}

func (es Elasticsearch) NewClient(config *EsConfig, log logr.Logger) (*Elasticsearch, error) {
	conf := elasticsearch.Config{Addresses: []string{config.String()}, Username: config.Username, Password: config.Password}
	client, err := elasticsearch.NewClient(conf)
	log.Info("elasticsearch client created successfully", "host", config)
	if err != nil {
		log.Error(err, "error while creating elasticsearch client", "host", config)
		return nil, err
	}

	es.log = log
	es.Config = config
	es.Client = client
	return &es, nil
}

func (es *Elasticsearch) PingES(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ElasticMainFnTimeout)
	defer cancel()

	info, err := esapi.InfoRequest{}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while connecting to elasticsearch", "host", es.Config)
		return err
	}
	defer info.Body.Close()
	es.log.Info("connected successfully to elasticsearch")
	return nil
}

func (es *Elasticsearch) existsIndex(ctx context.Context, indexName string) *bool {
	response, err := esapi.IndicesExistsRequest{Index: []string{indexName}}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while checking index exists", "indexName", indexName)
		return nil
	}
	defer response.Body.Close()
	is2xxStatusCode := is2xxStatusCode(response.StatusCode)
	return &is2xxStatusCode
}

func (es *Elasticsearch) existsTemplate(ctx context.Context, templateName string) *bool {
	response, err := esapi.IndicesExistsTemplateRequest{Name: []string{templateName}}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while checking template exists", "templateName", templateName)
		return nil
	}
	defer response.Body.Close()
	is2xxStatusCode := is2xxStatusCode(response.StatusCode)
	return &is2xxStatusCode
}

func (es *Elasticsearch) getNumberOfReplicasAndShards(ctx context.Context, indexName string) (*int32, *int32) {
	response, err := esapi.IndicesGetSettingsRequest{Index: []string{indexName}}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while getting settings", "indexName", indexName)
		return nil, nil
	}
	defer response.Body.Close()
	settings, err2 := StreamToString(response.Body)
	if err2 != nil {
		es.log.Error(err2, "error while converting stream to string to get settings", "indexName", indexName)
		return nil, nil
	}
	replicas, _ := (&EsSettings{Settings: settings}).GetNumberOfReplicas(indexName)
	shards, _ := (&EsSettings{Settings: settings}).GetNumberOfShards(indexName)
	return replicas, shards
}

func (es *Elasticsearch) getProperties(ctx context.Context, indexName string) (*string, error) {
	no := false
	response, err := esapi.IndicesGetMappingRequest{Index: []string{indexName}, IncludeTypeName: &no}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while getting properties", "indexName", indexName)
		return nil, err
	}
	defer response.Body.Close()
	mappings, err2 := StreamToString(response.Body)
	if err2 != nil {
		es.log.Error(err2, "error while converting stream to string to get properties", "indexName", indexName)
		return nil, err2
	}

	return (&EsMappings{Mappings: mappings}).GetProperties(indexName), nil
}

func (es *Elasticsearch) updateIndexReplicas(ctx context.Context, indexName string, numReplicas int32) (int, string, error) {
	settings := strings.NewReader(fmt.Sprintf(`{"index" : {"number_of_replicas" : %v}}`, numReplicas))
	response, err := esapi.IndicesPutSettingsRequest{Index: []string{indexName}, Body: settings}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while updating number_of_replicas", "indexName", indexName, "replicas", numReplicas)
		return -1, "", err
	}
	defer response.Body.Close()
	statusCode := response.StatusCode
	responseStr := response.String()
	return statusCode, responseStr, nil
}

func (es *Elasticsearch) updateIndexMapping(ctx context.Context, indexName string, properties string) (int, string, error) {
	mappingReader := strings.NewReader(properties)
	no := false
	response, err := esapi.IndicesPutMappingRequest{Index: []string{indexName}, Body: mappingReader, IncludeTypeName: &no}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while updating mapping", "indexName", indexName, "mapping", properties)
		return -1, "", err
	}
	defer response.Body.Close()
	statusCode := response.StatusCode
	responseStr := response.String()
	return statusCode, responseStr, nil
}

func (es *Elasticsearch) CreateOrUpdateIndex(ctx context.Context, indexName string, model string) (*EsStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, ElasticMainFnTimeout)
	defer cancel()

	exists := es.existsIndex(ctx, indexName)
	if exists == nil {
		errMsg := "error while checking index exists"
		es.log.Error(nil, errMsg, "indexName", indexName)
		return &EsStatus{Status: StatusError, Message: errMsg}, errors.New(errMsg)
	} else if *exists {
		es.log.Info("index already exists", "indexName", indexName)

		if status, err := es.updateIndexSettings(ctx, indexName, model); status != nil || err != nil {
			return status, err
		}

		if status, err := es.updateIndexProperties(ctx, indexName, model); status != nil || err != nil {
			return status, err
		}

		return &EsStatus{Status: StatusCreated, HttpCodeStatus: "200"}, nil
	}

	shouldIncludeTypeName := (&EsModel{Model: model}).IsMappingWithType()
	response, err := esapi.IndicesCreateRequest{Index: indexName, Body: strings.NewReader(model), IncludeTypeName: shouldIncludeTypeName}.Do(ctx, es.Client)

	if err != nil {
		es.log.Error(err, "error while creating index", "indexName", indexName)
		return &EsStatus{Status: StatusError, Message: err.Error()}, err
	}

	defer response.Body.Close()

	if !is2xxStatusCode(response.StatusCode) {
		es.log.Error(nil, "error while creating index", "indexName", indexName, "http-response", response)
		status := BuildEsStatus(response.StatusCode, response.String())
		return status, errors.New("error while creating index")
	}

	es.log.Info("index was created successfully", "indexName", indexName)
	return BuildEsStatus(response.StatusCode, response.String()), nil
}

func (es *Elasticsearch) updateIndexSettings(ctx context.Context, indexName string, model string) (*EsStatus, error) {
	oldNumReplicas, oldNumShards := es.getNumberOfReplicasAndShards(ctx, indexName)
	numReplicas, err := (&EsModel{Model: model}).GetNumberOfReplicas()
	numShards, err2 := (&EsModel{Model: model}).GetNumberOfShards()

	if err != nil {
		errMsg := fmt.Sprintf("error while getting number_of_repliacs from model %v", model)
		es.log.Error(err, errMsg, "indexName", indexName)
		return &EsStatus{Status: StatusError, Message: errMsg}, err
	}

	if err2 != nil {
		errMsg := fmt.Sprintf("error while getting number_of_shards from model %v", model)
		es.log.Error(err2, errMsg, "indexName", indexName)
		return &EsStatus{Status: StatusError, Message: errMsg}, err2
	}

	isShardsUpdated := oldNumShards == nil || *oldNumShards != *numShards
	if isShardsUpdated {
		errMsg := fmt.Sprintf("you cannot update number_of_shards from %v to %v on existing index %v", *oldNumShards, *numShards, indexName)
		es.log.Error(nil, errMsg, "indexName", indexName)
		return &EsStatus{Status: StatusError, Message: errMsg}, errors.New(errMsg)
	}

	isReplicasUpdated := (oldNumReplicas == nil || *oldNumReplicas != *numReplicas) && err == nil && err2 == nil
	if isReplicasUpdated {
		es.log.Info("index already exists and updating number_of_replicas", "indexName", indexName, "from", *oldNumReplicas, "to", *numReplicas)
		statusCode, responseStr, err := es.updateIndexReplicas(ctx, indexName, *numReplicas)
		if err != nil {
			errMsg := fmt.Sprintf("error while updating number_of_replicas from %v to %v", *oldNumReplicas, *numReplicas)
			es.log.Error(err, errMsg, "indexName", indexName)
			return &EsStatus{Status: StatusError, Message: errMsg}, err
		} else if !is2xxStatusCode(statusCode) {
			status := BuildEsStatus(statusCode, responseStr)
			errMsg := "error while updating index number_of_replicas"
			es.log.Error(nil, errMsg, "indexName", indexName, "http-response", responseStr)
			return status, errors.New(errMsg)
		} else {
			return BuildEsStatus(statusCode, responseStr), nil
		}
	}

	return nil, nil
}

func (es *Elasticsearch) updateIndexProperties(ctx context.Context, indexName string, model string) (*EsStatus, error) {
	oldProperties, err := es.getProperties(ctx, indexName)
	properties := (&EsModel{Model: model}).GetProperties()

	if err != nil {
		errMsg := fmt.Sprintf("error while getting old properties from index %v", indexName)
		es.log.Error(err, errMsg, "indexName", indexName)
		return &EsStatus{Status: StatusError, Message: errMsg}, err
	}

	arePropertiesUpdates := oldProperties != nil && properties != nil && !CompareJson(*oldProperties, *properties)
	if arePropertiesUpdates {
		es.log.Info("index already exists and updating properties", "indexName", indexName, "from", *oldProperties, "to", *properties)

		isValid := IsValidUpdateProperties(*oldProperties, *properties)
		if !isValid {
			errMsg := fmt.Sprintf("you cannot delete properties, error while updating properties from %v to %v", *oldProperties, *properties)
			es.log.Error(nil, errMsg, "indexName", indexName)
			return &EsStatus{Status: StatusError, Message: errMsg}, errors.New(errMsg)
		}

		statusCode, responseStr, err := es.updateIndexMapping(ctx, indexName, *properties)
		if err != nil {
			errMsg := fmt.Sprintf("error while updating properties from %v to %v", *oldProperties, *properties)
			es.log.Error(err, errMsg, "indexName", indexName)
			return &EsStatus{Status: StatusError, Message: errMsg}, err
		} else if !is2xxStatusCode(statusCode) {
			status := BuildEsStatus(statusCode, responseStr)
			errMsg := "error while updating index properties"
			es.log.Error(nil, errMsg, "indexName", indexName, "http-response", responseStr)
			return status, errors.New(errMsg)
		} else {
			return BuildEsStatus(statusCode, responseStr), nil
		}
	}

	return nil, nil
}

func (es *Elasticsearch) DeleteIndex(ctx context.Context, indexName string) error {
	ctx, cancel := context.WithTimeout(ctx, ElasticMainFnTimeout)
	defer cancel()

	exists := es.existsIndex(ctx, indexName)
	if exists != nil && !*exists {
		es.log.Info("index cannot be deleted because it does not exists", "indexName", indexName)
		return nil
	}

	response, err := esapi.IndicesDeleteRequest{Index: []string{indexName}}.Do(ctx, es.Client)
	if err != nil {
		es.log.Error(err, "error while deleting index", "indexName", indexName)
		return err
	}

	defer response.Body.Close()

	if !is2xxStatusCode(response.StatusCode) {
		es.log.Error(nil, "error while deleting index", "indexName", indexName, "http-response", response)
		return errors.New(fmt.Sprintf("error while deleting index %v: %v", indexName, response))
	}

	es.log.Info("index was deleted successfully", "indexName", indexName)
	return nil
}

func (es *Elasticsearch) CreateOrUpdateTemplate(ctx context.Context, templateName string, model string) (*EsStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, ElasticMainFnTimeout)
	defer cancel()

	exists := es.existsTemplate(ctx, templateName)

	shouldIncludeTypeName := (&EsModel{Model: model}).IsMappingWithType()
	response, err := esapi.IndicesPutTemplateRequest{Name: templateName, Body: strings.NewReader(model), IncludeTypeName: shouldIncludeTypeName}.Do(ctx, es.Client)
	if err != nil || exists == nil {
		es.log.Error(err, "error while creating template", "templateName", templateName)
		return &EsStatus{Status: StatusError, Message: err.Error()}, err
	}

	defer response.Body.Close()

	if !is2xxStatusCode(response.StatusCode) {
		es.log.Error(nil, "error while creating template", "templateName", templateName, "http-response", response)
		status := BuildEsStatus(response.StatusCode, response.String())
		return status, errors.New("error while creating template")
	}

	if *exists {
		es.log.Info("template already exists and was updated successfully", "templateName", templateName)
	} else {
		es.log.Info("template was created successfully", "templateName", templateName)
	}

	return BuildEsStatus(response.StatusCode, response.String()), nil
}

func (es *Elasticsearch) DeleteTemplate(ctx context.Context, templateName string) error {
	ctx, cancel := context.WithTimeout(ctx, ElasticMainFnTimeout)
	defer cancel()

	exists := es.existsTemplate(ctx, templateName)
	if exists != nil && !*exists {
		es.log.Info("template cannot be deleted because it does not exists", "templateName", templateName)
		return nil
	}

	response, err := esapi.IndicesDeleteTemplateRequest{Name: templateName}.Do(ctx, es.Client)

	if err != nil {
		es.log.Error(err, "error while deleting template", "templateName", templateName)
		return err
	}

	defer response.Body.Close()

	if !is2xxStatusCode(response.StatusCode) {
		es.log.Error(nil, "error while deleting template", "templateName", templateName, "http-response", response)
		return errors.New(fmt.Sprintf("error while deleting template %v: %v", templateName, response))
	}

	es.log.Info("template was deleted successfully", "templateName", templateName)
	return nil
}
