# Creating the AWS environment

To create the AWS environment and two VMs essential to the Cloud Foundry infrastructure,
you need to run `./deploy_aws_environment create CF_RELEASE_DIRECTORY DEPLOYMENT_DIR` **from this repository**.
This may take up to 30 minutes.

The `./deploy_aws_environment` script has three possible actions.
  * `create` spins up an AWS Cloud Formation Stack based off of the stubs filled out above
  * run `update` if you change your stubs under `DEPLOYMENT_DIR/stubs/infrastructure` or there was an update to this repository
  * `skip` will upgrade your bosh director, but will not touch the AWS environment

The second parameter is the **absolute path** to CF_RELEASE_DIRECTORY.

The third parameter is your `DEPLOYMENT_DIR` and must be structured as defined above. The deployment process
generates additional stubs that include the line "GENERATED: NO TOUCHING".

The generated stubs are:
```
DEPLOYMENT_DIR
|-stubs
| |-(director-uuid.yml) # the bosh director's unique id
| |-(aws-resources.yml)  # general metadata about our cloudformation deployment
| |-cf
| | |-(stub.yml) # networks, zones, s3 buckets for our Cloud Foundry deployment
| | |-(properties.yml) # consul configuration, shared secrets
| | |-(domain.yml) # networks, zones, s3 buckets for our Cloud Foundry deployment
| |-diego
| | |-(proprety-overrides.yml)
| | |-(iaas-settings.yml)
| |-infrastructure
|   |-(certificates.yml) # for our aws-provided elb
|   |-(cloudformation.json) # aws' deployed cloudformation.json
|-deployments
| |-bosh-init
|   |-(bosh-init.yml) # bosh director deployment
```

## stubs/cf/stub.yml

As part of our deploy_aws_environment script we generate a partial stub for your
Cloud Foundry deployment. It is a generated stub that contains AWS specific information.
This stub should not be edited manually.

## stubs/cf/properties.yml

As part of our deploy_aws_environment script we copy a partial stub for your
Cloud Foundry deployment. This stub is discussed in more detail in the
[generate manifest](DEPLOYING_CF.md#generate-manifest) section.

## stubs/diego/property-overrides.yml

This stub will be used as part of Diego manifest generation and was constructed from
your deployed AWS infrastructure, as well as our default template. This stub proveds
the skeleton for our certs generated in the [Prerequisites](SETUP.md#adding-security) section,
as well as setting components log level.

## stubs/diego/iaas-settings.yml

This stub will be used as part of Diego manifest generation.
It defines the infastructure specifc settings defined by your AWS environment.

# Route53 for BOSH Director (optional)

If you want your BOSH director to be accessible using the [Route53 hosted zone](SETUP.md#aws-requirements) earlier,
you need to perform the following steps:

  1. Obtain the public IP address of the BOSH director in the EC2 dashboard
  1. Click on the `Route53` link on the AWS console
  1. Click the `Hosted Zones` link
  1. Click on the hosted zone created earlier
  1. Click the `Create Record Set` button
  1. Enter `bosh` for the `Name`.
  1. Enter the public IP address of the bosh director for the value
  1. Click the `Create` button
