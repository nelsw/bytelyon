<?php

namespace App\Enums;

use App\Traits\HasArrays;

enum BotType: string
{
    use HasArrays;

    case News = 'news';
    case Search = 'search';
    case Sitemap = 'sitemap';
}
