/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SpotPoolSpec defines the desired state of SpotPool.
type SpotPoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SpotPool. Edit spotpool_types.go to remove/update
	SecretId           string   `json:"secretId,omitempty"`
	SecretKey          string   `json:"secretKey,omitempty"`
	Region             string   `json:"region,omitempty"`
	AvailabilityZone   string   `json:"availabilityZone,omitempty"`
	InstanceType       string   `json:"instanceType,omitempty"`
	Minimum            int32    `json:"minimum,omitempty"`
	Maximum            int32    `json:"maximum,omitempty"`
	SubnetId           string   `json:"subnetId,omitempty"`
	VpcId              string   `json:"vpcId,omitempty"`
	SecurityGroupIds   []string `json:"securityGroupIds,omitempty"`
	ImageId            string   `json:"imageId,omitempty"`
	InstanceChargeType string   `json:"instanceChargeType,omitempty"`
	KongGatewayIP      string   `json:"kongGatewayIP,omitempty"`
}

// SpotPoolStatus defines the observed state of SpotPool.
type SpotPoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Size       int32              `json:"size,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	Instances  []Instances        `json:"instances,omitempty"`
}

type Instances struct {
	InstanceId string `json:"instanceId,omitempty"`
	PublicIp   string `json:"publicIp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SpotPool is the Schema for the spotpools API.
type SpotPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpotPoolSpec   `json:"spec,omitempty"`
	Status SpotPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpotPoolList contains a list of SpotPool.
type SpotPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpotPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpotPool{}, &SpotPoolList{})
}
