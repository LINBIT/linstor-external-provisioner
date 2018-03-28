# Linstor External FLEX Provisioner for Kubernetes

This provisioner is meant for use with with the Linstor FlexVolume Plugin and allows
administrators to create storage classes with varying levels of redundancy and
users to create new Linstor resources using the familiar PV/PVC workflow.


# Building

This project is written in Go. If you haven't built a Go program before,
please refer to this [helpful guide](https://golang.org/doc/install).

Requires Go 1.8 or higher and a configured GOPATH

```bash
`mkdir -p $GOPATH/src/github.com/LINBIT/`

`cd $GOPATH/src/github.com/LINBIT/`

`git clone https://github.com/LINBIT/linstor-flex-provision`

`cd linstor-flex-provision`

make
```
This will create a binary named `linstor-external-provision` in the root of the project.

# Deployment

This provisioner must run directly on one of the Linstor nodes as root.
It needs to be passed the provisioner name, which will be referenced in
storage classes that use this provisioner.

As well as the path to a Kubernetes configuration file:

```bash
./linstor-external-provisioner -provisioner=external/linstor -kubeconfig=$HOME/.kube/config &> /path/to/logfile &
```
or the Kubernetes master IP:

```bash
./linstor-external-provisioner -provisioner=external/linstor -master=http://0.0.0.0:8080 &> /path/to/logfile &
```
# Usage

After the provisioner has been deployed you're free to create storage classes and
have your users start provisioning volumes from them. Administrators must set the
list of nodes that resources will be deployed to and the storage pool that these
resources will consume to create storage. Please see the class.yaml and pvc.yaml
files in the root of this project for examples.

On the successful creation of a PV, a new linstor resource with the same name as the
PV is created as well.

# License

Apache 2.0
