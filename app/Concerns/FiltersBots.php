<?php

namespace App\Concerns;

use App\Enums\BotType;
use Illuminate\Http\Request;

trait FiltersBots
{
    /** @return array{bots: array, filters: array, typeOptions: array} */
    protected function paginatedBots(Request $request): array
    {
        $filters = [
            'query' => $request->str('query', '')->trim()->value(),
            'type' => $request->str('type', '')->value(),
            'status' => $request->str('status', '')->value(),
            'mode' => $request->str('mode', '')->value(),
            'sort' => $request->str('sort', 'created_at_desc')->value(),
            'perPage' => $request->integer('perPage', 10),
        ];

        if (! in_array($filters['sort'], [
            'created_at_desc',
            'created_at_asc',
            'query_asc',
            'query_desc',
            'type_asc',
            'type_desc',
            'enabled_desc',
            'enabled_asc',
            'headless_desc',
            'headless_asc',
            'processed_at_desc',
            'processed_at_asc',
            'updated_at_desc',
            'updated_at_asc',
        ], true)) {
            $filters['sort'] = 'created_at_desc';
        }

        if (! in_array($filters['perPage'], [10, 25, 50], true)) {
            $filters['perPage'] = 10;
        }

        $bots = $request->user()->bots()
            ->when(
                $filters['query'] !== '',
                fn ($query) => $query->whereRaw('LOWER(query) LIKE ?', ['%'.strtolower($filters['query']).'%']),
            )
            ->when(
                BotType::tryFrom($filters['type']) !== null,
                fn ($query) => $query->type($filters['type']),
            )
            ->when(
                $filters['status'] !== '',
                fn ($query) => $query->enabled($filters['status'] === 'enabled'),
            )
            ->when(
                $filters['mode'] !== '',
                fn ($query) => $query->headless($filters['mode'] === 'headless'),
            );

        match ($filters['sort']) {
            'created_at_asc' => $bots->oldest(),
            'query_asc' => $bots->orderBy('query'),
            'query_desc' => $bots->orderByDesc('query'),
            'type_asc' => $bots->orderBy('type'),
            'type_desc' => $bots->orderByDesc('type'),
            'enabled_desc' => $bots->orderByDesc('enabled'),
            'enabled_asc' => $bots->orderBy('enabled'),
            'headless_desc' => $bots->orderByDesc('headless'),
            'headless_asc' => $bots->orderBy('headless'),
            'processed_at_desc' => $bots->orderByDesc('last_run_at'),
            'processed_at_asc' => $bots->orderBy('last_run_at'),
            'updated_at_desc' => $bots->orderByDesc('updated_at'),
            'updated_at_asc' => $bots->orderBy('updated_at'),
            default => $bots->latest(),
        };

        return [
            'bots' => $bots->paginate($filters['perPage'])->withQueryString(),
            'filters' => $filters,
            'typeOptions' => BotType::options(),
        ];
    }
}
