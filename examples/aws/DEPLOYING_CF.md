# Deploying Cloud Foundry

## Manifest Generation

To deploy Cloud Foundry, you need a stub similar to the one from the [Cloud Foundry Documentation](http://docs.cloudfoundry.org/deploying/aws/cf-stub.html).
The generated stub `DEPLOYMENT_DIR/stubs/cf/stub.yml` already has a number of these properties filled out for you.
The provided stub `DEPLOYMENT_DIR/stubs/cf/properties.yml` has some additional properties that need to be specified.
For more information on stubs for cf-release manifest generation, please refer to the documentation [here](http://docs.cloudfoundry.org/deploying/aws/cf-stub.html#editing).

### Fill in Properties Stub

In order to correctly generate a manifest for the cf-release deployment, you must
replace certain values in the provided `$DEPLOYMENT_DIR/stubs/cf/properties.yml`.
Every value that needs to be replaced is prefixed with `REPLACE_ME_WITH`.

### Diego Stub

Cloud Foundry Documention manifest generation doesn't create some VMs and properties that Diego depends on.
It also includes some unnecessary VMs and properties that Diego doesn't need. To correct this, the provided cf stub `./stubs/cf/diego.yml` is used when generating the Cloud Foundry manifest.

### Generate

After following the instructions to generate a `DEPLOYMENT_DIR/stubs/cf/stub.yml` stub and downloading the cf-release directory, run
the following command **inside this repository** to generate the Cloud Foundry manifest:

```
./scripts/generate_deployment_manifest aws \
  $DEPLOYMENT_DIR/stubs/director-uuid.yml \
  $DIEGO_RELEASE_DIR/examples/aws/stubs/cf/diego.yml \
  $DEPLOYMENT_DIR/stubs/cf/properties.yml \
  $DEPLOYMENT_DIR/stubs/cf/stub.yml \
  > $DEPLOYMENT_DIR/deployments/cf.yml
```

## Target the BOSH Director

Target the BOSH director using either the public IP address or the Route53 record created earlier.
The public IP address can be obtained from either the `$DEPLOYMENT_DIR/stubs/aws-resources.yml`
under `Resources.EIP.BoshInit` or from the EC2 dashboard in the AWS console.

```
bosh target bosh.YOUR_CF_DOMAIN
```

When prompted for the username and password, they are the credentials set in the `DEPLOYMENT_DIR/stubs/bosh-init/users.yml` stub.

## Upload the BOSH Stemcell

Upload the lastest BOSH stemcell for AWS to the bosh director.
You can find the latest stemcell [here](http://bosh.io/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent).

```
bosh upload stemcell /path/to/stemcell
```

## Create and Upload the CF Release

In order to deploy CF Release, you must create and upload the release to the director using the following commands:

```
cd $CF_RELEASE_DIR
bosh --parallel 10 create release
bosh upload release
```

## Deploy

Set the deployment manifest and deploy with the following commands:

```
bosh deployment $DEPLOYMENT_DIR/deployments/cf.yml
bosh deploy
```

From here, follow the documentation on [deploying a Cloud Foundry with BOSH](http://docs.cloudfoundry.org/deploying/common/deploy.html). Note that the deployment
can take up to 30 minutes.
