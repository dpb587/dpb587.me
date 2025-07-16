---
description: Reusing facts from build scripts.
params:
    nav:
        tag:
            ant: true
            facter: true
publishDate: "2013-02-19"
title: Using Facter in Ant Scripts
---

After using [puppet][1] for a while I have become use to some of the facts that [facter][2] automatically provides. When
working with [ant][3] build scripts, I started wishing I didn't have to generate similar facts myself through various
`exec` calls.

# One Fact {#one-fact}

Instead of fragile lookups like...

```xml
<exec executable="/bin/bash" outputproperty="lookup.eth0">
    <arg value="-c" />
    <arg value="/sbin/ifconfig eth0 | grep 'inet addr' | awk -F: '{print $2}' | awk '{print $1}'" />
</exec>
```

I can simplify it with...

```xml
<exec executable="/usr/bin/facter" outputproperty="lookup.eth0">
    <arg value="ipaddress_eth0" />
</exec>
```


# In Bulk {#in-bulk}

Or I can load all facts with...

```xml
<tempfile property="tmp.facter.properties" deleteonexit="true" />
<exec executable="/bin/bash" output="${tmp.facter.properties}" failonerror="true">
    <arg value="-c" />
    <arg value="/usr/bin/facter -p | /bin/sed -e 's/ => /=/'" />
</exec>
<property file="${tmp.facter.properties}" prefix="facter" />
```

And reference a fact in my task...

```xml
<exec executable="${basedir}/bin/configure-env">
    <arg value="--set-listen" />
    <arg value="${facter.ipaddress_eth0}" />
</exec>
```

# Summary {#summary}

So now it's much easier to reference environment information from property files (via interpolation), make targets more
conditional, and, of course, within actual tasks.


 [1]: https://puppetlabs.com/puppet/what-is-puppet/
 [2]: https://puppetlabs.com/puppet/related-projects/facter/
 [3]: http://ant.apache.org/
