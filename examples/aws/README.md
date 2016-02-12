# Deploying Diego to AWS

These instructions assume that AWS has been configured with a Cloud Foundry
deployed using [these instructions](./DEPLOYING_CF.md). In those instructions, [when setting up AWS](./DEPLOYING_CF.md#aws-requirements), `$DEPLOYMENT_DIR` was set. This variable must be set to continue.

###<a name="generating-ssh-proxy-host-key"></a>"Generating SSH Proxy Host Key and Fingerprint

In order for SSH to work for diego-release, you must generate the SSH Proxy host key and fingerprint.
This can be done by running:

```
ssh-keygen -f ssh-proxy-host-key.pem
ssh-keygen -lf ssh-proxy-host-key.pem.pub -E md5
```

The `ssh-proxy-host-key.pem` will contain the PEM encoded host key for the diego release manifest.

The md5 host key fingerprint needs to be added to the cf release manifest `cf.yml` under `properties.app_ssh.host_key_fingerprint` before you deploy cf release.

**NOTE: All the following commands should be run from the root of this repository.**

### Generate iaas-settings stub

Copy the example iaas-settings-internal.yml to the correct location.
```
cp examples/aws/templates/diego/iaas-settings-internal.yml $DEPLOYMENT_DIR/templates/diego
```

Optional: Edit this file to add overrides for the default disks for the database VMs. For example:
```
iaas_settings:
  disk_pools:
  - name: database_disks
    disk_size: 200_000
    cloud_properties: {type: gp2}
```

The following command will now generate the correct iaas-settings stub.
```
spiff merge manifest-generation/misc-templates/aws-iaas-settings.yml \
            $DEPLOYMENT_DIR/templates/diego/iaas-settings-internal.yml \
            $DEPLOYMENT_DIR/stubs/aws-resources.yml \
            > $DEPLOYMENT_DIR/stubs/diego/iaas-settings.yml
```

### Generate property-overrides stub

The `property-overrides.yml` is a stub for diego's private properties. Copy the example to the correct location.
```
cp examples/aws/templates/diego/property-overrides.yml $DEPLOYMENT_DIR/stubs/diego/property-overrides.yml
```

Edit the following keys:

  * ACTIVE_KEY_LABEL: any desired key name
  * "A SECURE PASSPHRASE": a unique passphrase
  * ALL THE CERTS: if you need to generate them, [see below](#user-content-generating-tls-certificates)
  * SSH_PROXY_HOST_KEY: this is the [key generated](#generating-ssh-proxy-host-key) earlier in these docs

### (Optional) Edit instance-count-overrides stub

Copy the example to the correct location.
```
cp examples/aws/diego/templates/instance-count-overrides-example.yml $DEPLOYMENT_DIR/stubs/diego/instance-count-overrides.yml
```
Edit it if you want to change the number of instances of each of the jobs to create.

### (Optional) Edit release-versions stub

Copy the example to the correct location.
```
cp examples/aws/diego/templates/release-versions.yml $DEPLOYMENT_DIR/stubs/diego/release-versions.yml
```

If you want to edit it, the format is:
```yml
release-versions:
  - diego: latest
  - etcd: 22
  - garden-linux: 331
```

### Generate the diego manifest

Remember that the last two arguments for `instance-count-overrides` and `release-versions`
are optional.
```
./scripts/generate-deployment-manifest \
  -c $DEPLOYMENT_DIR/deployments/cf.yml \
  -i $DEPLOYMENT_DIR/stubs/diego/iaas-settings.yml \
  -p $DEPLOYMENT_DIR/stubs/diego/property-overrides.yml \
  -n $DEPLOYMENT_DIR/stubs/diego/instance-count-overrides.yml \
  -v $DEPLOYMENT_DIR/stubs/diego/release-versions.yml \
  > $DEPLOYMENT_DIR/deployments/diego.yml
```

### Upload Garden and ETCD release
1. Upload the latest garden-linux-release:
    ```
    bosh upload release https://bosh.io/d/github.com/cloudfoundry-incubator/garden-linux-release
    ```

    If you wish to upload a specific version of garden-linux-release, or to download the release locally before uploading it, please consult directions at [bosh.io](http://bosh.io/releases/github.com/cloudfoundry-incubator/garden-linux-release).

1. Upload the latest etcd-release:
    ```
    bosh upload release https://bosh.io/d/github.com/cloudfoundry-incubator/etcd-release
    ```

    If you wish to upload a specific version of etcd-release, or to download the release locally before uploading it, please consult directions at [bosh.io](http://bosh.io/releases/github.com/cloudfoundry-incubator/etcd-release).

### Deploy Diego

These commands may take up to an hour. Be patient; it's worth it.
```
bosh deployment $DEPLOYMENT_DIR/deployments/diego.yml
bosh --parallel 10 create release --force
bosh upload release
bosh deploy
```
