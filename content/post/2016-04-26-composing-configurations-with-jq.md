---
date: 2016-04-26
title: "Composing Configurations with JQ"
description: "Alternative methods for manifests needing more than simple variable replacements."
tags:
- bosh
- concourse
- deployment
- jq
- manifest
- pipeline
aliases:
- /blog/2016/04/26/composing-configurations-with-jq.html
---

When managing configurations for services there are often variables which need to be changed depending on the environment or use case. Different tools deal with that sort of parameterization slightly differently. For example...

 * [AWS CloudFormation](https://aws.amazon.com/cloudformation/) - stack templates have a high level `Parameter` type which can contain user-supplied values. There are built-in functions to concatenate and do some other primitive transformations.
 * [BOSH](https://bosh.io/) - manifests are actually an [ERB](http://ruby-doc.org/stdlib-2.3.1/libdoc/erb/rdoc/ERB.html) template, allowing for dynamic inclusion of environment variables, file contents, settings from configuration files, or complicated logic.
 * [Concourse](https://concourse.ci/) - manifest keys can have values of `{{variable_name}}` whose variables can be sent to `fly` with the `--load-vars-from` argument. They're simple and cannot be used for complex interpolation (e.g. `prefix-{{variable_name}}`).

With each tool having its own interpretation, it is difficult to have a common experience and shared utilities. Initially I tried using [`spiff`](https://github.com/cloudfoundry-incubator/spiff) to generate configurations, but quickly abandoned it due to confusion and bugs. For a while I was using [custom tools](/blog/2014/09/17/simplifying-my-bosh-related-workflows.html) which, among other things, made all manifests a [Twig](http://twig.sensiolabs.org/) template. Lately, I have settled on using [`jq`](https://stedolan.github.io/jq/) as a configuration builder because of its flexibility, lightweight requirements, and broad utility.


## Start With a Stack

When creating a new AWS environment from scratch, I create everything through a CloudFormation template. I created a [base template](https://github.com/dpb587/cloque/blob/master/share/example-multi/global/core/infrastructure.json) a couple years ago which continues to be a good starting point for me. Nowadays, instead of Twig templates that used...

    "CidrBlock": "{{ env['network.local']['cidr']|cidr_network }}/{{ env['network.local']['cidr']|cidr_netmask }}",

My jq-JSON configuration now looks like...

    "CidrBlock": $network.intranet.cidr

And instead of AWS calls which could look like...

    $ aws cloudformation update-stack --stack-name "core" \
      --parameters ParameterKey=KeyPairName,ParameterValue=default \
      --template-body file://<( my-custom-and-slow-command )

I could do...

    $ aws cloudformation update-stack --stack-name "core" \
      --template-body file://<( jq -n --argfile network ../network.json --from-file template.jq )

Compared to hard-coding values directly in the stack template file, this overhead is of little value. But, when you start considering how configuration parameters get reused later it might seem a bit more reasonable. For now, it's the first step towards reusable configuration inputs...


## Export the Stack

Once the VPC, subnets, security groups, and IAM profiles are created, I could go through and update all configuration files which rely on the randomly assigned `sg-*` and `subnet-*` values. However, that's a tedious process and prohibits automation. A better way is to ask AWS CloudFormation for all the "physical" resource identifiers and save it in a form that can be reused. With `jq` I can convert the stack results into a key-value set.

    $ aws cloudformation describe-stack-resources --stack-name "core" \
      | jq -n -r '.Resources | map({ "key": .LogicalResourceId, "value": .PhysicalResourceId }) | from_entries'
    { "VpcId": "vpc-a1b2c3d4",
      "Zone1SubnetPublic": "subnet-b2c3d4e5"
      ...snip... }

To further improve and automate this, I made a Concourse resource, [`aws-cloudformation-stack`](https://github.com/dpb587/aws-cloudformation-stack-resource), which can export the stack resources, outputs, and ARN as JSON files which can then be easily referenced by `jq`'s `--argfile`.


## Init a Director

While AWS CloudFormation templates are JSON-native, BOSH manifests are typically YAML-based, so using JSON is a bit different. Of course, since JSON is a valid [subset of YAML](http://yaml.org/spec/1.2/spec.html#id2759572), existing tools will still work. Assuming you convert your manifest to jq-JSON, the network section might look like...

    { "name": "bosh",
      "subnets": [
        { "cloud_properties": { "subnet": $core_stack.Zone0SubnetPublicId },
          "range": $network.intranet.zones[0].segments.private,
          "gateway": $network.intranet.zones[0].gateways.private } ] }

And the manifest generator command might expand to something like...

    $ jq -n \
      --argfile network ../network.json \
      --argfile core_stack ../core_stack/resources.json \
      --from-file bosh-init.jq \
      > bosh-init.json

I can then run `bosh-init` like normal...

    $ bosh-init deploy bosh-init.json

Once I have a director running, there are a few more configuration files which also reuse variables...


## Dependent Cloud-Config and Deployment Manifests

The same principle of bosh-init's jq-JSON manifests also applies to cloud-config and deployment manifests. For example, instead of using a static cloud-config file which changes depending on which VPC it's deployed to, I can instead generate it from my network and stack variable files...

    { "networks": [
      { "name": "public",
        "subnets": [
          { "cloud_properties": { "subnet": $core_stack.Zone0SubnetPublicId },
            "range": $network.intranet.zones[0].segments.public,
            "gateway": $network.intranet.zones[0].gateways.public } ] } ] }

If my deployment needs to reference the Elastic IPs that its stack creates, my manifest can reference it like...

    { "static_ips": [ $frontend_stack.Eip1Id, $frontend_stack.Eip2Id ] }

Sometimes I keep deployment properties in files because it's easier to manage them with external workflows (like certificates). I could use `--arg` and `cat` to pass the file contents to my configuration file...

    $ jq --arg config_ca_crt "$( cat config/ca.crt )" ...

And then I would be able to reference it like...

    { "properties": { "ca_cert": $config_ca_crt } }

With arbitrary inputs to different configurations, the `jq` invocations can easily become unwieldly and difficult to maintain. With a few conventions though, a single shell script could render any type of `jq`-based configuration...


## Reusable Renderer 

For every configuration file, whether it's for a stack, deployment, or pipeline, I start off assuming there's a `jq` filter file which can build my configuration for me. This way, I can invoke my renderer and output the result to standard out...

    $ render-command $PWD/path/to/deployment/bosh.jq
    ...rendered config...

When configurations depend on AWS CloudFormation stacks, `render-command` should assume `$PWD/*-stack` directories will contain resources, outputs, and ARN. All of which could be appended as `jq` arguments...

    jq_args="$jq_args --argfile ${stack_name}_stack $stack_dir/resources.json"
    jq_args="$jq_args --arg ${stack_name}_stack_arn \"$stack_arn\""
    jq_args="$jq_args --argfile ${stack_name}_stack_output $stack_dir/outputs.json"

When configurations depend on local, arbitrary files, `render-command` should assume that anything in the `$PWD/path/to/deployment/config` directory might be referenced. All `*.json` files can be pre-parsed as JSON to allow `$config_auth.username`-type references, and all other files can be loaded as plain strings. Effectively...

    jq_args="$jq_args --argfile config_auth $manifest_dir/config/auth.json"
    jq_args="$jq_args --arg config_ca_crt \"$( cat $manifest_dir/config/ca.crt )\""

An example implementation of this script is gisted [here]({{< appendix-ref "2016-04-26-composing-configurations-with-jq/render.sh" >}}).


## Pipelining

Once a single command can be used to render configurations, it becomes easier to automate this in Concourse. One remaining step is to document the configuration requirements as a task. For example, a simple website might require a `core-stack` (which created the network subnets), a `bosh-stack` (which created the security group that bosh-agent needs), and a `website-stack` (which created the HTTP/HTTPS security group the deployment needs) all in order to generate the deployment manifest in `deployment/manifest.yml`...

    ---
    platform: "linux"
    image: "docker:///dpb587/local#jq"
    inputs:
      - name: "repo"
      - name: "core-stack"
      - name: "bosh-stack"
      - name: "website-stack"
    outputs:
      - name: "deployment"
    run:
      path: "repo/bin/render"
      args:
        - "repo/aws-use1/website/bosh.jq"
    params:
      STDOUT: "deployment/manifest.yml"

Once the Concourse task is created, a pipeline could watch for configuration changes and trigger deploys to AWS CloudFormation, BOSH, and/or Concourse. If an upstream service dependency changes, it could also trigger a local update as well.

![Screenshot: Concourse Pipeline](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2016-04-26-composing-configurations-with-jq/pipeline.jpg)


## Conclusion

Sometimes automation is ignored because things change so infrequently (e.g. you might rarely recreate a production VPC so manually copy/pasting resource identifiers is uncommon and considered acceptable), but any stage requiring manual intervention is error-prone and is not reproducible. Extracting those critical variables for environments and deployments allows for better automation. Having a common way to generate configurations from those variables makes integration of both external and internal tooling easier.
