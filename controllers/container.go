/*
Copyright (c) 2020 Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the spec.fic language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"

	// This dependency was copied into the operator to avoid client-go
	// dependency conflicts between flux and kubebuilder. This may or
	// may not be the best solution to an unfortunately common problem.
	//"github.com/fluxcd/flux/pkg/policy"

	slipwayk8sfacebookcomv1 "github.com/davidewatson/slipway/api/v1"
)

func init() {
	// Override the default keychain used by this process to follow the
	// Kubelet's keychain semantics.
	kc, err := k8schain.NewInCluster(k8schain.Options{})
	if err != nil {
		authn.DefaultKeychain = kc
	}
}

// Union takes two slices of string, say a and b, and returns a slice c
// such that for all x exist in c -> x exist in a _or_ x exist in b.
func Union(a, b []string) (c []string) {
	seen := make(map[string]bool)
	for _, item := range b {
		seen[item] = true
		c = append(c, item)
	}

	for _, item := range a {
		if _, ok := seen[item]; !ok {
			c = append(c, item)
		}
	}
	return
}

// Intersection takes two slices of string, say a and b, and returns a slice c
// such that for all x exist in c -> x exist in a _and_ x exist in b.
func Intersection(a, b []string) (c []string) {
	seen := make(map[string]bool)
	for _, item := range b {
		seen[item] = true
	}

	for _, item := range a {
		if _, ok := seen[item]; ok {
			c = append(c, item)
		}
	}
	return
}

// Difference takes two slices of string, say a and b, and returns a slice c
// such that for all x exist in c -> x exist in a and x _not_ exist in b.
func Difference(a, b []string) (c []string) {
	seen := make(map[string]bool)
	for _, item := range b {
		seen[item] = true
	}

	for _, item := range a {
		if _, ok := seen[item]; !ok {
			c = append(c, item)
		}
	}
	return
}

// Filter takes a slice of tags and returns a new slice such that each tag
// matches the policy determined by pattern.
func Filter(tags []string, pattern string) []string {
	p := NewPattern(pattern)

	passed := make([]string, 0)
	for _, tag := range tags {
		if p.Matches(tag) {
			passed = append(passed, tag)
		}
	}

	return passed
}

// GetOptions returns a slice of remote.Options including a keychain
// for verifying the server, and a token to authenticate ourselves.
func GetOptions(token string) (options []remote.Option) {
	if token != "" {
		options = append(options, remote.WithAuth(&authn.Bearer{Token: token}))
	}
	options = append(options, remote.WithAuthFromKeychain(authn.DefaultKeychain))

	return
}

// MirrorImage lists all tags for the image from the source repository
// and writes them to the destination repository iff they are not already
// there, and they match pattern. Returns the tags already mirrored, and
// an error, if any.
func MirrorImage(ctx context.Context, imageMirror slipwayk8sfacebookcomv1.ImageMirror, log logr.Logger,
	sourceToken, destToken string) ([]string, error) {

	sourceOptions := GetOptions(sourceToken)
	destOptions := GetOptions(destToken)
	sourceName := imageMirror.Spec.SourceRepo + imageMirror.Spec.ImageName
	destName := imageMirror.Spec.DestRepo + imageMirror.Spec.ImageName

	sourceRepo, err := name.NewRepository(sourceName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to NewRegistry")
	}

	sourceTags, err := remote.ListWithContext(ctx, sourceRepo, sourceOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ListWithContext")
	}
	log.Info("Source repository tags", "sourceTags", sourceTags)

	destRepo, err := name.NewRepository(destName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to NewRegistry")
	}

	destTags, err := remote.ListWithContext(ctx, destRepo, destOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ListWithContext")
	}
	log.Info("Destination repository tags", "destTags", destTags)

	filteredTags := Filter(sourceTags, imageMirror.Spec.Pattern)
	mirroredTags := Intersection(filteredTags, destTags)
	missingTags := Difference(filteredTags, destTags)

	log.Info("Filtered source repository tags", "filteredTags", filteredTags)
	log.Info("Mirrored destination tags", "mirroredTags", mirroredTags)
	log.Info("Missing destination tags", "missingTags", missingTags)

	for _, tag := range missingTags {
		sourceNameWithTag := sourceName + ":" + tag
		destNameWithTag := destName + ":" + tag

		sourceRef, err := name.ParseReference(sourceNameWithTag)
		if err != nil {
			return mirroredTags, errors.Wrap(err, "unable to ParseReference")
		}

		destRef, err := name.ParseReference(destNameWithTag)
		if err != nil {
			return mirroredTags, errors.Wrap(err, "unable to ParseReference")
		}

		img, err := remote.Image(sourceRef, sourceOptions...)
		if err != nil {
			return mirroredTags, errors.Wrap(err, "unable to Image")
		}

		err = remote.Write(destRef, img, destOptions...)
		if err != nil {
			return mirroredTags, errors.Wrap(err, "unable to Write")
		}
	}

	return Union(mirroredTags, missingTags), nil
}
