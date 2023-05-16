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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	funk "github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

func IsValidUpdateProperties(oldProperties string, newProperties string) bool {
	if gjson.Get(oldProperties, "properties").Exists() && gjson.Get(newProperties, "properties").Exists() {
		oldPropKeys := funk.Keys(gjson.Get(oldProperties, "properties").Map()).([]string)
		newPropKeys := funk.Keys(gjson.Get(newProperties, "properties").Map()).([]string)

		if len(newPropKeys) < len(oldPropKeys) ||
			len(funk.IntersectString(oldPropKeys, newPropKeys)) != len(oldPropKeys) {
			return false
		}

		fieldsWithProperties := getNestedFieldsWithProperties(oldProperties)
		for _, field := range fieldsWithProperties {
			oldSubProperties := getPropertiesFromPath(fmt.Sprintf("properties.%v.properties", field), oldProperties)
			newSubProperties := getPropertiesFromPath(fmt.Sprintf("properties.%v.properties", field), newProperties)
			if !IsValidUpdateProperties(*oldSubProperties, *newSubProperties) {
				return false
			}
		}
		return true
	}

	return false
}

func GetElasticsearchVersion(jsonBody string) (int, error) {
	if maybeValue := gjson.Get(jsonBody, "version.number"); maybeValue.Exists() {
		esVersion, err := strconv.Atoi(maybeValue.String()[0:1])
		if err != nil {
			return -1, fmt.Errorf("cannot retrieve elasticsearch version from this json %v", jsonBody)
		}
		if esVersion > 8 || esVersion < 6 {
			return -1, fmt.Errorf("elasticsearch version %v not supported", maybeValue.String())
		}
		return esVersion, nil
	}
	return -1, fmt.Errorf("cannot retrieve elasticsearch version from this json %v", jsonBody)
}

type EsModel struct {
	Model string
}

func (m *EsModel) AddSettings(replicas int32, shards int32) (string, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(m.Model), &result); err != nil {
		return "", err
	}

	settings, ok := result["settings"]
	if !ok {
		result["settings"] = map[string]interface{}{
			"number_of_replicas": replicas,
			"number_of_shards":   shards,
		}
	} else {
		settingsMap := settings.(map[string]interface{})
		settingsMap["number_of_replicas"] = replicas
		settingsMap["number_of_shards"] = shards
	}

	js, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

func (m *EsModel) GetNumberOfShards() (*int32, error) {
	return getIntFromPath(m.Model, "settings.number_of_shards")
}

func (m *EsModel) GetNumberOfReplicas() (*int32, error) {
	return getIntFromPath(m.Model, "settings.number_of_replicas")
}

func (m *EsModel) GetProperties() *string {
	isMappingWithType := m.IsMappingWithType()
	if isMappingWithType != nil && *isMappingWithType {
		path := "mappings.*.properties"
		return getPropertiesFromPath(path, m.Model)
	}
	path := "mappings.properties"
	return getPropertiesFromPath(path, m.Model)
}

func (m *EsModel) IsMappingWithType() *bool {
	if maybeMappings := gjson.Get(m.Model, "mappings"); maybeMappings.Exists() {
		if mappings := gjson.Get(m.Model, "mappings").Map(); len(mappings) == 0 {
			return nil
		} else if maybeProperties := gjson.Get(m.Model, "mappings.properties"); maybeProperties.Exists() {
			no := false
			return &no
		} else {
			yes := true
			return &yes
		}
	} else {
		return nil
	}
}

func (m *EsModel) IsValid(mType string) (bool, error) {
	var keywords []string
	var requiredField string

	if mType == "Index" {
		keywords = []string{"aliases", "mappings", "settings"}
	} else if mType == "Template" {
		keywords = []string{"aliases", "mappings", "settings", "index_patterns", "version"}
		requiredField = "index_patterns"
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(m.Model), &result); err != nil {
		return false, fmt.Errorf("%v model is not a valid json", mType)
	}

	keys := funk.Keys(result).([]string)

	if len(keys) > len(keywords) {
		return false, fmt.Errorf("%v model should contain at most these fields %v", mType, keywords)
	}

	if requiredField != "" && !funk.ContainsString(keys, requiredField) {
		return false, fmt.Errorf("%v model should contain required field %v", mType, requiredField)
	}

	for _, k := range keys {
		if !funk.ContainsString(keywords, k) {
			return false, fmt.Errorf("%v model should contain only fields from this list %v", mType, keywords)
		}
	}

	return true, nil
}

type EsSettings struct {
	Settings string
}

func (s *EsSettings) GetNumberOfShards(indexName string) (*int32, error) {
	path := fmt.Sprintf("%v.settings.index.number_of_shards", indexName)
	return getIntFromPath(s.Settings, path)
}

func (s *EsSettings) GetNumberOfReplicas(indexName string) (*int32, error) {
	path := fmt.Sprintf("%v.settings.index.number_of_replicas", indexName)
	return getIntFromPath(s.Settings, path)
}

type EsMappings struct {
	Mappings string
}

type EsPipelines struct {
	Pipeline string
}

func (m EsMappings) GetProperties(indexName string) *string {
	path := fmt.Sprintf("%v.mappings.properties", indexName)
	return getPropertiesFromPath(path, m.Mappings)
}

func getIntFromPath(json string, path string) (*int32, error) {
	if maybeValue := gjson.Get(json, path); maybeValue.Exists() {
		valueToReturn, err := strconv.Atoi(maybeValue.String())
		if err != nil {
			return nil, err
		}
		value := int32(valueToReturn)
		return &value, nil
	}
	return nil, fmt.Errorf("int value not found using path %v in json %v", path, json)
}

func getPropertiesFromPath(path string, json string) *string {
	if maybeProperties := gjson.Get(json, path); maybeProperties.Exists() {
		innerProperties := maybeProperties.Raw
		properties := fmt.Sprintf(`{"properties":%v}`, innerProperties)
		return &properties
	}
	return nil
}

func getNestedFieldsWithProperties(properties string) []string {
	var fieldsWithProperties []string
	for field, body := range gjson.Get(properties, "properties").Map() {
		if gjson.Get(body.Raw, "properties").Exists() {
			fieldsWithProperties = append(fieldsWithProperties, field)
		}
	}
	return fieldsWithProperties
}

type void struct{}

var member void

func (p EsPipelines) Validate() error {
	pipeline := p.Pipeline
	pipelineFields := []string{"description", "processors"}
	pipelineFieldsSet := make(map[string]void)
	processorFields := []string{"append", "attachment", "bytes", "circle", "community_id", "convert", "csv", "date", "date_index_name", "dissect", "dot_expander", "drop", "enrich", "fail", "fingerprint", "foreach", "geo_grid", "geoip", "grok", "gsub", "html_strip", "inference", "join", "json", "kv", "lowercase", "network_direction", "pipeline", "redact", "registered_domain", "remove", "rename", "script", "set", "set_security_user", "sort", "split", "trim", "uppercase", "url_decode", "uri_parts", "user_agent"}
	processorFieldsSet := make(map[string]void)
	for _, v := range processorFields {
		processorFieldsSet[v] = member
	}
	for _, v := range pipelineFields {
		pipelineFieldsSet[v] = member
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(pipeline), &result); err != nil {
		return fmt.Errorf("pipeline model is not a valid json")
	}
	keys := funk.Keys(result).([]string)

	if len(keys) > len(pipelineFields) {
		return fmt.Errorf("pipeline model should contain at most these fields %v", pipelineFields)
	}

	for _, v := range pipelineFields {
		_, exists := result[v]
		if !exists {
			return fmt.Errorf("pipeline model should contain all the fields from this list %v", pipelineFields)
		}
	}

	description, ok := result["description"].(string)
	if !ok || strings.TrimSpace(description) == "" {
		return fmt.Errorf("pipeline description field is not a valid string or empty")
	}

	processors, ok := result["processors"].([]interface{})
	if !ok {
		return fmt.Errorf("pipeline processors field is  not a json object")
	}

	for _, processor := range processors {
		processKeys := funk.Keys(processor).([]string)
		for _, processorKey := range processKeys {
			_, exists := processorFieldsSet[processorKey]
			if !exists {
				return fmt.Errorf("pipeline processors field should contain only processors from this list %v", pipelineFields)
			}
		}
	}

	return nil

}
