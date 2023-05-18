// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package acceptance

import (
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/acceptance"
)

func init() {
	acceptance.DefineFlags()
}

func TestAcceptanceNodeJs(t *testing.T) {
	imageCtx, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.Test{
		{
			Name:            "simple application",
			App:             "simple",
			MustUse:         []string{nodeRuntime, nodeNPM},
			EnableCacheTest: true,
		},
		{
			// Tests a specific versions of Node.js available on dl.google.com.
			Name:    "runtime version 16.17.1",
			App:     "simple",
			Path:    "/version?want=16.17.1",
			Env:     []string{"GOOGLE_NODEJS_VERSION=16.17.1"},
			MustUse: []string{nodeRuntime},
		},
		{
			Name:                "Dev mode",
			App:                 "simple",
			Env:                 []string{"GOOGLE_DEVMODE=1"},
			MustUse:             []string{nodeRuntime, nodeNPM},
			FilesMustExist:      []string{"/workspace/server.js"},
			MustRebuildOnChange: "/workspace/server.js",
		},
		{
			// This is a separate test case from Dev mode above because it has a fixed runtime version.
			// Its only purpose is to test that the metadata is set correctly.
			Name:    "Dev mode metadata",
			App:     "simple",
			Env:     []string{"GOOGLE_DEVMODE=1", "GOOGLE_RUNTIME_VERSION=14.19.3"},
			MustUse: []string{nodeRuntime, nodeNPM},
			BOM: []acceptance.BOMEntry{
				{
					Name: "nodejs",
					Metadata: map[string]interface{}{
						"version": "14.19.3",
					},
				},
				{
					Name: "devmode",
					Metadata: map[string]interface{}{
						"devmode.sync": []interface{}{
							map[string]interface{}{"dest": "/workspace", "src": "**/*.js"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.mjs"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.coffee"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.litcoffee"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.json"},
							map[string]interface{}{"dest": "/workspace", "src": "public/**"},
						},
					},
				},
			},
		},
		{
			Name:    "simple application (custom entrypoint)",
			App:     "custom_entrypoint",
			Env:     []string{"GOOGLE_ENTRYPOINT=node custom.js"},
			MustUse: []string{nodeRuntime, nodeNPM, entrypoint},
		},
		{
			Name:       "yarn",
			App:        "yarn",
			MustUse:    []string{nodeRuntime, nodeYarn},
			MustNotUse: []string{nodeNPM, nodePNPM},
		},
		{
			Name:       "yarn (Dev Mode)",
			App:        "yarn",
			Env:        []string{"GOOGLE_DEVMODE=1"},
			MustUse:    []string{nodeRuntime, nodeYarn},
			MustNotUse: []string{nodeNPM, nodePNPM},
		},
		{
			Name:       "pnpm",
			App:        "pnpm",
			MustUse:    []string{nodeRuntime, nodePNPM},
			MustNotUse: []string{nodeNPM, nodeYarn},
		},
		{
			Name:       "runtime version with npm ci",
			App:        "simple",
			Path:       "/version?want=16.18.1",
			Env:        []string{"GOOGLE_RUNTIME_VERSION=16.18.1"},
			MustUse:    []string{nodeRuntime, nodeNPM},
			MustNotUse: []string{nodePNPM, nodeYarn},
		},
		{
			Name:       "without package.json",
			App:        "no_package",
			Env:        []string{"GOOGLE_ENTRYPOINT=node server.js"},
			MustUse:    []string{nodeRuntime},
			MustNotUse: []string{nodeNPM, nodeYarn},
		},
		{
			Name: "NPM version specified",
			// npm@8 requires nodejs@12+
			VersionInclusionConstraint: ">= 12.0.0",
			App:                        "npm_version_specified",
			MustOutput:                 []string{"npm --version\n\n8.3.1"},
			Path:                       "/version?want=8.3.1",
		},
		{
			Name: "old NPM version specified",
			// npm@5 requires nodejs@8
			VersionInclusionConstraint: "8",
			App:                        "old_npm_version_specified",
			Path:                       "/version?want=5.5.1",
			MustUse:                    []string{nodeRuntime, nodeNPM},
			MustOutput:                 []string{"npm --version\n\n5.5.1"},
			// nodejs@8 is not available on Ubuntu 22.04
			SkipStacks: []string{"google.22", "google.min.22", "google.gae.22"},
		},
	}
	for _, tc := range acceptance.FilterTests(t, imageCtx, testCases) {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			acceptance.TestApp(t, imageCtx, tc)
		})
	}
}

func TestFailuresNodeJs(t *testing.T) {
	imageCtx, cleanup := acceptance.ProvisionImages(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.FailureTest{
		{
			Name:      "bad runtime version",
			App:       "simple",
			Env:       []string{"GOOGLE_RUNTIME_VERSION=BAD_NEWS_BEARS"},
			MustMatch: "invalid Node.js version specified",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			acceptance.TestBuildFailure(t, imageCtx, tc)
		})
	}
}
