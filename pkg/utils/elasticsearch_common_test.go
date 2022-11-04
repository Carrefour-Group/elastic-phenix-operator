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
