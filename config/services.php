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

    'sqs' => [
        'version' => 'latest',
        'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
        'credentials' => [
            'key' => env('AWS_ACCESS_KEY_ID'),
            'secret' => env('AWS_SECRET_ACCESS_KEY'),
        ],
        'scrape_jobs_queue_url' => env('SQS_SCRAPE_JOBS_QUEUE_URL', 'MyQueue.fifo'),
    ],

    'proxy' => [
        'bypass' => env('PROXY_BYPASS'),
        'host' => env('PROXY_HOST'),
        'name' => env('PROXY_NAME', 'default'),
        'pass' => env('PROXY_PASS'),
        'port' => env('PROXY_PORT'),
        'scheme' => env('PROXY_SCHEME', 'http'),
        'username' => env('PROXY_USERNAME'),
    ],
];
