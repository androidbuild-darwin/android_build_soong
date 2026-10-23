// Copyright 2019 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cc

import (
	"runtime"
	"strings"
	"testing"

	"android/soong/android"
)

func TestIsThirdParty(t *testing.T) {
	thirdPartyPaths := []string{
		"external/foo/",
		"vendor/bar/",
		"hardware/underwater_jaguar/",
	}
	nonThirdPartyPaths := []string{
		"vendor/google/cts/",
		"hardware/google/pixel",
		"hardware/interfaces/camera",
		"hardware/ril/supa_ril",
		"bionic/libc",
	}
	for _, path := range thirdPartyPaths {
		if !android.IsThirdPartyPath(path) {
			t.Errorf("Expected %s to be considered third party", path)
		}
	}
	for _, path := range nonThirdPartyPaths {
		if android.IsThirdPartyPath(path) {
			t.Errorf("Expected %s to *not* be considered third party", path)
		}
	}
}

func TestSizeOptimizationMlgo(t *testing.T) {
	for _, setting := range []string{"", "false", "true"} {
		t.Run("THINLTO_USE_MLGO="+setting, func(t *testing.T) {
			result := android.GroupFixturePreparers(
				prepareForCcTest,
				android.FixtureMergeEnv(map[string]string{"THINLTO_USE_MLGO": setting}),
			).RunTestWithBp(t, `
				cc_library_shared {
					name: "libfoo",
					srcs: ["foo.c"],
					optimize_for_size: true,
				}
			`)
			module := result.ModuleForTests("libfoo", "android_arm64_armv8-a_shared")
			flags := module.Rule("ld").Args["ldFlags"]
			got := strings.Contains(flags, "-enable-ml-inliner=release")
			if want := runtime.GOOS != "darwin" && setting != "false"; got != want {
				t.Errorf("MLGO inliner enabled = %v, want %v", got, want)
			}
			android.AssertStringDoesContain(t, "ThinLTO remains enabled", flags, "-flto=thin")
		})
	}
}
