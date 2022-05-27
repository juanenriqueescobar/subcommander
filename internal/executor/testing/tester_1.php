<?php

echo "".PHP_EOL;
echo '{"level":100}'.PHP_EOL;

echo '{"level":100,"message":"step1"}'.PHP_EOL;
echo '{"message":"step2"}'.PHP_EOL;
$msg=file_get_contents("php://stdin");
$msg=json_decode($msg, true);
$metadata=$msg["metadata"];
$body=json_decode($msg["body"], true);
echo '{"level":200,"message":"step3","metadata":'.json_encode($metadata).'}'.PHP_EOL;
echo '{"level":300,"message":"step4","body":'.json_encode($body).'}'.PHP_EOL;
echo '{"level":400,"message":"step5"}'.PHP_EOL;
echo 'step6'.PHP_EOL;
