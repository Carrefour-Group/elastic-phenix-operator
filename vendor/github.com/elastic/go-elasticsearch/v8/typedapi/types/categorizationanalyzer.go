// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.


// Code generated from the elasticsearch-specification DO NOT EDIT.
// https://github.com/elastic/elasticsearch-specification/tree/4316fc1aa18bb04678b156f23b22c9d3f996f9c9


package types

// CategorizationAnalyzer holds the union for the following types:
//
//	CategorizationAnalyzerDefinition
//	string
//
// https://github.com/elastic/elasticsearch-specification/blob/4316fc1aa18bb04678b156f23b22c9d3f996f9c9/specification/ml/_types/Analysis.ts#L124-L125
type CategorizationAnalyzer interface{}

// CategorizationAnalyzerBuilder holds CategorizationAnalyzer struct and provides a builder API.
type CategorizationAnalyzerBuilder struct {
	v CategorizationAnalyzer
}

// NewCategorizationAnalyzer provides a builder for the CategorizationAnalyzer struct.
func NewCategorizationAnalyzerBuilder() *CategorizationAnalyzerBuilder {
	return &CategorizationAnalyzerBuilder{}
}

// Build finalize the chain and returns the CategorizationAnalyzer struct
func (u *CategorizationAnalyzerBuilder) Build() CategorizationAnalyzer {
	return u.v
}

func (u *CategorizationAnalyzerBuilder) CategorizationAnalyzerDefinition(categorizationanalyzerdefinition *CategorizationAnalyzerDefinitionBuilder) *CategorizationAnalyzerBuilder {
	v := categorizationanalyzerdefinition.Build()
	u.v = &v
	return u
}

func (u *CategorizationAnalyzerBuilder) String(string string) *CategorizationAnalyzerBuilder {
	u.v = &string
	return u
}
