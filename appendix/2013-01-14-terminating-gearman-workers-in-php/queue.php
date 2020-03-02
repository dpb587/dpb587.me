<?php

$gearman = new GearmanClient();
$gearman->addServer();
$gearman->doBackground(
    $_SERVER['argv'][1],
    isset($_SERVER['argv'][2]) ? $_SERVER['argv'][2] : ''
);