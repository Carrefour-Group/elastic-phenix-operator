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

func TestCompactJson(t *testing.T) {
	assert := assert.New(t)
	scenarios := []struct {
		json      string
		error     bool
		compacted string
	}{
		{json: `{"name": "test", "type": "unit test"}`, error: false, compacted: `{"name":"test","type":"unit test"}`},
		{json: `{"name": "test, "`, error: true},
	}

	for _, s := range scenarios {
		got, err := CompactJson(s.json)
		if s.error {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.Equal(got, s.compacted)
		}
	}
}

func TestCompareJson(t *testing.T) {
	scenarios := []struct {
		json1 string
		json2 string
		equal bool
	}{
		{json1: `{"name": "test", "type": "unit test"}`, json2: `{"type":"unit test","name":"test"}`, equal: true},
		{json1: `{"name": "test, "`, json2: `{"name": "test, "`, equal: false},
		{json1: `{}`, json2: ``, equal: false},
	}

	for _, s := range scenarios {
		got := CompareJson(s.json1, s.json2)
		var doesEqual string
		if s.equal {
			doesEqual = "does not equal"
		} else {
			doesEqual = "equals"
		}
		assert.Equal(t, got, s.equal, fmt.Sprintf("json '%v' %v json '%v'", s.json1, doesEqual, s.json2))
	}
}
