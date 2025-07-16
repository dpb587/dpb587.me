<?php

declare(ticks = 1);

$context = array(
    // ought to be unique within the server's workforce
    'id' => $_SERVER['argv'][1],
    'pid' => getmypid(),
    'terminate' => false,
);

pcntl_signal(
    SIGTERM,
    function () use (&$context) {
        fwrite(STDOUT, sprintf('[%s] SIGTERM %s', date('H:i:s'), $context['id']) . "\n");

        $context['terminate'] = true;
    }
);

$worker = new GearmanWorker(); 
$worker->addOptions(GEARMAN_WORKER_NON_BLOCKING);
// maximum time that gearman will block from userspace code
$worker->setTimeout(2500);
$worker->addServer();

$worker->addFunction(
    'sleep',
    function ($job) use ($context) {
        fwrite(STDOUT, sprintf('[%s] ASLEEP %s', date('H:i:s'), $context['id']) . "\n");

        for ($i = 0; $i < $job->workload(); $i ++) {
            // signals interrupt sleep, so loop for one second instead
            sleep(1);
        }

        fwrite(STDOUT, sprintf('[%s] AWAKE %s', date('H:i:s'), $context['id']) . "\n");
    }
);

$worker->addFunction(
    '_worker_' . $context['id'],
    function (GearmanJob $job) use ($context) {
        switch ($job->workload()) {
            case 'terminate':
                posix_kill($context['pid'], SIGTERM); # exec(sprintf('/bin/kill -s TERM %d', $context['pid']));

                break;
        }
    }
);

fwrite(STDOUT, sprintf('[%s] READY %s (%d)', date('H:i:s'), $context['id'], $context['pid']) . "\n");

// work on jobs as they're available
while (
    (!$context['terminate'])
    && (
        $worker->work()
        || (GEARMAN_IO_WAIT == $worker->returnCode())
        || (GEARMAN_NO_JOBS == $worker->returnCode())
    )
) {
    if (GEARMAN_SUCCESS == $worker->returnCode()) {
        continue;
    }

    $worker->wait();
}

$worker->unregisterAll();

fwrite(STDOUT, sprintf('[%s] EXIT %s', date('H:i:s'), $context['id']) . "\n");