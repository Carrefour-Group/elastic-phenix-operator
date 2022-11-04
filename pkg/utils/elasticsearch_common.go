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
	"github.com/go-logr/logr"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	Version  int
}

func (conf EsConfig) String() string {
	return fmt.Sprintf("%v://%v:%v", conf.Scheme, conf.Host, conf.Port)
}

func (conf *EsConfig) FromURI(rawurl string) (*EsConfig, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	var esConfig *EsConfig
	if u.User != nil {
		password, _ := u.User.Password()
		esConfig = &EsConfig{Scheme: u.Scheme, Host: u.Hostname(), Port: u.Port(), Username: u.User.Username(), Password: password}
	} else {
		esConfig = &EsConfig{Scheme: u.Scheme, Host: u.Hostname(), Port: u.Port()}
	}

	if esConfig.Host == "" {
		return nil, fmt.Errorf("Cannot retrieve host from rawurl %v", rawurl)
	}

	esVersion, err := EsVersion(rawurl)
	if err != nil {
		return esConfig, err
	}
	esConfig.Version = esVersion

	return esConfig, nil
}

type EsStatus struct {
	Status         string
	HttpCodeStatus string
	Message        string
}

func EsVersion(rawurl string) (int, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return -1, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	return GetElasticsearchVersion(string(body))
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

type Elasticsearch interface {
	NewClient(config *EsConfig, log logr.Logger) error
	PingES(ctx context.Context) error
	CreateOrUpdateIndex(ctx context.Context, indexName string, model string) (*EsStatus, error)
	DeleteIndex(ctx context.Context, indexName string) error
	CreateOrUpdateTemplate(ctx context.Context, templateName string, model string, order *int) (*EsStatus, error)
	DeleteTemplate(ctx context.Context, templateName string) error
}
