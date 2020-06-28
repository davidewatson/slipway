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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file

// ImageMirrorSpec defines the desired state of ImageMirror
type ImageMirrorSpec struct {
	// SourceRepository is the url resource (e.g. ncr.io).
	SourceRepository string `json:"source_repository,omitempty"`
	// ImageName is the name of the image without tag (e.g. cuda).
	ImageName string `json:"image_name,omitempty"`
	// TagRegex is a regex matching the tags which should be mirrored.
	TagRegex string `json:"tag_regex,omitempty"`
	// StartTime is
	StartTime string `json:"start_time,omitempty"`
}

// ImageMirrorStatus defines the observed state of ImageMirror
type ImageMirrorStatus struct {
	// MirroredTags is an array of tags which have already been mirrored.
	MirroredTags []string `json:"mirrored_tags"`
	// Reported Tags is array of tags which have the reported by the registry.
	SeenTags []string `json:"available_tags"`
}

// +kubebuilder:object:root=true

// ImageMirror is the Schema for the imagemirrors API
type ImageMirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageMirrorSpec   `json:"spec,omitempty"`
	Status ImageMirrorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageMirrorList contains a list of ImageMirror
type ImageMirrorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageMirror `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageMirror{}, &ImageMirrorList{})
}
