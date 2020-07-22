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
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestContainsString(t *testing.T) {
	slice := []string{"aa", "bb", "cc"}
	scenarios := []struct {
		contains bool
		value    string
	}{
		{contains: true, value: "aa"},
		{contains: true, value: "bb"},
		{contains: true, value: "cc"},
		{contains: false, value: "dd"},
	}

	for _, s := range scenarios {
		got := ContainsString(slice, s.value)
		if got != s.contains {
			if s.contains {
				t.Errorf("slice %v does not contain value %v", slice, s.value)
			} else {
				t.Errorf("slice %v should not contain value %v", slice, s.value)
			}
		}
	}
}

func TestRemoveString(t *testing.T) {
	slice := []string{"aa", "bb", "cc"}
	scenarios := []struct {
		slice  []string
		remove string
		expect []string
	}{
		{remove: "aa", expect: []string{"bb", "cc"}},
		{remove: "bb", expect: []string{"aa", "cc"}},
		{remove: "cc", expect: []string{"aa", "bb"}},
		{remove: "dd", expect: []string{"aa", "bb", "cc"}},
	}

	for _, s := range scenarios {
		got := RemoveString(slice, s.remove)
		if !reflect.DeepEqual(got, s.expect) {
			t.Errorf("expected slice %v does not equal got slice %v after removing value %v from slice %v", s.expect, got, s.remove, slice)
		}
	}
}

func TestStreamToString(t *testing.T) {
	assert := assert.New(t)
	got, err := StreamToString(strings.NewReader("Hello !"))
	assert.Nil(err, "error should be nil")
	assert.Equal("Hello !", got)
}
