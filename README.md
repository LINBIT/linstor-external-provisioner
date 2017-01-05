# Kubernetes external FLEX provisioner

This is an example external provisioner for kubernetes meant for use with FLEX based volume plugins.

**To Build**

```bash
make
```

**To Deploy**

Edit *examples/pod-provisioner.yaml* and make sure that *-execCommand=/opt/go/src/github.com/childsb/flex-provisioner/flex/flex/flex * points to the correct shell script.  The shell script is called when provisioning and deleting a volume

The shell script must be on all nodes in the cluster.

