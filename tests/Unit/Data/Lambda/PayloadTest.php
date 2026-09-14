<?php

namespace Tests\Unit\Data\Lambda;

use App\Data\Lambda\Payload;
use Aws\Result;
use GuzzleHttp\Psr7\Utils;
use Tests\TestCase;

class PayloadTest extends TestCase
{
    public function test_make_populates_fields_from_result_payload(): void
    {
        $body = json_encode([
            'content_key' => 'content/123.html',
            'screenshot_key' => 'screenshots/123.png',
            'url' => 'https://example.com/article',
        ]);

        $result = new Result([
            'Payload' => Utils::streamFor($body),
        ]);

        $payload = Payload::make($result);

        $this->assertSame('content/123.html', $payload->contentKey);
        $this->assertSame('screenshots/123.png', $payload->screenshotKey);
        $this->assertSame('https://example.com/article', $payload->url);
    }

    public function test_make_defaults_missing_fields_to_empty_strings(): void
    {
        $result = new Result([
            'Payload' => Utils::streamFor(json_encode([])),
        ]);

        $payload = Payload::make($result);

        $this->assertSame('', $payload->contentKey);
        $this->assertSame('', $payload->screenshotKey);
        $this->assertSame('', $payload->url);
    }
}
