<?php

namespace App\Providers;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\ServiceProvider;
use Illuminate\Support\Str;

class MacroServiceProvider extends ServiceProvider
{
    public function boot(): void
    {
        $this->configureArr();
        $this->configureURL();
    }

    private function configureArr(): void
    {
        Arr::macro('toCamelCase', function (Model|array|null $arg) {
            if ($arg === null) {
                return [];
            }
            if ($arg instanceof Model) {
                $arg = $arg->toArray();
            }
            $ƒ = function (array $arr) use (&$ƒ): array {
                $result = [];
                foreach ($arr as $key => $val) {
                    $result[Str::camel($key)] = is_array($val)
                        ? $ƒ($val)
                        : $val;
                }
                return $result;
            };
            return $ƒ($arg);
        });
    }

    private function configureURL(): void
    {
        URL::macro('toDomain', function (?string $url) {
            if (str($url)->substrCount('.') === 1) {
                return $url;
            }
            return $url !== null
                ? parse_url($url, PHP_URL_HOST)
                    |> (fn ($x) => explode('.', (string) $x))
                    |> (fn ($x) => array_slice($x, -2))
                    |> (fn ($x) => implode('.', $x))
                : '';
        });

        URL::macro('toFavicon', function (?string $url, int $size = 64) {
            if ($url === null) {
                $url = 'bytelyon.com';
            } else {
                $url = URL::toDomain($url);
            }
            return "https://www.google.com/s2/favicons?domain=$url&sz=$size";
        });
    }
}
