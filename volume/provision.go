/*
Copyright 2017 LINBIT USA LLC.
Copyright 2016 The Kubernetes Authors.

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

package volume

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
	dm "github.com/linbit/godrbdmanage"

	"github.com/kubernetes-incubator/nfs-provisioner/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
)

const (
	// Name of the file where an s3fsProvisioner will store its identity
	identityFile = "flex-provisioner.identity"

	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "flex-dynamic-provisioner"

	// A PV annotation for the identity of the s3fsProvisioner that provisioned it
	annProvisionerId = "Provisioner_Id"
)

func NewFlexProvisioner(client kubernetes.Interface) controller.Provisioner {
	return newFlexProvisionerInternal(client)
}

func newFlexProvisionerInternal(client kubernetes.Interface) *flexProvisioner {
	var identity types.UID

	provisioner := &flexProvisioner{
		client:   client,
		identity: identity,
	}

	return provisioner
}

type flexProvisioner struct {
	client   kubernetes.Interface
	identity types.UID

	driver string
	fsType string
	isRO   bool

	replicas      string
	requestedSize string
}

var _ controller.Provisioner = &flexProvisioner{}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *flexProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	if err := p.validateOptions(options); err != nil {
		return nil, err
	}

	err := p.createVolume(options)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy

	annotations[annProvisionerId] = string(p.identity)
	/*
		This PV won't work since there's nothing backing it.  the flex script
		is in flex/flex/flex  (that many layers are required for the flex volume plugin)
	*/
	pv := &v1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:        options.PVName,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{

				FlexVolume: &v1.FlexVolumeSource{
					Driver:   p.driver,
					Options:  map[string]string{},
					FSType:   p.fsType,
					ReadOnly: p.isRO,
				},
			},
		},
	}

	return pv, nil
}

func (p *flexProvisioner) createVolume(volumeOptions controller.VolumeOptions) error {
	resourceName := volumeOptions.PVName

	r := dm.Resource{
		Name: resourceName,
	}

	ok, err := r.Exists()
	if err != nil {
		return fmt.Errorf("failed to check for previous resource assignment: %v", err)
	}
	if ok {
		return fmt.Errorf("resource %s already exists in the cluster, refusing to reassign", r.Name)
	}

	glog.Infof("Calling drbdmanage with the following args: %s %s %s %s %s", "add-volume",
		resourceName, p.requestedSize, "--deploy", p.replicas)

	cmd := exec.Command("drbdmanage", "add-volume", resourceName, p.requestedSize, "--deploy", p.replicas)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Errorf("Failed to create volume %s, output: %s, error: %s", volumeOptions, output, err.Error())
		return err
	}

	ok, err = r.WaitForAssignment(5)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("resource assignment failed for unknown reason")
	}

	return nil
}

func (p *flexProvisioner) validateOptions(volumeOptions controller.VolumeOptions) error {
	// We can tolerate no replicationLevel being set. Let's assume they want some
	// level of redundancy, since that's why people use DRBD.
	p.driver = "linbit/drbd"
	p.fsType = "ext4"
	p.isRO = true
	p.replicas = "2"
	for k, v := range volumeOptions.Parameters {
		switch strings.ToLower(k) {
		case "replicationlevel":
			if i, err := strconv.Atoi(strings.ToLower(v)); err != nil || i < 1 {
				return fmt.Errorf("bad StorageClass configuration: replicationLevel must be an interger larger than zero, got %s", v)
			}
			p.replicas = v
			// External provisioner spec says to reject unknown parameters.
		case "driver":
			p.driver = v
		case "filesystem":
			p.fsType = v
		case "readonly":
			if isRO, err := strconv.ParseBool(v); err == nil {
				p.isRO = isRO
			}
		default:
			glog.Warningf("Unknown StorageClass Parameter: %s", k)
		}
	}

	capacity := volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	requestedBytes := capacity.Value()
	requestedSize := fmt.Sprintf("%d", int((requestedBytes/1024)+1))
	p.requestedSize = requestedSize + "KiB"

	if err := dm.EnoughFreeSpace(requestedSize, p.replicas); err != nil {
		return fmt.Errorf("not enough space to create a new resource: %v", err)
	}

	return nil
}
