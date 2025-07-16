---
description: Locally and remotely stopping workers without interrupting jobs.
params:
    nav:
        tag:
            deploy: true
            gearman: true
            pcntl: true
            php: true
publishDate: "2013-01-14"
title: Terminating Gearman Workers in PHP
---

I use [Gearman][1] as a queue/job server. An application gives it a job to do, and Gearman passes the job along to a
worker that can finish it. Handling both synchronous and asynchronous tasks, the workers can be running anywhere -- the
same server as Gearman, a server across the country, or even a workstation at a local office.

This makes things a bit complicated when it comes time to push out software or configuration changes to workers. When
controlling workers locally, PHP's [gearman module][2] doesn't have a built-in way to terminate a worker without
possibly interrupting a running job. And by design, Gearman cannot broadcast a job to every worker, nor send a generic
job to a specific worker. I wanted a way where I could:

 * ask a worker to stop in the middle of its task (standard `SIGINT`)
 * ask a worker to stop after its current task
 * remotely terminate a worker
 * remotely terminate all workers

Even after doing a bit of [research][10] [and][10] [reading][12] [posts][13] there didn't seem to be a fully agreeable,
developed solution. So, I took an afternoon to figure things out, with the working result ending up in a
[gist]({{< appendix-ref "2013-01-14-terminating-gearman-workers-in-php/" >}}) and some of the background below.


# Graceful Termination {#graceful-termination}

For the first part, it was simply a matter of handling a `SIGTERM` signal with PHP's [pcntl module][3] and setting a
termination flag. The main worker loop could then check the flag every time it finished a job and cleanly exit. The
Gearman library complicated things a bit though because while it's waiting for a job, none of the signals are
acknowledged. The workaround was to use its non-blocking alternative. Although it still seemed to do some blocking, it
was at least a configurable duration. Abbreviated,
[worker.php]({{< appendix-ref "2013-01-14-terminating-gearman-workers-in-php/worker.php" >}}) looks like:

```php
declare(ticks = 1);

$terminate = false;

pcntl_signal(SIGTERM, function () use (&$terminate) { $terminate = true; });

$worker = new GearmanWorker(); 
$worker->addOptions(GEARMAN_WORKER_NON_BLOCKING); 
$worker->setTimeout(2500);
$worker->addServer();
$worker->addFunction(...);

while ((!$terminate) && ($worker->work())) {
    $worker->wait();
}
```

When sent a `SIGTERM` while running a job, it would wait to finish before exiting:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    (php worker.php test1 &)
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [15:45:33] READY test1 (25244)
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    php queue.php sleep 20
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [15:45:37] ASLEEP test1
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    kill -s TERM 25244
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [15:45:39] SIGTERM test1
    [15:45:57] AWAKE test1
    [15:45:57] EXIT test1
    ```

  {{< /terminal-output >}}

{{< /terminal >}}

# Remote Termination {#remote-termination}

Sometimes it's easier to remotely terminate workers when they need new code or configuration (and allowing a process
manager to restart them). Since Gearman doesn't support sending a job to every single worker, an alternative is to have
a terminate function for every worker (as mentioned in [this][5] response). Assuming every worker has a unique
identifier, this becomes trivial:

```php
$worker->addFunction(
    '_worker_' . $context['id'],
    function (GearmanJob $job) {
        if ('terminate' == $job->workload()) {
            posix_kill(getmypid(), SIGTERM);
        }
    }
);
```

From the console, it looks like:

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    (php worker.php test1 &)
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [16:19:33] READY test1 (25372)
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    php queue.php _worker_test1 terminate
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [16:19:38] SIGTERM test1
    [16:19:38] EXIT test1
    ```

  {{< /terminal-output >}}

{{< /terminal >}}


# Batch Remote Termination {#batch-remote-termination}

So now I can remotely terminate workers as needed. However, during deploys it's much more common to ask all the workers
to restart. Using Gearman's [protocol][4] to find running workers I can distribute the termination job and then wait
until all workers have received it. The result was
[`terminate.php`]({{< appendix-ref "2013-01-14-terminating-gearman-workers-in-php/terminate.php" >}}) working something
like.

{{< terminal >}}

  {{< terminal-input >}}

    ```bash
    (php worker.php test1 &) ; (php worker.php test2 &) ; (php worker.php test3 &) ; (php worker.php test4 &)
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [16:37:55] READY test1 (25479)
    [16:37:55] READY test3 (25483)
    [16:37:55] READY test2 (25481)
    [16:37:55] READY test4 (25485)
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    php queue.php sleep 4 ; php queue.php sleep 8 ; php queue.php sleep 16
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [16:37:57] ASLEEP test2
    [16:37:57] ASLEEP test3
    [16:37:57] ASLEEP test4
    ```

  {{< /terminal-output >}}

  {{< terminal-input >}}

    ```bash
    php terminate.php
    ```

  {{< /terminal-input >}}

  {{< terminal-output >}}

    ```
    [16:37:59] UP test4
    [16:37:59] UP test3
    [16:37:59] UP test2
    [16:37:59] UP test1
    [16:37:59] SIGTERM test1
    [16:37:59] EXIT test1
    [16:37:59] DOWN test1
    [16:38:01] AWAKE test2
    [16:38:01] SIGTERM test2
    [16:38:01] EXIT test2
    [16:38:01] DOWN test2
    [16:38:05] AWAKE test3
    [16:38:05] SIGTERM test3
    [16:38:05] EXIT test3
    [16:38:05] DOWN test3
    [16:38:08] waiting for: test4
    [16:38:13] AWAKE test4
    [16:38:13] SIGTERM test4
    [16:38:13] EXIT test4
    [16:38:13] DOWN test4
    ```

  {{< /terminal-output >}}

{{< /terminal >}}


# Summary {#summary}

The result is an extra bit of code, but it makes automating tasks, especially around deploys, much easier. This really
just demonstrates one method of creating an internal workers API - termination is just one possibility. Other more
complex possibilities could be self-performing updates, lighter config reloads (instead of full restarts), or
dynamically registering/unregistering functions depending on application load.


 [1]: http://gearman.org/
 [2]: http://php.net/manual/en/book.gearman.php
 [3]: http://php.net/manual/en/book.pcntl.php
 [4]: http://gearman.org/protocol
 [5]: http://stackoverflow.com/questions/7663922/gearman-using-php-possible-to-send-job-message-to-all-workers/7664139#7664139

 [10]: http://gearman.org/php_reference
 [11]: https://groups.google.com/forum/?fromgroups=#!topic/gearman/ST6Ikw7__kY
 [12]: http://stackoverflow.com/questions/2270323/stopping-gearman-workers-nicely
 [13]: http://stackoverflow.com/questions/7663922/gearman-using-php-possible-to-send-job-message-to-all-workers
