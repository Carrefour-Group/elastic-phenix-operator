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

func TestIs2xxStatusCode(t *testing.T) {
	scenarios := []struct {
		statusCode int
		is2xx      bool
	}{
		{statusCode: 102, is2xx: false},
		{statusCode: 200, is2xx: true},
		{statusCode: 401, is2xx: false},
		{statusCode: 503, is2xx: false},
	}

	for _, s := range scenarios {
		got := is2xxStatusCode(s.statusCode)
		assert.Equal(t, got, s.is2xx, fmt.Sprintf("StatusCode %v does %vmatch %s ?", s.statusCode, maybeNot(s.is2xx), Http2xx))
	}
}

func TestIs4xxStatusCode(t *testing.T) {
	scenarios := []struct {
		statusCode int
		is4xx      bool
	}{
		{statusCode: 102, is4xx: false},
		{statusCode: 200, is4xx: false},
		{statusCode: 401, is4xx: true},
		{statusCode: 503, is4xx: false},
	}

	for _, s := range scenarios {
		got := is4xxStatusCode(s.statusCode)
		assert.Equal(t, got, s.is4xx, fmt.Sprintf("StatusCode %v does %vmatch %s !", s.statusCode, maybeNot(s.is4xx), Http4xx))
	}
}

func TestIs5xxStatusCode(t *testing.T) {
	scenarios := []struct {
		statusCode int
		is5xx      bool
	}{
		{statusCode: 102, is5xx: false},
		{statusCode: 200, is5xx: false},
		{statusCode: 401, is5xx: false},
		{statusCode: 503, is5xx: true},
	}

	for _, s := range scenarios {
		got := is5xxStatusCode(s.statusCode)
		assert.Equal(t, got, s.is5xx, fmt.Sprintf("StatusCode %v does %vmatch %s !", s.statusCode, maybeNot(s.is5xx), Http5xx))
	}
}

func TestIsXxxStatusCode(t *testing.T) {
	scenarios := []struct {
		statusCode         int
		httpCodeStatusType HttpCodeStatus
		valid              bool
	}{
		{statusCode: 102, httpCodeStatusType: Http1xx, valid: true},
		{statusCode: 100, httpCodeStatusType: Http5xx, valid: false},
		{statusCode: 200, httpCodeStatusType: Http4xx, valid: false},
		{statusCode: 202, httpCodeStatusType: Http2xx, valid: true},
		{statusCode: 401, httpCodeStatusType: Http4xx, valid: true},
		{statusCode: 503, httpCodeStatusType: Http4xx, valid: false},
	}

	for _, s := range scenarios {
		got := isXxxStatusCode(s.statusCode, s.httpCodeStatusType)
		assert.Equal(t, got, s.valid, fmt.Sprintf("StatusCode %v does %vmatch %s !", s.statusCode, maybeNot(s.valid), s.httpCodeStatusType))
	}
}

func maybeNot(b bool) string {
	if b {
		return "not "
	}
	return ""
}
