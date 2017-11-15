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

// Package drbd provides helper functions for interacting with DRBD Manage.
package drbd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// EnoughFreeSpace checks to see if there's enough free space to create a new resource.
func EnoughFreeSpace(requestedKiB, replicas string) error {

	out, err := exec.Command("drbdmanage", "list-free-space", replicas).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	if _, err := doEnoughFreeSpace(requestedKiB, string(out)); err != nil {
		return err
	}

	return nil
}

func doEnoughFreeSpace(requestedKiB, dmListFreeSpaceOut string) (bool, error) {
	request, err := strconv.Atoi(requestedKiB)
	if err != nil || request < 1 {
		return false, fmt.Errorf("requsted storage must be a positive interger, got %s", requestedKiB)
	}

	free, err := strconv.Atoi(strings.Split(dmListFreeSpaceOut, ",")[0])
	if err != nil {
		return false, fmt.Errorf("unable to determine free space: %s", dmListFreeSpaceOut)
	}

	if ok := request < free; !ok {
		return ok, fmt.Errorf("not enough space available to provision a new resource: want %dKiB have %dKiB",
			request, free)
	}

	return true, nil
}
