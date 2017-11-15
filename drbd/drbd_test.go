/*
Copyright 2017 LINBIT USA LLC.

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

package drbd

import "testing"

func TestDoEnoughFreeSpace(t *testing.T) {

	var freeSpaceTests = []struct {
		requestedKiB       string
		dmListFreeSpaceOut string
		out                bool
	}{
		{"5000", "3136828,16760832\n", true},
		{"50000000", "3136828,16760832\n", false},
		{"38967", "Error: Deployment node count exceeds the number of nodes in the cluster\n", false},
		{"banana", "3136828,16760832\n", false},
		{"-40", "3136828,16760832\n", false},
		{"3.14", "3136828,16760832\n", false},
	}

	for _, tt := range freeSpaceTests {
		ok, err := doEnoughFreeSpace(tt.requestedKiB, tt.dmListFreeSpaceOut)
		if ok != tt.out {
			t.Fatalf("TestDoEnoughFreeSpace(%q, %q) should be %v, got %v :%v ",
				tt.requestedKiB, tt.dmListFreeSpaceOut, tt.out, ok, err)
		}
	}
}
