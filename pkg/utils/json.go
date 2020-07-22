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
	"bytes"
	"encoding/json"
	"reflect"
)

func CompactJson(data string) (string, error) {
	compactedBuffer := new(bytes.Buffer)
	if err := json.Compact(compactedBuffer, []byte(data)); err != nil {
		return "", err
	}
	return compactedBuffer.String(), nil
}

func CompareJson(data1 string, data2 string) bool {
	var result1 interface{}
	if err := json.Unmarshal([]byte(data1), &result1); err != nil {
		return false
	}

	var result2 interface{}
	if err := json.Unmarshal([]byte(data2), &result2); err != nil {
		return false
	}

	return reflect.DeepEqual(result1, result2)
}
