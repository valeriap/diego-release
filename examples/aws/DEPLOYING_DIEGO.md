# Deploying Diego to AWS

These instructions assume that AWS has been configured with a Cloud Foundry
deployed using [these instructions](./DEPLOYING_CF.md).
`$DEPLOYMENT_DIR` must be set to the absoulte path to your local deployment directory.

**NOTE: All the following commands should be run from the root of this repository.**

### Fill in Generated Property Overrides Stub

In order to generate a manifest for diego-release, you need to replace certain values in the provided `$DEPLOYMENT_DIR/stubs/diego/property-overrides.yml`.
Every property that needs to be replaced is prefixed with `REPLACE_ME_WITH_`.

Here is a summary of the properties that need to be changed:
  * REPLACE_ME_WITH_ACTIVE_KEY_LABEL: any desired key name
  * REPLACE_ME_WITH_A_SECURE_PASSPHRASE: a unique passphrase associated with the active key label
  * ETCD and BBS certs: if you need to generate them, [see these instructions](#SETUP.md#adding-security)
  * SSH Proxy Host Key: this is the [key generated](SETUP.md#generating-ssh-proxy-host-key) earlier in these docs

### (Optional) Edit instance-count-overrides stub

Copy the example to the correct location.
```
cp examples/aws/diego/stubs/instance-count-overrides-example.yml $DEPLOYMENT_DIR/stubs/diego/instance-count-overrides.yml
```
Edit it if you want to change the number of instances of each of the jobs to create.

### (Optional) Edit release-versions stub

Copy the example to the correct location.
```
cp examples/aws/diego/stubs/release-versions.yml $DEPLOYMENT_DIR/stubs/diego/release-versions.yml
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
