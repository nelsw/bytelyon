<?php

namespace App\Contracts;

use App\Enums\BotType;
use App\Models\Bot;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

interface Processable
{
    function process(array $args): bool;

    function type(): BotType;

    function headless(): bool;
}
