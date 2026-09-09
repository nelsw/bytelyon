<?php

namespace App\Dto;

readonly class Links
{
    private function __construct(
        public string $host,
        public array  $map,
    ){}

    public static function of(string $domain, array $values): self
    {
        $map = [];
        foreach ($values as $key => $val) {

            $val = trim($val);
            $val = explode('#', $val)[0];
            if (!str_starts_with($val, 'https://')) {
                $val = "https://$val";
            }

            if (
                empty($val) ||
                isset($map[$key]) ||
                !is_string($val) ||
                !str_starts_with($val, "https://$domain") ||
                !filter_var($val, FILTER_VALIDATE_URL) ||
                preg_match('/^(mailto|tel|sms|fax|callto|geo|javascript|about):.*/', $val)
            ) {
                continue;
            }
            $map[$key] = $val;
        }

        return new self($domain, $map);
    }
}
