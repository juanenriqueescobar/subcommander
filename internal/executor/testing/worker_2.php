<?php

usleep(100000);
echo '{"level":100,"message":"starting"}'.PHP_EOL;

usleep(100000);
$data=file_get_contents("php://stdin");
$json=json_decode($data, true);

usleep(100000);
echo '{"level":200,"message":'.$json["order_id"].'}'.PHP_EOL;

echo '{"level":300,"message":"redis slow"}'.PHP_EOL;

usleep(100000);
echo '{"level":400,"message":"mysql not respond"}'.PHP_EOL;
