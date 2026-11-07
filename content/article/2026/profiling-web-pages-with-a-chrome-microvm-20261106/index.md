---
description: Building a unikernel-style VM image and Firecracker worker pool.
publishDate: 2026-11-06T00:00:00Z
title: Profiling Web Pages with a Chrome microVM
---

{{< callout >}}{{< markdown >}}

This is part of my ongoing experiments with AI. I asked it to document the architecture and evolution of a specific feature I have been working on. Aside from this callout, the article is generated (via [Cursor](https://cursor.com); notably [Grok 4.7](https://x.ai/news/grok-4-7), [GLM 5.3](https://z.ai/blog/glm-5.3)) using the prompts below &mdash; Sincerely, Human

{{< /markdown >}}
   
{{< details summary="Prompt 1" >}}{{< markdown >}}

Create a new blog post article in `dpb587.me/content/article` which documents the process of building a unikernel-style image for chrome browser to profile web pages. The main content of the article should be the general concept of (a) building a unikernal, (b) maintaining a pool of firecracker vms from snapshot, (c) using static ips in different network namespaces to maintain connectivity, (d) tunneling traffic/rpc from browsers via control process for proxying.

The latest implementation and best practices can be found in `structuredtext`, but the broader history can be found in `monorepo`. Review all the commits related to the `webio` components to identify (a) changes being made, and (b) inferring what bugs were being fixed based on the commit message or line diff. The history of changes offers valuable insights into the decisions we made and is a valuable perspective for the article. The webio component is used by a few different layers, so consider how it is integrated with other components as it reflects interesting dependencies worth commenting on for the article.

Consider how to target or focus the article to avoid it being a rambling "tech article" and to ensure it offers a valuable, meaningful, quality content that has value.

{{< /markdown >}}{{< /details >}}

{{< details summary="Prompt 2" >}}{{< markdown >}}

   Rewrite the post with a more traditional style of writing. Currently, it relies on many short sentences focused on little tidbit facts, and it is difficult for a human to read. Maintain the existing storyline, but improve the written form. The author should have a straightforward prose personality which enjoys the technical aspects but does not dwell on any one detail for too long.

   Additionally:

   1. Add a mermaid-style diagram near the top which gives a big picture summary of the layers (e.g. cloud, vms (worker, egress, api), routes, bridges, chrome, controller).
   2. Ensure the title and description are more concise and represent on the content.
   3. Ensure the overall content has a clear audience and purpose.

{{< /markdown >}}{{< /details >}}

{{< /callout >}}

For a while now I have been building the part of a system that profiles web pages: it takes a URL, loads it in a real browser, and hands back a screenshot, the DOM the browser actually built, and a log of the requests behind it. If you work on anything that renders untrusted pages on demand — archiving, indexing, screenshotting, testing — you have probably run into the same pressures I did, and this post walks through the design that came out of them: a small unikernel-style image built around headless Chrome, a pool of Firecracker microVMs restored from a snapshot of it, and a network that makes the browser's only route run back through the worker that owns it.

None of the pressures is exotic on its own. The browser must be fresh for every page, because a long-lived one carries the last page's cookies, cache, and DNS answers into the next job. It must already be running when the request arrives, because a cold boot spends the whole fetch on startup. And everything it does must pass through a proxy the worker controls, because that is where the crawler's rules live — robots.txt, DNS policy, egress. A Firecracker snapshot satisfies all three at once: boot one guest, wait for Chrome to be ready, pause it, and restore that memory for each page.

The catch is what a snapshot actually freezes: a guest that has already booted. Chrome is listening, the user-agent is already set, and `eth0` already holds `192.168.200.2/30` with a default route through `192.168.200.1`. None of that can be renegotiated after the fact — the restored guest cannot be renumbered, handed a DHCP lease, or pointed at a different proxy. Cloning a snapshot only works if every copy is allowed to keep that identity while the host translates between them, and nearly every decision in the rest of this post follows from that one constraint.

Before getting into the details, here is the shape of the whole system, from the client request down to the browser:

{{< generated-diagram lang="mermaid" dir="diagram-mermaid" caption="One fetch, from the client request through the API and queue, down to the snapshot-restored guest whose only route back out is the worker's proxy." >}}

A client asks the API for a page, the API charges a credit and publishes a job, and the worker pulls jobs off the queue and drives a guest from its pool. I call that worker webio, and the small program that runs beside the browser inside the guest is the vmcontroller. Chrome talks to the vmcontroller on loopback, the vmcontroller carries that traffic back over the bridge, and only the worker's proxy can reach the outside world.

## The Image Is Finished Before Anyone Asks

The guest is deliberately small: a read-only Ubuntu root filesystem, a headless Chromium, and a small Go program I call the controller. It is unikernel-style in the narrow sense that it boots into one program and is never administered, though what matters here is that boot is completely finished before the first request arrives. A second virtio disk, mounted at `/webio`, carries the home directory, temp space, and Chrome's user data. Init's last act is to assign the only address the guest will ever have, and then to replace itself with the controller:

```sh
ip link set lo up
ip link set eth0 up
ip addr add 192.168.200.2/30 dev eth0
ip route add default via 192.168.200.1

exec runuser -u webio -g webio -w TMPDIR,XDG_RUNTIME_DIR /usr/local/bin/webio-vmcontroller
```

A `/30` is the entire network plan: two usable addresses, one for the guest and one for the tap device that will be its gateway, and deliberately no third address for anyone else. The controller starts Chrome pointed at a proxy on loopback (`--proxy-server=http://127.0.0.1:3128`) and gives it a resolver rule — `MAP * ~NOTFOUND , EXCLUDE 127.0.0.1` — that makes every name except loopback unresolvable inside the guest, because real DNS belongs to the worker, not the browser. Then it waits to be sure Chrome is genuinely ready: polling `127.0.0.1:9223/json/version` is only the start, since that document can answer before the browser is actually usable, so the controller sleeps another eight seconds — a blunt instrument, with a comment admitting a better waiter is still wanted — and only then starts listening for the worker. The host will not snapshot the guest until it does.

Even the user-agent obeys the same rule at a smaller scale. The image build runs Chromium once and writes its default agent string into the rootfs, while the crawler's identity arrives on the kernel command line as a base64-encoded `webio.ua=` argument, which the controller patches together and passes as `--user-agent` before Chrome ever starts. A kernel argument is an appealing channel for this: it needs no socket back to the host, and it is present from the first instruction of userspace. Once the snapshot is taken, the string is just an argument of a running process, out of reach of any restore — so anything that has to differ per page has to be applied from outside, over the debugging port, after the memory is running again.

The snapshot itself comes from a single cold boot that the worker calls the seed. It waits until the controller accepts a Yamux connection on port 2222, closes that connection, pauses the VM, and writes the result to `vm.snapshot` and `vm.mem`. From then on, every guest the system runs is that one pause, resumed.

## A Pool of Clean Restores

At startup the worker builds the seed and then fills a pool by restoring it: `poolSize` guests are loaded with `ResumeVM` set, and one extra network slot is kept free so that the replacement for a guest handed to a fetch starts restoring immediately instead of waiting for the fetch to finish. When the fetch does finish, the guest is destroyed — release closes the control session, stops the VMM, deletes the temporary directory, and returns the slot — and the next restore of that slot loads the same snapshot. Cleanliness comes from never reusing a browser rather than from wiping one down; the cookies, cache, and mounted filesystem inside the image are whatever the seed had, and they end when the VM stops.

The one piece of writable host state needs a little care. The snapshot remembers the host path of the seed's scratch ext4, so before Firecracker starts, the worker bind-mounts a fresh ext4 over that path inside a private mount namespace; without that, two guests would be writing to one host file. The bind mount isolates writers, but it is worth being precise about what it does not do — the guest still sees the filesystem it already has mounted. A later page sees a clean state because the snapshot is what gets loaded, and because the VM that did the writing is gone.

Guests are also checked on their way out of the pool. If a guest's Yamux session has closed, the worker tries to open another; a failed dial, or a debugging port that no longer accepts a TCP connection, discards the guest, and a restore takes its place. Three unhealthy guests in a row and the acquire gives up. Inside the guest the controller would happily start Chrome again, but the host does not try to nurse a guest back — the whole point of the pool is that a clean replacement is seconds away. A snapshot can resume into a guest whose browser has already died, and the pool's job is to notice.

For a while the pool also sat behind a semaphore, and the two disagreed about what capacity meant: the semaphore counted in-flight fetches while the pool counted guests, so a fetch could be admitted with no guest free, and a guest could sit idle while the semaphore was full. Once the pool learned to refuse an acquire, the semaphore was deleted and the slot count became the only concurrency limit. The queue honors the same number — RabbitMQ delivers messages only while fewer than `poolSize` are unacked, and a message is acked only after the fetch returns. A busy worker nacks and the broker requeues; a cancelled one simply never acks, and the broker requeues it anyway.

## The Same Address, in a Different Namespace

The guest's address cannot be unique, because it is baked into the snapshot; a per-guest subnet would mean configuring the guest after restore, which is exactly what the snapshot forbids. So the host repeats the subnet instead. Each pool slot is a network namespace, `webio-vm-N`, and inside every one of them `tap0` is `192.168.200.1/30`. The same prefix can recur because each namespace has its own routing table, and since the Firecracker process for a slot is started with `ip netns exec`, the tap device it opens is that namespace's tap — which is precisely the gateway the guest's default route already points at.

The worker, meanwhile, runs in the host namespace, where `192.168.200.2` names nothing in particular. What the worker dials instead is a veth pair the guest never configures. The namespace end of the pair holds `192.168.(101+N).2/30`, and iptables rules inside the namespace DNAT ports 2222 and 9222, when destined to that address, onto the guest at `192.168.200.2` — port 2222 reaching the controller, port 9222 reaching a small forwarder the controller runs to expose Chrome's `127.0.0.1:9223`. Masquerade on the way out of `tap0` rewrites the source back to the gateway address:

```
worker
   |
   |  dial 192.168.(101+N).2 :2222 and :9222
   v
+-------------------------------- webio-vm-N ------------------+
|  eth1  192.168.(101+N).2                                     |
|     |  DNAT those two ports                                  |
|     v                                                        |
|  guest eth0  192.168.200.2/30  -->  tap0  192.168.200.1/30   |
+--------------------------------------------------------------+
```

The guest never learns which slot it landed in. The seed is built in a namespace of the same shape — `webio-seed`, with a veth at `192.168.99.2` — so the snapshot is taken on a network that looks exactly like every slot it will later be restored into. A small aside I find telling: the same file contains a helper that computes a distinct MAC address per slot, and nothing calls it. The seed's NIC has one fixed MAC, and the restore path passes no network configuration at all; it just loads the snapshot. A slot's only obligation is to provide a device named `tap0` at the gateway address the guest already has installed.

The earlier version of this system ran Chrome directly on the host, in the same network namespace as the proxy, driven in-process by chromedp. It had no address to keep stable, but it also had no boundary: cookies, resolver state, and proxy sessions were all state of one machine, and the proxy had already been split into separate handlers in an attempt to keep one page's traffic out of the next. Moving Chrome into a microVM is what turned the address into a constant. The repeated `/30` is what makes one snapshot usable in every slot — a static IP here is not an optimization but the enabling trick.

## The Browser's Only Route Is Back to the Worker

Chrome cannot resolve names, and its only proxy is on loopback, so nothing it does can leave the guest except through the controller. When Chrome opens a connection to `127.0.0.1:3128`, the controller accepts it, opens a stream on the Yamux session the worker dialed on port 2222, and bridges the bytes in both directions:

```go
stream, err := session.OpenStream()
// ...
bridgeConns(chromeConn, stream)
```

On the worker side, accepting those streams and serving the proxy on them is the entire data path. The proxy is an HTTP intercepting proxy living in the worker process, and it is where the crawler's policy lives: robots.txt evaluation, DNS checks, and egress all happen there. The worker's own initial fetch of the page goes out through a SOCKS gateway, while requests coming back from Chrome pass through the proxy into a Squid cache that forwards its misses to the same gateway. If the controller has no live session with the worker, it closes the proxy connection, and Chrome has nowhere else to send it.

The debugging connection runs the other direction and is deliberately not on Yamux. chromedp wants an HTTP endpoint, so the worker opens `http://192.168.(101+N).2:9222` and talks through the forwarder, a plain TCP bridge to Chrome's `9223`. Putting CDP on the multiplexer would have meant teaching an existing client a private transport for no real gain, and the proxy connections want the opposite shape anyway: one pipe per TCP connection, which is exactly what a Yamux stream is. So the many anonymous connections ride a multiplexer, and the one client that wants a URL gets a port.

The two paths meet at a header. The debugging session pauses each request, appends `Webio-{session}-Request-Id`, and continues it; the host proxy looks for that header, dispatches to the session it names, and strips it before the request goes upstream. A request without the header is answered with a 503 — even Chrome's own background telemetry, the update checks and account probes, is quietly logged and refused the same way — so a request the controller did not pause does not leave the machine.

I did write a gRPC service for this guest once, an `Execute` and `Ping` surface whose generated code is still sitting in the tree. It was the wrong abstraction: the protobuf described a guest that ran fetches and returned blobs, but fetching and archiving happen on the host, and the guest turned out to be a splice. The controller accepts a connection, serves Yamux, and copies bytes.

The nicest property of this arrangement is how mobile the policy is. Egress moved from an internal cache beside the worker to a SOCKS gateway with Squid in front of it, and the guest image did not change at all. Archive replay — re-rendering a previously captured page — became just another proxy handler registered for that one session on the same tunnel, and only a worker that was given those dependencies will accept such a task. Through both changes the snapshot kept the same Chrome, the same loopback proxy, and the same address.

## What the Rest of the System May Depend On

Zooming back out: [Structured Text](https://structuredtext.app/) is the product that asks for pages. Its API charges a credit, publishes a job to the queue, and later reads a tar archive containing the screenshot, the DOM, and the resource log. It has no view of slots or snapshots and never acquires a guest. At one point it kept a session row in SQLite and watched it, but the row just repeated state the worker already owned, and once the worker owned execution outright, it went away. What the API can rely on is modest and durable: the archive shows up, and a queue message stays unacked until the fetch returns.

The worker is the only component that knows about the snapshot, the namespaces, and the request header; the browser knows only a loopback proxy and a debugging port. That division is the cut the snapshot forces. Robots rules, DNS checks, and egress are changes to the host proxy, and guests already restored stay valid; Chrome's flags, the user-agent file in the rootfs, and the address in init are changes to the memory the next seed will freeze — the identity the snapshot is not allowed to renegotiate.
