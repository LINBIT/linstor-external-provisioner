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
	"os"
	"strconv"
	"strings"

	"github.com/LINBIT/golinstor"
	"github.com/golang/glog"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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

	nodeList            []string
	replicasOnSame      []string
	replicasOnDifferent []string
	storagePool         string
	disklessStoragePool string
	blockSize           string
	force               string
	xfsdiscardblocks    string
	xfsDataSU           string
	xfsDataSW           string
	xfsLogDev           string
	mountOpts           string
	fsOpts              string
	autoPlace           uint64
	doNotPlaceWithRegex string
	controllers         string
	requestedSize       uint64
	encryption          bool
}

var _ controller.Provisioner = &flexProvisioner{}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *flexProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	if err := p.validateOptions(options); err != nil {
		return nil, err
	}

	resourceName := fmt.Sprintf("%s-%s",
		options.PVC.ObjectMeta.Namespace, options.PVC.ObjectMeta.Name)

	err := p.createVolume(options, resourceName)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy

	annotations[annProvisionerId] = string(p.identity)
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        resourceName,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: options.PVC.Spec.Resources.Requests[v1.ResourceStorage],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{

				FlexVolume: &v1.FlexPersistentVolumeSource{
					Driver: p.driver,
					Options: map[string]string{
						"disklessStoragePool": p.disklessStoragePool,
						"blockSize":           p.blockSize,
						"force":               p.force,
						"xfsDiscardBlocks":    p.xfsdiscardblocks,
						"xfsDataSU":           p.xfsDataSU,
						"xfsDataSW":           p.xfsDataSW,
						"xfsLogDev":           p.xfsLogDev,
						"fsOpts":              p.fsOpts,
						"mountOpts":           p.mountOpts,
						"controllers":         p.controllers,
					},
					FSType:   p.fsType,
					ReadOnly: p.isRO,
				},
			},
		},
	}

	return pv, nil
}

func (p *flexProvisioner) createVolume(volumeOptions controller.VolumeOptions, resourceName string) error {

	if volumeOptions.PVC.Spec.Selector != nil {
		val, ok := volumeOptions.PVC.Spec.Selector.MatchLabels["linstorDoNotPlaceWith"]
		if ok && val == "true" {
			p.doNotPlaceWithRegex = fmt.Sprintf("%s-.*", resourceName)
		}
	}

	r := linstor.NewResourceDeployment(
		linstor.ResourceDeploymentConfig{
			Name:                resourceName,
			NodeList:            p.nodeList,
			SizeKiB:             p.requestedSize,
			StoragePool:         p.storagePool,
			DisklessStoragePool: p.disklessStoragePool,
			AutoPlace:           p.autoPlace,
			DoNotPlaceWithRegex: p.doNotPlaceWithRegex,
			ReplicasOnSame:      p.replicasOnSame,
			ReplicasOnDifferent: p.replicasOnDifferent,
			Encryption:          p.encryption,
			Controllers:         p.controllers,
			LogOut:              os.Stderr,
		})

	return r.CreateAndAssign()
}

func (p *flexProvisioner) validateOptions(volumeOptions controller.VolumeOptions) error {

	// These need to be cleared as they seem to retain old values
	p.autoPlace = 0
	p.blockSize = ""
	p.controllers = ""
	p.disklessStoragePool = ""
	p.doNotPlaceWithRegex = ""
	p.driver = "linbit/linstor-flexvolume"
	p.encryption = false
	p.force = ""
	p.fsType = "ext4"
	p.isRO = true
	p.nodeList = []string{}
	p.replicasOnSame = []string{}
	p.replicasOnDifferent = []string{}
	p.storagePool = ""
	p.xfsDataSU = ""
	p.xfsDataSW = ""
	p.xfsLogDev = ""
	p.xfsdiscardblocks = ""
	p.fsOpts = ""
	p.mountOpts = ""

	for k, v := range volumeOptions.Parameters {
		switch strings.ToLower(k) {
		case "nodelist":
			p.nodeList = strings.Split(v, " ")
		case "replicasonsame":
			p.replicasOnSame = strings.Split(v, " ")
		case "replicasondifferent":
			p.replicasOnDifferent = strings.Split(v, " ")
		case "driver":
			p.driver = v
		case "filesystem":
			p.fsType = v
		case "storagepool":
			p.storagePool = v
		case "disklessstoragepool":
			p.disklessStoragePool = v
		case "autoplace":
			if v == "" {
				v = "0"
			}
			autoplace, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return fmt.Errorf("unable to parse %q as an integer", v)
			}
			p.autoPlace = autoplace
		case "donotplacewithregex":
			p.doNotPlaceWithRegex = v
		case "blocksize":
			p.blockSize = v
		case "force":
			p.force = v
		case "xfsdiscardblocks":
			p.xfsdiscardblocks = v
		case "xfsdatasu":
			p.xfsDataSU = v
		case "xfsdatasw":
			p.xfsDataSW = v
		case "xfslogdev":
			p.xfsLogDev = v
		case "mountopts":
			p.mountOpts = v
		case "fsopts":
			p.fsOpts = v
		case "controllers":
			p.controllers = v
		case "encryptvolumes":
			if strings.ToLower(v) == "yes" {
				p.encryption = true
			}
		case "readonly":
			if isRO, err := strconv.ParseBool(v); err == nil {
				p.isRO = isRO
			}
			// External provisioner spec says to reject unknown parameters.
		default:
			glog.Warningf("Unknown StorageClass Parameter: %s", k)
		}
	}

	capacity := volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceStorage]
	requestedBytes := capacity.Value()
	p.requestedSize = uint64((requestedBytes / 1024) + 1)

	return nil
}
