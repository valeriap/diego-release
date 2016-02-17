# Set Up Dependencies

## Setting Up Local Environment

As part of our deployment process, you must install the following dependencies:

* [Go 1.4.3](https://golang.org/doc/install)
* [godep](https://github.com/tools/godep)
* [boosh](https://github.com/vito/boosh)
* [spiff](https://github.com/cloudfoundry-incubator/spiff)
* [aws cli](https://aws.amazon.com/cli/)
* [jq](https://stedolan.github.io/jq/)
* [ruby](https://www.ruby-lang.org/en/documentation/installation/)
* [bosh cli](http://bosh.io/docs/bosh-cli.html)
* [bosh init](https://bosh.io/docs/install-bosh-init.html)

You must also clone the following github repositories:

* [cf-release](https://github.com/cloudfoundry/cf-release)
* [diego-release](https://github.com/cloudfoundry-incubator/diego-release)

## Deployment Directory

Our deployment process requires that you create a directory for each deployment
which will hold the necessary configuration to deploy bosh, cf-release, and
diego-release.
This directory will be referred to as `$DEPLOYMENT_DIR` later in these instructions.

## AWS Requirements

Before deploying the bosh director, you must create the following resources in
your AWS account through the AWS console:

* IAM User Policy
  1. From the AWS console homepage, click on `Identity & Access Management`
  2. Click on the `Policies` link
  3. Click on the `Create Policy` button
  4. Select `Create Your Own Policy`
  5. Enter `bosh-aws-policy` as the `Policy Name`
  6. Enter:
  ```json
  {
    "Version": "1",
      "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "iam:DeleteServerCertificate",
          "iam:UploadServerCertificate",
          "iam:ListServerCertificates",
          "iam:GetServerCertificate",
          "cloudformation:*",
          "ec2:*",
          "s3:*",
          "vpc:*"
          "elasticloadbalancing:*",
          "route53:*"
        ],
        "Resource": "*"
      }
    ]
  }
  ```

* IAM User
  1. From the AWS console homepage, click on `Identity & Access Management`
  2. Click on `Users` link
  3. Click on the `Create New Users` button
  4. Fill in only one user name
  5. Make sure that the `Generate an access key for each user` checkbox is checked and click `Create`
  6. Click `Download Credentials` at the bottom of the screen
  7. Click  the `Cancel` link to return to the IAM Users page
  8. Select the user that you created
  9. Click the `Attach Policy` button
  10. Filter for `bosh-aws-policy` in the filter box
  11. Select `bosh-aws-policy` and click the `Attach Policy` button

* AWS keypair for your bosh director
  1.  From your AWS EC2 page click on the `Key Pairs` link
  2.  Click the `Create Key Pair` button at the top of the page
  3.  When prompted for the key name, enter `bosh`
  4.  Move the downloaded `bosh.pem` key to `$DEPLOYMENT_DIR/keypair/` and rename the key to `id_rsa_bosh`

* Route 53 Hosted Zone
  1.  From the aws console homepage, click on `Route 53`
  2.  Select `Hosted zones` from the left sidebar
  3.  Click the `Create Hosted Zone` button
  4.  Fill in the domain name for your Cloud Foundry deployment

  By default, the domain name for your hosted zone will be the root domain of all apps deployed to your cloud foundry instance.

  eg:
   ```
   domain = foo.bar.com
   app name = `hello-world`. This will create a default route of hello-world.domain

   http://hello-world.foo.bar.com will be the root url address of your application
   ```

## Deployment Directory Setup

After creating the necessary resources in AWS, you must populate the
`DEPLOYMENT_DIR` in the following format. Each of the files is explained further
below.

```
DEPLOYMENT_DIR
|-(bootstrap_environment)
|-keypair
| |-(id_rsa_bosh)
|-certs
| |-(elb-cfrouter.key)
| |-(elb-cfrouter.pem)
|-stubs
| |-(domain.yml)
| |-infrastructure
| | |-(availablity_zones.yml)
| |-bosh-init
|   |-(releases.yml)
|   |-(users.yml)
|   |-(stemcell.yml)
```

#### bootstrap_environment

This script exports your AWS default region and access/secret keys as environment variables.
The `AWS_ACCESS_KEY_ID` key must match the AWS IAM user's access key id and the `AWS_SECRET_ACCESS_KEY`
is the private key generated during the [IAM user creation](#aws-requirements).

eg:
```
export AWS_DEFAULT_REGION=us-east-1
export AWS_ACCESS_KEY_ID=xxxxxxxxxxxxxxxxxxx
export AWS_SECRET_ACCESS_KEY='xxxxxxxxxxxxxxxxxxxxxx'
```

#### keypair/id_rsa_bosh

This is the private key pair generated for the BOSH director when the [AWS keypair](#aws-requirements) was created.

#### certs/elb-cfrouter.key && certs/elb-cfrouter.pem

An SSL certificate for the domain where Cloud Foundry will be accessible is required. If you do not already provide a certificate,
you can generate a self signed cert following the commands below:

```
openssl genrsa -out elb-cfrouter.key 2048
openssl req -new -key elb-cfrouter.key -out elb-cfrouter.csr
```
You can leave all of the requested inputs blank. Then run:

```
openssl x509 -req -in elb-cfrouter.csr -signkey elb-cfrouter.key -out elb-cfrouter.pem
```

#### stubs/domain.yml

The `domain.yml` should be assigned to the domain that was generated when the [route 53 hosted zone](#aws-requirements) was created.

eg:
```yaml
---
domain: <your-domain.com>
```

#### stubs/infrastructure/availability_zones.yml

This yaml file defines the 3 zones that will host your Cloud Foundry Deployment.

eg:
```yaml
---
meta:
  availability_zones:
    - us-east-1a
    - us-east-1c
    - us-east-1d
```

Note: These zones could become restricted by AWS. If at some point during the `deploy_aws_cli` script and you see an error
similar to the following message:

```
Value (us-east-1b) for parameter availabilityZone is invalid Subnets can currently only be created in the following availability zones: us-east-1d, us-east-1b, us-east-1a, us-east-1e
```

you will need to update this file with acceptable availability zone values.

#### stubs/bosh-init/releases.yml

To deploy the bosh director, bosh-init's `releases.yml` must specify `bosh` and `bosh-aws-cpi` releases by `url` and `sha1`.

eg:
```yaml
---
releases:
  - name: bosh
    url: URL_TO_LATEST_BOSH_BOSH_RELEASE
    sha1: SHA1_OF_LATEST_BOSH_BOSH_RELEASE
  - name: bosh-aws-cpi
    url: URL_TO_LATEST_BOSH_AWS_CPI_BOSH_RELEASE
    sha1: SHA1_OF_LATEST_BOSH_AWS_CPI_BOSH_RELEASE
```

Releases for `bosh` can be found [here](http://bosh.io/releases/github.com/cloudfoundry/bosh?all=1).
Releases for `bosh-aws-cpi` can be found [here](http://bosh.io/releases/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?all=1).

#### stubs/bosh-init/users.yml

This file defines the admin users for your bosh director.

eg:
```yaml
---
BoshInitUsers:
  - {name: admin, password: YOUR_PASSWORD}
```

#### stubs/bosh-init/stemcell.yml

This file defines which stemcell to use on the bosh director. Stemcells can be found
[here](http://bosh.io/stemcells/bosh-aws-xen-ubuntu-trusty-go_agent), and must be specified by their `url` and `sha1`.

eg:
```yaml
---
BoshInitStemcell:
  url: https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3091
  sha1: 21ce6eb039179bb5b1706adfea4c161ea20dea1f
```

Currently bosh.io does not provide the sha1 of stemcells. You must download the
stemcell locally and calcuate the sha1 manually. This can be done on OSX by running:

```
shasum /path/to/stemcell
```

## Adding Security

In order to properly secure your Cloud Foundry deployment, you must generate SSL certificates and keys to secure traffic between components.

We provide two scripts to help generate the necessary SSL certificates:

1. consul cert generation
```
cd $DEPLOYMENT_DIR/certs
$CF_RELEASE_DIRECTORY/scripts/generate-consul-certs
```
1. diego cert generation
```
$DIEGO_RELEASE_DIRECTORY/scripts/generate-diego-certs
mv $DIEGO_RELEASE_DIR/diego-certs/* $DEPLOYMENT_DIR/certs
```

After running theses scripts, you should see the following output:
```
DEPLOYMENT_DIR
|-certs
|  |-consul-certs # generated via cf-release/scripts/generate-consul-certs
|  |  |- agent.crt
|  |  |- agent.key
|  |  |- server-ca.crt
|  |  |- server-ca.key
|  |  |- server.crt
|  |  |- server.key
|  |-etcd-certs # generated via cf-release/scripts/generate-diego-certs
|  |  |- client.crt
|  |  |- client.key
|  |  |- server.crt
|  |  |- server.key
|  |  |- peer.crt
|  |  |- peer.key
|  |-bbs-certs # generated via cf-release/scripts/generate-diego-certs
|  |  |- client.crt
|  |  |- client.key
|  |  |- server.crt
|  |  |- server.key
|  |- diego-ca.crt
|  |- diego-ca.key
|  |- etcd-peer-ca.crt
|  |- etcd-peer-ca.key
|-keypair
| |-(ssh-proxy-hostkey.pem)
| |-(ssh-proxy-hostkey.pem.pub)
| |-(ssh-proxy-hostkey-fingerprint)
| |-(uaa)
| |-(uaa.pem.pub)
```

###<a name="generating-ssh-proxy-host-key"></a>Generating SSH Proxy Host Key and Fingerprint

In order for SSH to work for diego-release, you must generate the SSH Proxy host key and fingerprint.
This can be done by running:

```
ssh-keygen -f $DEPLOYMENT_DIR/keypair/ssh-proxy-hostkey.pem
ssh-keygen -lf $DEPLOYMENT_DIR/keypair/ssh-proxy-hostkey.pem.pub -E md5 | cut -d ' ' -f2 | sed "s/MD5://" > $DEPLOYMENT_DIR/keypair/ssh-proxy-hostkey-fingerprint
```

The `ssh-proxy-host-key.pem` will contain the PEM encoded host key for the diego release manifest.

The md5 host key fingerprint needs to be added to the cf release manifest `cf.yml` under `properties.app_ssh.host_key_fingerprint` before you deploy cf release.

### Generating UAA Private/Public Keys

In order to properly configure UAA, you need to generate an RSA keypair.
This can be done by running the following:

```
ssh-keygen -t rsa -b 4096 -f $DEPLOYMENT_DIR/keypair/uaa
openssl rsa -in $DEPLOYMENT_DIR/uaa -pubout > $DEPLOYMENT_DIR/uaa.pub
```

#### certs/consul

These generated certificates are used to set SSL properties for the consul VMs.
By default, these properties will be set in your `stubs/cf/properties.yml`.
For more information on how to configure SSL for consul, please see [these instructions](http://docs.cloudfoundry.org/deploying/common/consul-security.html).

#### certs/etcd and certs/bbs

These generated certificates are used to configure SSL between components in Diego.
This ensures that communication with the database is secure and encrypted.
