---
date: 2015-11-12
title: "Tempore limites: BOSH Veneer"
description: "Experimenting with a browser frontend to working with BOSH."
tags:
- bosh
- browser
- frontend
- user interface
aliases:
- /blog/2015/11/12/tempore-limites-bosh-veneer.html
---

For all the low-level handling of things, BOSH is a good tool for system administration. But when it comes to
configuring everything, I think it leaves something to be desired for the average Joe. Opening my text editor, making
changes to the YAML, copying and pasting security groups from AWS Console, `git diff`ing to make sure I did what I
think, `git commit`ing in case things go bad, `bosh deploy`ing to make it so... it can become quite the process. For me,
I'm much more a visual person and prefer a browser-based tool. Since I've had a bit extra free-time lately, I've spent
some time experimenting on ideas to help improve my BOSH-quality-of-life.

<!--more-->


## BOSH

![Screenshot: core-login](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/core-login.png)

The `bosh` CLI can work with multiple directors and uses the `target` command to switch between instances. With a
browser-based tool, I just need to browse to the director or whatever dedicated instance I've deployed the release to.
From there, I login with my credentials as I would with `bosh login`.

While working with the project, I've been referring to it as "veneer", as in "a thin decorative covering of fine wood
applied to a coarser wood or other material."

![Screenshot: bosh-release-version](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/bosh-release-version.png)

One of the core features is to simply provide browser-based pages to view BOSH resources. For example, it's easy to see
the list of releases and details about specific release versions. This makes the release and configuration process much
more discoverable to end users. The screenshot shows details about the logsearch release, something which I deploy
alongside all deployments to collect logs and metrics.

![Screenshot: core-login](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/bosh-deployment-vm.png)

Of course, the most common BOSH resource is deployments. I can quickly pull up a specific VM to see what's installed and
how it is configured in the cloud. Since I'm using the AWS CPI, an extra link is shown on the side which links directly
to the instance in AWS Console. Further down on that page is a section which describes the persistent disk on the VM.

![Screenshot: bosh-deployment-job-disk](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/bosh-deployment-job-disk.png)

The AWS component of veneer knows the various CloudWatch metrics which are available for instances and disks. Here the
persistent disk metrics are shown, including timing, queue length, and idle time below. This allows me to quickly pull
up graphs if I'm trying to investigate an issue. If I do need to diagnose further in AWS Console, the sidebar link will
take me straight to the EBS Volume there.

![Screenshot: bosh-deployment-job](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/bosh-deployment-job.png)

I mentioned I included logsearch alongside all my deployments. Similar to veneer's AWS component, I also have a
logsearch component which advertises many internal metrics for the BOSH resources. Here, on a specific job, I can
quickly see load and memory usage over the past few hours. I can hover over the chart for specific values, or click
into the graph to change the time span, granularity, and statistical method used.


## Marketplace

![Screenshot: marketplace-home](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/marketplace-home.png)

One of the reasons I like BOSH is because I can use releases from both the open-source community, but also my own
internally built releases. The marketplace component provides that central view into the various sources where I can
pull my releases and stemcells from. For example, `theloopyewe` marketplace enumerates a private S3 bucket using a regex
to identify artifacts and their release name/version. Of course, the `boshio` one scrapes and uses the API to pull down
the public [bosh.io][4] resources.

![Screenshot: marketplace-stemcells](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/marketplace-stemcells.png)

From bosh.io, I can easily view the list of stemcells which are available. There are many more stemcells than I actually
use from a single director, so the checkmark helps me identify which one(s) I have already uploaded to the director. If
I want to see the full list of versions, I click on the name for a similar view. Versions which follow [semver][5] are
parsed to provide intelligent advice about whether deployments are up to date in their release and stemcell usage.

![Screenshot: marketplace-stemcell-upload](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/marketplace-stemcell-upload.png)

When viewing a specific stemcell version, I get a quick summary and, if it's not already installed, I have the option to
upload it to the director right on-screen. Assuming the director has internet access, I can click "Upload" where the
task will be started and I get redirected to the task detail page. The release version page is similar.

![Screenshot: marketplace-stemcell-task](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/marketplace-stemcell-task.png)

The task page automatically updates until it has completed successfully at which point it'll redirect me to the main
stemcell summary page indicating it was completed. If an error occurs, it'll show me the full error and wait for me to
diagnose and figure out a resolution myself.


## Operations

![Screenshot: ops-deployment-editor](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/ops-deployment-editor.png)

I've mentioned the BOSH, AWS, Logsearch, and Marketplace component, but the most intriguing component is Operations.
This component handles more of the management tasks, most notably, editing deployment configuration. It provides the
core forms for deployment manifests, but it also imports the forms that the CPI-specific component provides.

![Screenshot: ops-deployment-editor-resourcepool](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/ops-deployment-editor-resourcepool.png)

For example, the Cloud Properties section of the resource pool uses the AWS-specific form including Instance Type, but
also properties like Availability Zone and ELB Names below the fold. You can also see the stemcell field is
intelligently populated based on the stemcell names and versions which are installed on the director.

![Screenshot: ops-deployment-editor-job](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/ops-deployment-editor-job.png)

Editing a job is also straightforward - it references the resource and disk pools already configured in the manifest so
they're easy to select. The templates are also enumerated based on which releases are already configured in the manifest
and installed on the director. The forms also clearly indicate which properties are required vs which are optional
(since there are often more properties available than are typically needed).

![Screenshot: ops-deployment-editor-properties](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/ops-deployment-editor-properties.png)

Properties are another piece of deployments which are frequently changed. Here, properties are enumerated based on which
releases and templates are referenced in the deployment manifest. A green plus on the right indicates the property is
not currently set, while a blue pencil button shows a setting is currently set.

![Screenshot: ops-deployment-editor-property](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/ops-deployment-editor-property.png)

If I do want to change the property, a simple form comes up where I can input my YAML-friendly value. If the release's
job spec provides the metadata, the help information includes description and example information.

![Screenshot: ops-deployment-editor-pending](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/ops-deployment-editor-pending.png)

When changes have been saved, they are not immediately sent to the director. This allows multiple changes to be made and
then deployed at a coordinated time. It's important not to forget changes though, so the banner provides a reminder that
changes are pending and provides a link to compare the changes before applying them.


## Core

![Screenshot: core-repo-deployment](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/core-repo-deployment.png)

I mentioned changes are not immediately applied, and this is because they are actually written to new branch in the git
repository where everything is maintained. The git repository provides the support for versioning and merging - when
clicking the Review button, it's actually just showing an intelligent `diff` between `master` and the drafted branch.

![Screenshot: core-repo-deployment-arbitrary](https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/blog/2015-11-12-tempore-limites-bosh-veneer/core-repo-deployment-arbitrary.png)

Similarly, as a git repository it can be cloned over HTTPS from veneer for backup purposes or advanced editing and then
even pushed back. This makes veneer more of a tool which can function alongside other infrastructure tools which also
commit their configurations. For example, in the earlier photo you'll see `cloudformation.json` templates - something
which I currently manage externally yet can still reference from my deployment manifests using pre-processing
capabilities that veneer provides.


## Summary

For enterprisey-types, I've heard there's such a thing as [Ops Manager][1] which helps provide a bit of a frontend for
deploying [certain software][2] (like [Cloud Foundry][3]). I'm not quite an enterprisey-type and don't have an
enterprisey budget, but I still appreciate having shiny tools where I can point my browser to manage, monitor, and
cross-reference my technical resources.

Since my extra free-time is coming to a close as I move on to another chapter in my life, this project will sit on my
backburner. I still like the features and ideas though, so I figured I could write a post summarizing some of them. At
the very least, if you encounter a project with similar features leave a comment - I'd love to use it myself!



 [1]: http://docs.pivotal.io/pivotalcf/customizing/
 [2]: https://network.pivotal.io/
 [3]: https://www.cloudfoundry.org/
 [4]: https://bosh.io/
 [5]: http://semver.org/
