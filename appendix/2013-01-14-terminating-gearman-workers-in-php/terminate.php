<?php

// connect to gearman
if (false === $sh = fsockopen('127.0.0.1', '4730', $errno, $errstr, 10)) {
    fwrite(STDERR, sprintf('Unable to connect to gearman: %s: %s', $errno, $errstr) . "\n");

    exit(1);
}

fwrite($sh, "status\n");

$workers = array();

$gearman = new GearmanClient();
$gearman->addServer();

// find running workers and ask them to terminate
while ((!feof($sh)) && (".\n" !== $line = fgets($sh))) {
    if (preg_match('/^_worker_([^\s]+)\s+\d+\s+\d+\s+(\d+)/', $line, $match)) {
        if ($match[2]) {
            fwrite(STDOUT, sprintf('[%s] UP %s', date('H:i:s'), $match[1]) . "\n");

            $workers[$match[1]] = $gearman->doHighBackground(
                '_worker_' . $match[1],
                'terminate',
                $match[1]
            );
        }
    }
}

fclose($sh);

// callback to update the list of who is still down
$gearman->setStatusCallback(
    function (GearmanTask $task, $context) use (&$workers) {
        if (!$task->isKnown()) {
            unset($workers[$context]);
            fwrite(STDOUT, sprintf('[%s] DOWN %s', date('H:i:s'), $context) . "\n");
        }
    }
);

$loop = 0;

// poll gearman about the termination jobs
do {
    foreach ($workers as $worker => $handle) {
        $gearman->addTaskStatus($handle, $worker);
    }

    $gearman->runTasks();

    if ((0 == ++ $loop % 20) && ($workers)) {
        // remind who is still down every 10 seconds
        fwrite(STDOUT, sprintf('[%s] waiting for: %s', date('H:i:s'), implode(' ', array_keys($workers))) . "\n");
    }

    usleep(500000);
} while ($workers);