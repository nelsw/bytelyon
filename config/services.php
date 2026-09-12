<?php

return [

    /*
    |--------------------------------------------------------------------------
    | Third Party Services
    |--------------------------------------------------------------------------
    |
    | This file is for storing the credentials for third party services such
    | as Mailgun, Postmark, AWS and more. This file provides the de facto
    | location for this type of information, allowing packages to have
    | a conventional file to locate the various service credentials.
    |
    */

    'ses' => [
        'key' => env('AWS_ACCESS_KEY_ID'),
        'secret' => env('AWS_SECRET_ACCESS_KEY'),
        'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
    ],

    'lambda' => [
        'version' => env('LAMBDA_VERSION', 'latest'),
        'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
        'credentials' => [
            'key' => env('AWS_ACCESS_KEY_ID'),
            'secret' => env('AWS_SECRET_ACCESS_KEY'),
        ],
    ],

    'proxy' => [
        'name' => env('PROXY_NAME', 'default'),
        'protocol' => env('PROXY_PROTOCOL', 'http'),
        'bypass' => env('PROXY_BYPASS'),
        'password' => env('PROXY_PASSWORD'),
        'server' => env('PROXY_SERVER'),
        'port' => env('PROXY_PORT'),
        'username' => env('PROXY_USERNAME'),
    ]
];
