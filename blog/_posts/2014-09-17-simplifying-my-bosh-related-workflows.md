---
title: "Simplifying My BOSH-related Workflows"
layout: "post"
tags: [ "aws", "bosh", "cloudformation", "cloudfoundry", "cloque", "docker", "ec2", "packaging", "snapshots", "twig" ]
description: "Discussing some commands and wrappers I've been adding on top of BOSH."
---

Over the last nine months I've been getting into [BOSH][1] quite a bit. Historically, I've been [reluctant][20] to
invest in BOSH because I don't entirely agree with its architecture and steep learning curve. BOSH
[describes itself][1] with...

 > BOSH installs and updates software packages on large numbers of VMs over many IaaS providers with the absolute
 > minimum of configuration changes.
 > 
 > BOSH orchestrates initial deployments and ongoing updates that are:
 > 
 >  * Predictable, repeatable, and reliable
 >  * Self-healing
 >  * Infrastructure-agnostic

With continued use and experience necessitated from the [logsearch][2] project, I saw ways it would solve more critical
problems for me than it would create. For that reason, I started experimenting and migrating some services over to
BOSH to better evaluate it for my own uses. To help bridge the gap between BOSH inconveniences and some of my
architectural/practical differences I've been making a tool called [`cloque`][3].

You might find the ideas more useful rather than the `cloque` code itself - it is, after all, experimental and written
in PHP (since that's why I'm most productive in) whereas `bosh` is more Ruby/Go-oriented.


## Infrastructure First

Generally speaking, BOSH needs some help with infrastructure (i.e. it can't create its own VPC, network routing tables,
etc). Additionally, sometimes deployments don't even need the BOSH overhead. Within `cloque`, I've split management
tasks into two components:

 * Infrastructure - this is more of the "physical" layer defining the networking layer, some independent services (e.g.
   NAT gateways, VPN servers), security groups, and other core or non-BOSH functionality.
 * BOSH - everything related to BOSH (e.g. director, deployment, snapshots, releases, stemcells) which is deployed onto
   the infrastructure somewhere.

Since BOSH depends on some infrastructure, we'll get started with that first. One key to a `cloque`-managed environment
is that each environment has its own directory which includes a `network.yml` in the top-level. The network may be
located in a single datacenter, or it could span multiple countries. The file defines all the basics about the network
including subnets, reserved IPs, basic cloud properties, and some logical names.

I've committed an example network to the [`share`][7] directory within `cloque` and will use that in the examples here.
To get started, we'll copy the example and work with it...

    # copy the sample environment
    $ cp -r ~/cloque/share/example-multi ~/cloque-acme-dev
    $ cd ~/cloque-acme-dev

    # this will help the command know where to look for configs later
    $ export CLOQUE_BASEDIR="$PWD"

If you take a look at the sample [`network.yml`][18], you'll see a couple regions with their individual network
segments, VPN networks, and a few reserved IP addresses which can be referenced elsewhere. Once `network.yml` is
created, the `utility:initialize-network` task can take care of bootstrapping the following:

 * create stub folders for your different regions; e.g. `aws-apne1/core`, `global/private`)
 * create a new SSH key (in `global/private/cloque-{yyyymmdd}*.pem`) and upload it to the AWS regions being used
 * create a new IAM user, access key, and EC2 policy for BOSH to use
 * create a certificate authority for [OpenVPN][8] usage
 * create both client/server certificates for the inter-region VPN connections (requires interactive prompts for
   passwords/confirmations)
 * create an S3 bucket for shared configuration storage

When run, it assumes AWS credentials can be discovered from the environment...

    $ cloque utility:initialize-network
    > local:fs/global -> created
    ...snip...

 > I created `utility:initiailize-network` because I found myself reusing keys and buckets across multiple environments
 > (such as development vs production) because they were annoying to manage by hand. I wanted to make security easier
 > for myself and, in the process, simplify the processes through automation.

The top-level `global` directory is intended for configuration which applies to all areas. With the example I use it to
create an additional IAM role which allows VPN gateways to securely download their VPN keys and configuration files...

    $ ( cd global/core && cloque infra:put --aws-cloudformation 'Capabilities=["CAPABILITY_IAM"]' )
    > validating...done
    > checking...missing
    > deploying...done
    > waiting...CREATE_IN_PROGRESS...........................CREATE_COMPLETE...done

The `infra:put` is the core command responsible for managing the low-level, infrastructure-related resources. The
command looks for an `infrastructure.json` file (see the [example][27]) and since I'm focused on [AWS][4], the files
are [CloudFormation][5] scripts.

 > One thing I dislike about BOSH is how it uses a state file or global options to specify the director/deployment. It
 > makes it very inconvenient to quickly switch between directors/deployments even between multiple terminal sessions.
 > To help with that, `cloque` respects environment variables (or command line options) to know where it should be
 > working from. The `CLOQUE_BASEDIR` (exported earlier) is the most significant, and it was able to detect when it was
 > working from the `global` region/director and `core` deployment based on the current directory.

Now that the global resources have been created, we can create our "core" resources for the `us-west-2` region. If you
take a look at the [infrastructure.json][28] file, you'll see it creates a VPC, multiple subnets for each availability
zone, a couple base security groups, and a gateway instance which will function as a VPN server to allow inter-region
communication. You'll also notice it's using [Twig][10] templating to load `network.yml` and simplify what would be a
lot of repeated resources. We'll use the `infra:put` command again, but this time within the `aws-usw2/core`
directory...

    $ cd aws-usw2
    $ ( cd core && cloque infra:put )
    ...snip...
    > waiting...CREATE_IN_PROGRESS.........................CREATE_COMPLETE...done

 > BOSH supports ERB-templated deploy manifests. With ERB I found myself repeating a lot of code in each manifest when
 > trying to make it dynamic. After trying [spiff][21] (which I found a bit limited and difficult to understand), I
 > decided to use a different approach - one that would allow for the same dynamic, peer-config referencing, and
 > (later) transformational capabilities for both infrastructure configuration and BOSH deployment manifests.

Once the `infra:put` command finishes, the `aws-usw2` part of the environment is complete which means the OpenVPN
server is ready for a client. First we'll need to create and sign a client certificate though...

    # temporary directory
    $ mkdir tmp-myovpn
    $ cd tmp-myovpn

    # create a key (named after the hostname and current date)
    $ TMPOVPN_CN=$(hostname -s)-$(date +%Y%m%da)
    $ openssl req \
      -subj "/C=US/ST=CO/L=Denver/O=ACME Inc/OU=client/CN=${TMPOVPN_CN}/emailAddress=`git config user.email`" \
      -days 3650 -nodes \
      -new -out openvpn.csr \
      -newkey rsa:2048 -keyout openvpn.key
    Generating a 2048 bit RSA private key
    .............................+++
    ................+++
    writing new private key to 'openvpn.key'
    -----

    # sign the certificate (you'll need to enter the PKI password you used in the first step)
    $ cloque openvpn:sign-certificate openvpn.csr

    # now create the OpenVPN configuration profile for connecting to aws-usw2
    $ ( \
        cloque openvpn:generate-profile aws-usw2 $TMPOVPN_CN \
        ; echo '<key>' \
        ; cat openvpn.key \
        ; echo '</key>' \
      ) > acme-dev-aws-usw2.ovpn

    # opening should install it with a GUI connection manager like Tunnelblick
    $ open acme-dev-aws-usw2.ovpn

    # cleanup
    $ cd ../
    $ rm -fr tmp-myovpn
    $ unset TMPOVPN_CN

 > I created the `openvpn:sign-certificate` and, namely, `openvpn:generate-profile` commands to make the steps highly
 > reproducible to encourage better certificate usage practices through it's "trivialness".

Since I'm using `example.com` in the `share` scripts as the domain, DNS won't resolve it. For now, the easiest solution
is to manually add an entry to `/etc/hosts`...

    $ echo "`cd core && cloque infra:get '.Z0GatewayEipId'` gateway.aws-usw2.acme-dev.cloque.example.com" \
      | sudo tee -a /etc/hosts

 > The `infra:get` command allows me to programmatically fetch configuration details about the current deployment. For
 > infrastructure, this allows me to extract the created resource IDs/names using [jq][12] statements. This makes it
 > extremely easy to automate basic lookup tasks (as in this case), but also allows for more complex IP or security
 > group enumeration which can be used for other composable, automated tasks.

Once `/etc/hosts` is updated, I can connect with an OpenVPN client like [Tunnelblick][13] and ping the network...

    $ ping -c 5 10.101.0.4
    PING 10.101.0.4 (10.101.0.4): 56 data bytes
    64 bytes from 10.101.0.4: icmp_seq=0 ttl=64 time=59.035 ms
    64 bytes from 10.101.0.4: icmp_seq=1 ttl=64 time=61.288 ms
    64 bytes from 10.101.0.4: icmp_seq=2 ttl=64 time=78.194 ms
    64 bytes from 10.101.0.4: icmp_seq=3 ttl=64 time=57.850 ms
    64 bytes from 10.101.0.4: icmp_seq=4 ttl=64 time=57.956 ms

    --- 10.101.0.4 ping statistics ---
    5 packets transmitted, 5 packets received, 0.0% packet loss
    round-trip min/avg/max/stddev = 57.850/62.865/78.194/7.764 ms


## BOSH Director

Now that we have a VPC and a private network to deploy things into, we can start a BOSH Director. Here it's important
to note that I'm using "region", "network segment", and "director" interchangeably. Typically you'll have a single BOSH
Director within an environment's region, and since that Director will tag it's deployment resources with a "director"
tag, I decided to make them all synonyms. The effect is twofold:

 * when you see a "director" name (whether it's in the context of BOSH or not) it refers to where resources are
   provisioned
 * you can consistently use a "director" tag (BOSH or not) to identify where something is deployed which makes AWS
   resource management much simpler (and AWS Billing reports by tag much more valuable).

Back to getting BOSH deployed though. First, we'll create some additional BOSH-specific, region-specific infrastructure
(specifically, security groups for the director and agents)...

    $ ( cd bosh && cloque infra:put )
    ...snip...
    > waiting...CREATE_IN_PROGRESS...............CREATE_COMPLETE...done

 > Here I start using the `bosh` directory. I put Director-related configuration in the `bosh` deployment. Individual
 > BOSH deployments get their own directory.

Once the security groups are available, we can create the BOSH Director. The `boshdirector:*` commands deal with the
Director tasks (i.e. they don't depend on a specific deployment). To get started, the `boshdirector:inception:start`
command takes care of provisioning the inception instance (it takes a few minutes to get everything installed and
configured)...

    $ cloque boshdirector:inception:start \
      --security-group $( cloque --deployment=core infra:get '.TrustedPeerSecurityGroupId' ) \
      --security-group $( cloque --deployment=core infra:get '.PublicGlobalEgressSecurityGroupId' ) \
      $(cloque infra:get '.SubnetZ0PublicId') \
      t2.micro
    > finding instance...missing
      > instance-id -> i-f84169f3
      > tagging director -> acme-dev-aws-usw2
      > tagging deployment -> cloque/inception
      > tagging Name -> main
    > waiting for instance...pending.........running...done
    > waiting for ssh.......done
    > installing...
    ...snip...
    > uploading compiled/self...
    ...snip...
    > uploading global/private...
    ...snip...

 > You'll notice the `cloque --deployment=core infra:get` usage to to load the security groups. The `--deployment`
 > option is an alternative to running `cd ../core` before the command. Another alternative would be to use the
 > `CLOQUE_DEPLOYMENT` environment variable. Whatever the case, `cloque` is intelligent and flexible about figuring out
 > where it should be working from.

Before continuing, there's still a manual process of finding the correct stemcell. If we were in `us-east-1`, we could
use the "light-bosh" stemcell (which is really just an alias to a pre-compiled AMI that Cloud Foundry publishes).
Unfortunately, we need to take the slower route of compiling our own AMI for `us-west-2`. To do this, we need to lookup
the latest stemcell URL from the [published artifacts][15], then we pass that URL to the next command...

    $ cloque boshdirector:inception:provision \
      https://s3.amazonaws.com/bosh-jenkins-artifacts/bosh-stemcell/aws/bosh-stemcell-2710-aws-xen-ubuntu-trusty-go_agent.tgz
    > finding instance...found
      > instance-id -> i-f84169f3
    > deploying...
    WARNING! Your target has been changed to `https://10.101.16.8:25555'!
    Deployment set to '/home/ubuntu/cloque/self/bosh/bosh.yml'

    Verifying stemcell...
    File exists and readable                                     OK
    Verifying tarball...
    Read tarball                                                 OK
    Manifest exists                                              OK
    Stemcell image file                                          OK
    Stemcell properties                                          OK

    Stemcell info
    -------------
    Name:    bosh-aws-xen-ubuntu-trusty-go_agent
    Version: 2710

      Started deploy micro bosh
      Started deploy micro bosh > Unpacking stemcell. Done (00:00:18)
      Started deploy micro bosh > Uploading stemcell. Done (00:05:16)
      Started deploy micro bosh > Creating VM from ami-8fe7a1bf. Done (00:00:19)
      Started deploy micro bosh > Waiting for the agent. Done (00:01:19)
      Started deploy micro bosh > Updating persistent disk
      Started deploy micro bosh > Create disk. Done (00:00:02)
      Started deploy micro bosh > Mount disk. Done (00:00:09)
         Done deploy micro bosh > Updating persistent disk (00:00:19)
      Started deploy micro bosh > Stopping agent services. Done (00:00:01)
      Started deploy micro bosh > Applying micro BOSH spec. Done (00:00:21)
      Started deploy micro bosh > Starting agent services. Done (00:00:01)
      Started deploy micro bosh > Waiting for the director. Done (00:00:19)
         Done deploy micro bosh (00:08:13)
    Deployed `bosh/bosh.yml' to `https://10.101.16.8:25555', took 00:08:13 to complete
    > fetching bosh-deployments.yml...
    receiving file list ... 
    1 file to consider
    bosh-deployments.yml
            1025 100% 1000.98kB/s    0:00:00 (xfer#1, to-check=0/1)

    sent 38 bytes  received 723 bytes  101.47 bytes/sec
    total size is 1025  speedup is 1.35
    > tagging...done

 > The `:start` command took care of pushing the compiled manifest, but this `:provision` command is responsible for
 > pushing everything to the director and, once complete, downloading the resulting configuration locally. I created
 > these two commands because they were a common task and the manual, iterative process was getting tiresome. It also
 > helps unify both the intitial provisioning vs upgrade process *and* deploying from AMI vs TGZ. Instead of ~12 manual
 > steps spread out over ~30 minutes, I only need to intervene at three points (including instance termination).

Once the provisioning step is complete, I can login and talk to BOSH...

    # default username/password is admin/admin
    $ bosh target https://10.101.16.8:25555
    $ bosh status
    Config
                 /Users/dpb587/cloque-acme-dev/aws-usw2/.bosh_config

    Director
      Name       acme-dev-aws-usw2
      URL        https://10.101.16.8:25555
      Version    1.2710.0 (00000000)
      User       admin
      UUID       f38d685c-9a72-4fc0-bc84-558979cc80bf
      CPI        aws
      dns        enabled (domain_name: microbosh)
      compiled_package_cache disabled
      snapshots  disabled

    Deployment
      not set

Since BOSH Director is successfully running, it's safe to terminate the inception instance. Whenever there's a new BOSH
version I want to deploy, I can just rerun the two `start` and `provision` commands (with an updated stemcell URL)
and it will take care of upgrading it.


### More on Stemcells

While inception was deploying the BOSH Director, it ended up making a stemcell that I can reuse for our BOSH
deployments. Unfortunately, the Director doesn't know about it. The following command takes care of publishing it...

    $ cloque boshutil:create-bosh-lite-stemcell-from-ami \
      https://s3.amazonaws.com/bosh-jenkins-artifacts/bosh-stemcell/aws/light-bosh-stemcell-2710-aws-xen-ubuntu-trusty-go_agent.tgz \
      ami-8fe7a1bf
    Uploaded Stemcell: https://example-cloque-acme-dev.s3.amazonaws.com/bosh-stemcell/aws/us-west-2/light-bosh-stemcell-2710-aws-xen-ubuntu-trusty-go_agent.tgz

 > The command uses the URL (the light-bosh stemcell of the same version from the [artifacts][15] page) as a template
 > and patches in the correct metadata for the local region. It then takes care of uploading it to the environment's S3
 > bucket and to the Director so it's immediately usable.

Another task I frequently need to do is convert the standard stemcells (which only support the PV virtualization) into
HVM stemcells that I can use with AWS's newer instance types. This next command takes care of all those steps
and, once complete, there will be a new `*-hvm` stemcell ready for use on the Director.

    $ cloque boshutil:convert-pv-stemcell-to-hvm \
      https://example-cloque-acme-dev.s3.amazonaws.com/bosh-stemcell/aws/us-west-2/light-bosh-stemcell-2710-aws-xen-ubuntu-trusty-go_agent.tgz \
      ami-d13845e1 \
      $( cloque --deployment=core infra:get '.SubnetZ0PrivateId , .TrustedPeerSecurityGroupId' )
    Created AMI: ami-f3e3a5c3
    Uploaded Stemcell: https://example-cloque-acme-dev.s3.amazonaws.com/bosh-stemcell/aws/us-west-2/light-bosh-stemcell-2710-aws-xen-ubuntu-trusty-go_agent-hvm.tgz

 > The command needs the light-bosh TGZ and AMI for the existing PV stemcell as well as a subnet and security group for
 > it to provision the conversion instances in.


## BOSH Deployment

Now that the BOSH Director is running, I can deploy something interesting onto it. Let's use [logearch][2] as an
example. First I'll need to clone the repository...

    $ git clone https://github.com/logsearch/logsearch-boshrelease.git ~/logsearch-boshrelease
    $ cd ~/logsearch-boshrelease

Since I've changed directories away from our environment, `cloque` will no longer know where to find its environment
information. To help, I'll use a `.env` file...

    $ ( \
        echo 'export CLOQUE_BASEDIR=~/cloque-acme-dev' \
        ; echo 'export CLOQUE_DIRECTOR=aws-usw2' \
        ; echo 'export CLOQUE_DEPLOYMENT=logsearch' \
      ) > .env

 > I mentioned before that `cloque` uses the current working directory, environment variables, and command options to
 > figure out where to look for things. If it's still missing information, it will check and load a `.env` file from
 > the current directory as a last resort. This is normally only useful during development where I already use `.env`
 > for other project-specific BASH `alias`es and variables.

Now I can upload the release...

    $ cloque boshdirector:releases:put releases/logsearch-latest.yml

 > Since releases are Director-specific and unrelated to a particular deployment, It uses the `boshdirector:*`
 > namespace.

The example has the configuration files for infrastructure (EIP and security groups) and BOSH (deploy manifest), but
I still need to generate a certificate locally...

    $ openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
      -keyout ~/cloque-acme-dev/aws-usw2/ssl.key \
      -out ~/cloque-acme-dev/aws-usw2/ssl.crt

 > Having a directory per deployment helps keep everything scoped and organized when there are additional artifacts.
 > The templating nature of `cloque` allows the files to be embedded into its own deployment manifest, but also other
 > deployment manifests. With the example of logsearch, this means I don't need to copy and paste the `ssl.crt` into
 > other deployments, just embed it using a relative path (embeds are always relative to the config file - something
 > BOSH ERBs struggle with): `{% raw %}{{ env.embed('../logsearch/ssl.crt') }}{% endraw %}`.

Once uploaded, I can use the `infra:put` and mirrored `bosh:put` command to push the infrastructure and BOSH
deployment (`-n` meaning non-interactive, just like with `bosh`)...

    $ cloque infra:put
    ...snip...
    > waiting...CREATE_IN_PROGRESS.....................CREATE_COMPLETE...done

    $ cloque -n bosh:put
    Getting deployment properties from director...
    ...snip...
    Deployed `bosh.yml' to `acme-dev-aws-usw2'

Once complete, I can see the [elasticsearch][19] service running...

    $ wget -qO- '10.101.17.26'
    {
      "status" : 200,
      "name" : "elasticsearch/0",
      "version" : {
        "number" : "1.2.1",
        "build_hash" : "6c95b759f9e7ef0f8e17f77d850da43ce8a4b364",
        "build_timestamp" : "2014-06-03T15:02:52Z",
        "build_snapshot" : false,
        "lucene_version" : "4.8"
      },
      "tagline" : "You Know, for Search"
    }

And I can see the ingestor listening on its EIP:

    $ echo 'QUIT' | openssl s_client -showcerts -connect $( cloque infra:get '.Z0IngestorEipId' ):5614
    CONNECTED(00000003)

And I can SSH into the instance...

    $ cloque bosh:ssh
    ...snip...
    bosh_j51114xze@c989cf2f-91e4-407e-a7d7-bdc03ef79511:~$ 

 > The `bosh:ssh` command is a little more intelligent than `bosh ssh`. It will peek at the manifest to know if there's
 > only a single job running, in which case the job/index argument becomes meaningless. Additionally, it always will
 > use a default `sudo` password of `c1oudc0w` (avoiding the interactive delay and prompt that `bosh ssh` requires).


## Package Development

When I need to create a new package, I started using a convention where I'd add the origin URL where I found a
blob/file. This provides me with more of an audit over time, but also allows me to automate a `spec` file which looks
like:

    ---
    name: "nginx"
    files:
      # http://nginx.org/download/nginx-1.7.2.tar.gz
      - "nginx-blobs/nginx-1.7.2.tar.gz"
      # ftp://ftp.csx.cam.ac.uk/pub/software/programming/pcre/pcre-8.35.tar.gz
      - "nginx-blobs/pcre-8.35.tar.gz"
      # https://www.openssl.org/source/openssl-1.0.1h.tar.gz
      - "nginx-blobs/openssl-1.0.1h.tar.gz"
      ...snip...

Into a series of `wget`s with the `boshutil:package-downloads` command...

    $ cloque boshutil:package-downloads nginx
    mkdir -p 'blobs/nginx-blobs'
    [ -f 'blobs/nginx-blobs/nginx-1.7.2.tar.gz' ] || wget -O 'blobs/nginx-blobs/nginx-1.7.2.tar.gz' 'http://nginx.org/download/nginx-1.7.2.tar.gz'
    [ -f 'blobs/nginx-blobs/pcre-8.35.tar.gz' ] || wget -O 'blobs/nginx-blobs/pcre-8.35.tar.gz' 'ftp://ftp.csx.cam.ac.uk/pub/software/programming/pcre/pcre-8.35.tar.gz'
    [ -f 'blobs/nginx-blobs/openssl-1.0.1h.tar.gz' ] || wget -O 'blobs/nginx-blobs/openssl-1.0.1h.tar.gz' 'https://www.openssl.org/source/openssl-1.0.1h.tar.gz'
    ...snip...

 > I was tired of having to manually download files, `bosh add blob` them with the correct parameters and then having
 > to manually delete the originals. This lets me completely avoid that step and ensures I'm using the files I expect.
 > Whenever a blob is an internal file or `src`, I just take care of it manually like before.

When I'm working on a `packaging` script I use [Docker][22] images to emulate the build environment. Since 99% of my
build issues come from `configure` arguments and environment variables, this is normally sufficient. This also lets me
iteratively debug my packaging scripts as opposed to the slow, guess and check method of re-releasing and deploying the
whole thing to BOSH to test fixes. The `boshutil:package-docker-build` command helps me here...

    $ cloque boshutil:package-docker-build ubuntu:trusty nginx
    > compile/packaging...done
    > compile/nginx-blobs/nginx-1.7.2.tar.gz...done
    > compile/nginx-blobs/pcre-8.35.tar.gz...done
    > compile/nginx-blobs/openssl-1.0.1h.tar.gz...done
    ...snip...
    Sending build context to Docker daemon 7.571 MB
    Sending build context to Docker daemon 
    Step 0 : FROM ubuntu:trusty
     ---> ba5877dc9bec
    Step 1 : RUN apt-get update && apt-get -y install build-essential cmake m4 unzip wget
    ...snip...
    root@347c1d4ca07b:/var/vcap/data/compile/nginx# 

 > This command mirrors the BOSH environment by using the `spec` file to add the referenced blobs, uploads the
 > packaging script, configures the `BOSH_COMPILE_TARGET` and `BOSH_INSTALL_TARGET` variables, creates the directories,
 > and switches to the compile directory, ready for me to type `./packaging` or paste commands iteratively. It also has
 > the `--import-package` and `--export-package` options to import/dump the resulting `/var/vcap/packages/{name}`
 > directory to support dependencies.


## Snaphots

One easy feature that BOSH has is snapshotting to get a full backup of its persistent disks. You can run its `take
snapshot` command for a particular job or for an entire deployment. Or, if "dirty" snapshots are okay, the Director can
schedule them automatically. To manage all those snapshots, I created a few commands. The first command takes care of
snapshots that the BOSH Director creates of itself...

    $ cloque boshdirector:snapshots:cleanup-self 3d
    snap-4219f4fb -> 2014-09-13T06:01:14+00:00 -> deleted
    snap-2e6588e4 -> 2014-09-13T06:03:55+00:00 -> deleted
    snap-1acd90d3 -> 2014-09-13T06:06:36+00:00 -> deleted
    snap-618c7da9 -> 2014-09-14T06:01:15+00:00 -> retained
    snap-dce22315 -> 2014-09-14T06:03:55+00:00 -> retained
    snap-a9e81a60 -> 2014-09-14T06:06:35+00:00 -> retained
    snap-d35ea51a -> 2014-09-15T06:01:18+00:00 -> retained
    snap-3742b88e -> 2014-09-15T06:03:58+00:00 -> retained
    snap-0b8b40c2 -> 2014-09-15T06:06:38+00:00 -> retained
    snap-ea16dfd3 -> 2014-09-16T06:01:18+00:00 -> retained
    snap-913df459 -> 2014-09-16T06:03:58+00:00 -> retained
    snap-82d5fc4b -> 2014-09-16T06:06:38+00:00 -> retained

 > This command is simplistic and trims all snapshots earlier than a given period (in this case three days). I got very
 > tired and forgetful about regularly cleaning up snapshots from the AWS Console. It communicates directly with the
 > AWS API since the `bosh` command doesn't seem to enumerate them.

The command for individual deployment snapshots is a bit more intelligent. It allows writing logic which, when passed a
given snapshot, determines whether it should be retained or deleted. For example...

    $ cloque boshdirector:snapshots:cleanup
    ...snip...
    snap-7837f7d4 -> 2014-08-01T07:01:30+00:00 -> dirty -> retained
    snap-62cca4de -> 2014-08-04T07:00:28+00:00 -> dirty -> retained
    snap-bdd29512 -> 2014-08-04T22:51:57+00:00 -> clean -> retained
    snap-4dd5a3e1 -> 2014-08-04T23:46:23+00:00 -> clean -> retained
    snap-2bb7c784 -> 2014-08-11T07:00:46+00:00 -> dirty -> retained
    snap-5239b7fc -> 2014-08-18T07:00:40+00:00 -> dirty -> retained
    snap-cf6fcb6e -> 2014-08-25T07:00:39+00:00 -> dirty -> retained
    snap-9d00103c -> 2014-08-28T13:34:39+00:00 -> clean -> retained
    snap-9d80103d -> 2014-09-01T07:00:43+00:00 -> dirty -> retained
    snap-79c18cda -> 2014-09-08T07:00:44+00:00 -> dirty -> retained
    snap-87f47a24 -> 2014-09-09T07:00:57+00:00 -> dirty -> deleted
    snap-5fec87fc -> 2014-09-10T07:00:55+00:00 -> dirty -> retained
    snap-bdfeda1e -> 2014-09-11T07:00:58+00:00 -> dirty -> retained
    snap-246b6987 -> 2014-09-12T07:00:54+00:00 -> dirty -> retained
    snap-c234d870 -> 2014-09-13T07:00:43+00:00 -> dirty -> retained
    snap-28ed128a -> 2014-09-14T07:00:55+00:00 -> dirty -> retained
    snap-ef6ac34d -> 2014-09-15T07:00:55+00:00 -> dirty -> retained
    snap-72c156d3 -> 2014-09-16T07:00:42+00:00 -> dirty -> retained

 > The command looks for a deployment-specific file which receives information about the snapshot (ID, date,
 > clean/dirty) and returns `true` to cleanup/delete or `false` to retain. This allows me to create some very custom
 > retention policies for individual deployments, depending on their requirements. In this example, clean snapshots are
 > kept 3 months, Mondays are kept for 6 months, first of month is kept indefinitely, everything else kept for 1 week.


## Revitalizing

In the past I've typically used local VMs with [VirtualBox][23] or [VMWare Fusion][24] for personal development.
Unfortunately they always seemed to drift from production servers, which made things inconvenient, at best. With BOSH,
it became trivial for me to start/stop deployments and guarantee they have a known environment. When my VMs were local
I always had scripts which would pull down backups, restore them, and clean up data for development. With `cloque` I've
been using a `revitalize` concept which allows me to restore data from snapshots or run arbitrary commands. For
example, I can add the following to my database job to restore data from a slave's most recent snapshot...

    jobs:
      - name: "mysql"
        ...snip...
        cloque.revitalize:
          - method: "snapshot_copy"
            director: "example-acme-aws-usw2"
            deployment: "wordpress-demo-hotcopy"
            job: "mysql"
          - method: "script"
            script: "{{ env.embed('revitalize.sh') }}"

 > The `snapshot_copy` method takes care of finding the most recent snapshot with the given parameters and would copy
 > the data onto the local `/var/vcap/store` directory (trashing anything it replaces). The `script` method allows an
 > arbitrary script to run, in this case, one that resets the MySQL users/passwords and cleans data for development
 > purposes.

Whenever I want to reload my dev deployment with more recent production data (or after I've sufficiently polluted my
dev data), I can just run the `bosh:revitalize` task...

    $ cloque bosh:revitalize
    > mysql/0
      > finding 10.101.17.41...
        > instance-id -> i-fe0e23f3
        > availability-zone -> us-west-2w
      > stopping services...
        > waiting...............done
      > snapshot_copy
        > finding snapshot...
          > snapshot-id -> snap-3867159a
          > start-time -> 2014-09-16T06:58:31.000Z
        > creating volume...
          > volume-id -> vol-edc5bfe9
          > waiting...creating...available...done
        > attaching volume...
          > waiting...in-use...done
        > mounting volume...
        > transferring data...
          > removing mysql...done
          > restoring mysql...done
        > unmounting volume...
        > detaching volume...
          > waiting...in-use......available...done
        > destroying volume...
      > script...
      > starting services...
    ...snip...

 > This also makes it easy for me to condense services which run on multiple machines in production onto a single
 > machine for development by restoring from multiple snapshots (as long as the services `store` directories are
 > properly named).


## Configuration Transformations

I mentioned earlier that configuration files are templates. In addition to basic templating capabilities, I added some
transformation options. Transformations allow a processor to receive the current state of the configuration, do some
magic to it, and return a new configuration. The easiest example of this is with logging - I want to centralize all my
log messages and [`collectd`][26] measurements. Here I'll use [logsearch-shipper-boshrelease][25], but regardless of
how it's done, it typically requires adding a new release to your deployment, adding the job template to every job, and
adding the correct properties. When you have multiple deployments, this becomes tedious and this is where a
transformation shines. The transform could take care of the following:

 * adding the `logsearch` properties (SSL key, `bosh_director` field to messages, EIP lookup for the ingestor)
 * add the `logsearch-shipper` release to the deployment
 * add the `logsearch-shipper` job template to every job

And raw code for that transform could go in `aws-usw2/logsearch/shipper-transform.php`:

    <?php return function ($config, array $options, array $params) {
        // add our required properties
        $config['properties']['logsearch'] = [
            'logs' => [
                '_defaults' => implode("\n", [
                    '---',
                    'files:',
                    '  "**/*.log":',
                    '    fields:',
                    '      type: "unknown"',
                    '      bosh_director: "' . $params['network_name'] . '-' . $params['director_name'] . '"',
                ]),
                'server' => $params['env']['self/infrastructure/logsearch']['Z0IngestorEipId'] . ':5614',
                'ssl_ca_certificate' => $params['env']->embed(__DIR__ . '/ssl.crt'),
            ],
            'metrics' => [
                'frequency' => 60,
            ],
        ];

        // add the template job to all jobs
        foreach ($config['jobs'] as &$job) {
            $job['templates'][] = [
                'release' => 'logsearch-shipper',
                'name' => 'logsearch-shipper',
            ];
        }

        // add the release, if it's not explicitly using a version
        if (!in_array('logsearch-shipper', array_map(function ($a) { return $a['name']; }, $config['releases']))) {
            $config['releases'][] = [
                'name' => 'logsearch-shipper',
                'version' => '1',
            ];
        }

        return $config;
    };

And then whenever I want a deployment to forward its logs with `logsearch-shipper`, I only need to add the following to
the root level of my `bosh.yml` deployment manifest...

    _transformers:
      - path: "../logsearch/shipper-transform.php"

 > This approach helps me keep my deployment manifests concise. Rather than clutter up my definitions with ancillary
 > configuration and sidekick jobs, they remain focused on the services they're actually providing.


## Tagging

Since starting with BOSH, I've used AWS tags more heavily. I consistently use the `director` tag to represent the
`{network_name}-{region_name}` (e.g. `acme-dev-aws-usw2`) and the `deployment` tag to represent the logical set of
services (regardless of whether BOSH is managing them or not). I made another command which can enumerate relevant
resources and ensure they have the expected tags:

    $ cloque utility:tag-resources
    > reviewing us-west-2...
      > acme-dev-aws-usw2/bosh/microbosh -> i-298fb0c6
        > /dev/xvda -> vol-d46fa79b
          > adding director -> acme-dev-aws-usw2
          > adding deployment -> microbosh
          > adding Name -> microbosh/0/xvda
        > /dev/sdb -> vol-8b6c46c6
          > adding director -> acme-dev-aws-usw2
          > adding deployment -> microbosh
          > adding Name -> microbosh/0/sdb
        > /dev/sdf -> vol-8a6d46c6
          > adding director -> acme-dev-aws-usw2
          > adding deployment -> microbosh
          > adding Name -> microbosh/0/sdf
      > acme-dev-aws-usw2/logsearch/main/0 -> i-46be80b9
        > /dev/sda -> vol-fa4e57b5
          > adding director -> acme-dev-aws-usw2
          > adding deployment -> logsearch
          > adding Name -> main/0/sda
        > /dev/sdf -> vol-73e0ce3e
      > acme-dev-aws-usw2/infra/core/z1/gateway -> i-8d60f6a2
        > /dev/sda1 -> vol-7b5b7838

 > I added this command because I wanted to be sure my volumes were all accurately tagged. This helps me when using the
 > AWS Console, but it also provides more detail in the AWS Billing Reports when the `director` and `deployment` tags
 > are included for detailed billing.


## Conclusion

BOSH is far from perfect, in my mind, but with a little help it is enabling me to be more productive and effective 
than other tools I've tried in the areas which are most important to me.


 [1]: http://docs.cloudfoundry.org/bosh/
 [2]: https://github.com/logsearch/logsearch-boshrelease
 [3]: https://github.com/dpb587/cloque
 [4]: http://aws.amazon.com/
 [5]: http://aws.amazon.com/cloudformation/
 [6]: http://www.terraform.io/
 [7]: https://github.com/dpb587/cloque/blob/master/share/
 [8]: http://openvpn.net/
 [9]: https://github.com/dpb587/cloque/blob/master/share/local-core-infrastructure.yml
 [10]: http://twig.sensiolabs.org/
 [11]: http://console.aws.amazon.com/
 [12]: http://stedolan.github.io/jq/
 [13]: https://code.google.com/p/tunnelblick/
 [14]: https://gist.githubusercontent.com/dpb587/c0427635b3316584e12e/raw/183ccda6c504fac02754b79b5a5b267848a70025/transfer-ami.sh
 [15]: http://bosh_artifacts.cfapps.io/
 [16]: https://github.com/cloudfoundry/bosh/tree/master/bosh_cli_plugin_micro
 [18]: https://github.com/dpb587/cloque/blob/master/share/example-multi/network.yml
 [19]: http://www.elasticsearch.org/
 [20]: /blog/2014/02/28/distributed-docker-containers.html#the-alternatives
 [21]: https://github.com/cloudfoundry-incubator/spiff
 [22]: https://www.docker.com/
 [23]: https://www.virtualbox.org/
 [24]: http://www.vmware.com/products/fusion
 [25]: https://github.com/logsearch/logsearch-shipper-boshrelease/
 [26]: http://collectd.org/
 [27]: https://github.com/dpb587/cloque/blob/master/share/example-multi/global/core/infrastructure.json
 [28]: https://github.com/dpb587/cloque/blob/master/share/example-multi/aws-usw2/core/infrastructure.json
