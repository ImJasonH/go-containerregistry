// Copyright 2021 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crane

import (
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	baseDigestAnnotation = "org.opencontainers.image.base.digest"
	baseRefAnnotation    = "org.opencontainers.image.base.ref.name"
)

// Rebase parses the references and uses them to perform a rebase.
//
// If oldBase or newBase are "", Rebase attempts to derive them using
// annotations in the original image. If those annotations are not found,
// Rebase returns an error.
//
// If rebasing is successful, base image annotations are set on the resulting
// image to facilitate implicit rebasing next time.
func Rebase(orig, oldBase, newBase string, opt ...Option) (v1.Image, error) {
	o := makeOptions(opt...)
	origRef, err := name.ParseReference(orig, o.name...)
	if err != nil {
		return nil, fmt.Errorf("parsing tag %q: %v", origRef, err)
	}
	origImg, err := remote.Image(origRef, o.remote...)
	if err != nil {
		return nil, err
	}

	m, err := origImg.Manifest()
	if err != nil {
		return nil, err
	}
	if newBase == "" && m.Annotations != nil {
		newBase = m.Annotations[baseRefAnnotation]
		if newBase != "" {
			log.Printf("Detected new base from %q annotation: %s", baseRefAnnotation, newBase)
		}
	}
	if newBase == "" {
		return nil, fmt.Errorf("either newBase or %q annotation is required", baseRefAnnotation)
	}
	newBaseRef, err := name.ParseReference(newBase, o.name...)
	if err != nil {
		return nil, err
	}
	if oldBase == "" && m.Annotations != nil {
		oldBase = m.Annotations[baseDigestAnnotation]
		if oldBase != "" {
			oldBase = newBaseRef.Context().Digest(oldBase).String()
			log.Printf("Detected old base from %q annotation: %s", baseDigestAnnotation, oldBase)
		}
	}
	if oldBase == "" {
		return nil, fmt.Errorf("either oldBase or %q annotation is required", baseDigestAnnotation)
	}

	oldBaseRef, err := name.ParseReference(oldBase, o.name...)
	if err != nil {
		return nil, err
	}
	oldBaseImg, err := remote.Image(oldBaseRef, o.remote...)
	if err != nil {
		return nil, err
	}
	newBaseImg, err := remote.Image(newBaseRef, o.remote...)
	if err != nil {
		return nil, err
	}

	rebased, err := mutate.Rebase(origImg, oldBaseImg, newBaseImg)
	if err != nil {
		return nil, err
	}

	// Update base image annotations for the new image manifest.
	d, err := newBaseImg.Digest()
	if err != nil {
		return nil, err
	}
	newDigest := d.String()
	newTag := newBaseRef.String()
	log.Printf("Setting annotation %q: %q", baseDigestAnnotation, newDigest)
	log.Printf("Setting annotation %q: %q", baseRefAnnotation, newTag)
	return mutate.Annotations(rebased, map[string]string{
		baseDigestAnnotation: newDigest,
		baseRefAnnotation:    newTag,
	}), nil
}
