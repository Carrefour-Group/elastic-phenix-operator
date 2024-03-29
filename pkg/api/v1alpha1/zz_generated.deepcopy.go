//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticIndex) DeepCopyInto(out *ElasticIndex) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticIndex.
func (in *ElasticIndex) DeepCopy() *ElasticIndex {
	if in == nil {
		return nil
	}
	out := new(ElasticIndex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticIndex) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticIndexList) DeepCopyInto(out *ElasticIndexList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticIndex, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticIndexList.
func (in *ElasticIndexList) DeepCopy() *ElasticIndexList {
	if in == nil {
		return nil
	}
	out := new(ElasticIndexList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticIndexList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticIndexSpec) DeepCopyInto(out *ElasticIndexSpec) {
	*out = *in
	if in.IndexName != nil {
		in, out := &in.IndexName, &out.IndexName
		*out = new(string)
		**out = **in
	}
	in.ElasticURI.DeepCopyInto(&out.ElasticURI)
	if in.NumberOfShards != nil {
		in, out := &in.NumberOfShards, &out.NumberOfShards
		*out = new(int32)
		**out = **in
	}
	if in.NumberOfReplicas != nil {
		in, out := &in.NumberOfReplicas, &out.NumberOfReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Model != nil {
		in, out := &in.Model, &out.Model
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticIndexSpec.
func (in *ElasticIndexSpec) DeepCopy() *ElasticIndexSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticIndexSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticIndexStatus) DeepCopyInto(out *ElasticIndexStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticIndexStatus.
func (in *ElasticIndexStatus) DeepCopy() *ElasticIndexStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticIndexStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticTemplate) DeepCopyInto(out *ElasticTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticTemplate.
func (in *ElasticTemplate) DeepCopy() *ElasticTemplate {
	if in == nil {
		return nil
	}
	out := new(ElasticTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticTemplateList) DeepCopyInto(out *ElasticTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticTemplateList.
func (in *ElasticTemplateList) DeepCopy() *ElasticTemplateList {
	if in == nil {
		return nil
	}
	out := new(ElasticTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticTemplateSpec) DeepCopyInto(out *ElasticTemplateSpec) {
	*out = *in
	if in.TemplateName != nil {
		in, out := &in.TemplateName, &out.TemplateName
		*out = new(string)
		**out = **in
	}
	in.ElasticURI.DeepCopyInto(&out.ElasticURI)
	if in.NumberOfShards != nil {
		in, out := &in.NumberOfShards, &out.NumberOfShards
		*out = new(int32)
		**out = **in
	}
	if in.NumberOfReplicas != nil {
		in, out := &in.NumberOfReplicas, &out.NumberOfReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Order != nil {
		in, out := &in.Order, &out.Order
		*out = new(int)
		**out = **in
	}
	if in.Model != nil {
		in, out := &in.Model, &out.Model
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticTemplateSpec.
func (in *ElasticTemplateSpec) DeepCopy() *ElasticTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticTemplateStatus) DeepCopyInto(out *ElasticTemplateStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticTemplateStatus.
func (in *ElasticTemplateStatus) DeepCopy() *ElasticTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticURISource) DeepCopyInto(out *ElasticURISource) {
	*out = *in
	if in.SecretKeyRef != nil {
		in, out := &in.SecretKeyRef, &out.SecretKeyRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticURISource.
func (in *ElasticURISource) DeepCopy() *ElasticURISource {
	if in == nil {
		return nil
	}
	out := new(ElasticURISource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EsObjectInfo) DeepCopyInto(out *EsObjectInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EsObjectInfo.
func (in *EsObjectInfo) DeepCopy() *EsObjectInfo {
	if in == nil {
		return nil
	}
	out := new(EsObjectInfo)
	in.DeepCopyInto(out)
	return out
}
