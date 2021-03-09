// Copyright 2018 Google LLC All Rights Reserved.
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

package cmd

import (
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
)

// NewCmdRebase creates a new cobra.Command for the rebase subcommand.
func NewCmdRebase(options *[]crane.Option) *cobra.Command {
	var orig, oldBase, newBase, rebased string

	rebaseCmd := &cobra.Command{
		Use:   "rebase",
		Short: "Rebase an image onto a new base image",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if orig == "" {
				orig = args[0]
			} else if len(args) != 0 || args[0] != "" {
				return fmt.Errorf("cannot use --original with positional argument")
			}

			rebasedImg, err := crane.Rebase(orig, oldBase, newBase)
			if err != nil {
				return fmt.Errorf("rebasing image: %v", err)
			}

			// If the new ref isn't provided, write over the original image.
			// If that ref was provided by digest (e.g., output from
			// another crane command), then strip that and push the
			// rebased image by digest instead.
			if rebased == "" {
				log.Println("pushing rebased image as", orig)
				rebased = orig
			}
			digest, err := rebasedImg.Digest()
			if err != nil {
				return fmt.Errorf("digesting new image: %v", err)
			}
			r, err := name.ParseReference(rebased)
			if err != nil {
				return fmt.Errorf("rebasing: %v", err)
			}

			if err := crane.Push(rebasedImg, rebased, *options...); err != nil {
				return fmt.Errorf("pushing %s: %v", rebased, err)
			}
			if _, ok := r.(name.Digest); ok {
				rebased = r.Context().Digest(digest.String()).String()
			}

			if err := crane.Push(rebasedImg, rebased, *options...); err != nil {
				return fmt.Errorf("pushing %s: %v", rebased, err)
			}

			rebasedRef, err := name.ParseReference(rebased)
			if err != nil {
				return fmt.Errorf("parsing %q: %v", rebased, err)
			}

			fmt.Println(rebasedRef.Context().Digest(digest.String()))
			return nil
		},
	}
	rebaseCmd.Flags().StringVar(&orig, "original", "", "Original image to rebase; use positional arg instead")
	rebaseCmd.Flags().StringVar(&oldBase, "old_base", "", "Old base image to remove")
	rebaseCmd.Flags().StringVar(&newBase, "new_base", "", "New base image to insert")
	rebaseCmd.Flags().StringVarP(&rebased, "tag", "t", "", "Tag to apply to rebased image")
	return rebaseCmd
}
