<?php

return [
    'version' => env('LAMBDA_VERSION', 'latest'),
    'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
    'credentials' => [
        'key' => env('AWS_ACCESS_KEY_ID'),
        'secret' => env('AWS_SECRET_ACCESS_KEY'),
    ],
    'proxy' => [
        'name' => env('PROXY_NAME', 'default'),
        'bypass' => env('PROXY_BYPASS'),
        'password' => env('PROXY_PASSWORD'),
        'server' => env('PROXY_SERVER'),
        'username' => env('PROXY_USERNAME'),
    ],
];
