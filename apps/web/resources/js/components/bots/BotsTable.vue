<script setup lang="ts">
import { Link, router } from '@inertiajs/vue3';
import { ArrowDown, ArrowUp, ArrowUpDown } from '@lucide/vue';
import { computed } from 'vue';
import DeleteBotButton from '@/components/bots/DeleteBotButton.vue';
import EditButton from '@/components/EditButton.vue';
import { Badge } from '@/components/ui/badge';
import BotTypeIcon from '@/components/ui/BotTypeIcon.vue';
import { Button } from '@/components/ui/button';
import { formatDate, formatFromNow } from '@/lib/utils';
import type { BotFilters, BotRow, PaginatedBots } from '@/types/bots';

const props = withDefaults(
    defineProps<{
        bots: PaginatedBots;
        filters: BotFilters;
        basePath?: string;
    }>(),
    {
        basePath: '/bots',
    },
);

type SortField =
    | 'query'
    | 'type'
    | 'enabled'
    | 'headless'
    | 'processed_at'
    | 'created_at'
    | 'updated_at';
type SortDirection = 'asc' | 'desc';

const hasActiveFilters = computed(
    () =>
        props.filters.query !== '' ||
        props.filters.type !== '' ||
        props.filters.status !== '' ||
        props.filters.mode !== '' ||
        props.filters.sort !== 'created_at_desc' ||
        props.filters.perPage !== 10,
);

const pageLinks = computed(() =>
    props.bots.links.filter(
        (link) =>
            link.label !== '&laquo; Previous' && link.label !== 'Next &raquo;',
    ),
);

const defaultSortDirection: Record<SortField, SortDirection> = {
    query: 'asc',
    type: 'asc',
    enabled: 'desc',
    headless: 'desc',
    processed_at: 'desc',
    created_at: 'desc',
    updated_at: 'desc',
};

function currentSortDirection(field: SortField): SortDirection | null {
    if (props.filters.sort === `${field}_asc`) {
        return 'asc';
    }

    if (props.filters.sort === `${field}_desc`) {
        return 'desc';
    }

    return null;
}

function nextSort(field: SortField): string {
    const currentDirection = currentSortDirection(field);

    if (!currentDirection) {
        return `${field}_${defaultSortDirection[field]}`;
    }

    return `${field}_${currentDirection === 'asc' ? 'desc' : 'asc'}`;
}

function botsIndexHref(overrides: Partial<BotFilters> = {}): string {
    const merged: BotFilters = {
        ...props.filters,
        ...overrides,
    };

    const params = new URLSearchParams();

    if (merged.query !== '') {
        params.set('query', merged.query);
    }

    if (merged.type !== '') {
        params.set('type', merged.type);
    }

    if (merged.status !== '') {
        params.set('status', merged.status);
    }

    if (merged.mode !== '') {
        params.set('mode', merged.mode);
    }

    if (merged.sort !== 'created_at_desc') {
        params.set('sort', merged.sort);
    }

    if (merged.perPage !== 10) {
        params.set('perPage', String(merged.perPage));
    }

    const query = params.toString();

    return query === '' ? props.basePath : `${props.basePath}?${query}`;
}

function sortHref(field: SortField): string {
    return botsIndexHref({ sort: nextSort(field) });
}

function onRowClick(row: BotRow): void {
    if (row.type === 'news') {
        router.visit(`/bots/${row.id}/articles`);
    } else if (row.type === 'search') {
        router.visit(`/serps/${row.childId}`);
    } else {
        router.visit(`/sitemaps/${row.childId}`);
    }
}
</script>

<template>
    <div
        v-if="bots.data.length === 0 && !hasActiveFilters"
        class="rounded-lg border border-dashed p-8 text-center"
    >
        <h2 class="text-lg font-semibold">No bots yet</h2>
        <p class="mt-2 text-sm text-muted-foreground">
            Create your first bot to start tracking queries and schedules.
        </p>
        <Button as-child class="mt-4">
            <Link href="/bots/create">Create your first bot</Link>
        </Button>
    </div>

    <div
        v-else-if="bots.data.length === 0"
        class="rounded-lg border border-dashed p-8 text-center"
    >
        <h2 class="text-lg font-semibold">No bots match these filters</h2>
        <p class="mt-2 text-sm text-muted-foreground">
            Try adjusting your search or clearing one or more filters.
        </p>
        <Button as-child class="mt-4" variant="outline">
            <Link :href="basePath">Clear filters</Link>
        </Button>
    </div>

    <template v-else>
        <div class="overflow-x-auto rounded-lg border">
            <table class="w-full min-w-225 text-left text-sm">
                <thead class="bg-muted/50 text-muted-foreground">
                    <tr class="border-b">
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('query')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Query
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('query') === 'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('query') === 'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('type')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Type
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('type') === 'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('type') === 'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 font-medium">Pages</th>
                        <th class="px-4 py-3 font-medium">Frequency</th>
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('enabled')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Status
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('enabled') ===
                                        'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('enabled') ===
                                        'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('headless')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Mode
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('headless') ===
                                        'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('headless') ===
                                        'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('processed_at')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Processed
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('processed_at') ===
                                        'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('processed_at') ===
                                        'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('updated_at')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Updated
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('created_at') ===
                                        'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('created_at') ===
                                        'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 font-medium">
                            <Link
                                :href="sortHref('created_at')"
                                class="inline-flex items-center gap-1 hover:text-foreground"
                            >
                                Created
                                <ArrowUp
                                    v-if="
                                        currentSortDirection('created_at') ===
                                        'asc'
                                    "
                                    class="size-4"
                                />
                                <ArrowDown
                                    v-else-if="
                                        currentSortDirection('created_at') ===
                                        'desc'
                                    "
                                    class="size-4"
                                />
                                <ArrowUpDown v-else class="size-4 opacity-60" />
                            </Link>
                        </th>
                        <th class="px-4 py-3 text-right font-medium">
                            <span class="m-3">Actions</span>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr
                        v-for="bot in bots.data"
                        :key="bot.id"
                        class="cursor-pointer border-b transition-colors last:border-b-0 hover:bg-muted/30"
                        @click="onRowClick(bot)"
                    >
                        <td
                            class="px-4 py-3 align-middle font-medium text-foreground"
                        >
                            <div class="max-w-sm truncate" :title="bot.query">
                                {{ bot.query }}
                            </div>
                        </td>
                        <td class="px-4 py-3 align-middle">
                            <Badge class="capitalize" variant="outline">
                                <BotTypeIcon
                                    :bot-type="bot.type"
                                    style="color: var(--info)"
                                />
                                <span>{{ bot.type }}</span>
                            </Badge>
                        </td>
                        <td class="px-4 py-3 align-middle">
                            <span
                                class="w-16 text-right font-mono text-muted-foreground tabular-nums"
                            >
                                {{ bot.pageCount }}
                            </span>
                        </td>
                        <td
                            class="px-4 py-3 align-middle text-muted-foreground"
                        >
                            <span class="capitalize">{{ bot.frequency }}</span>
                        </td>
                        <td class="px-4 py-3 align-middle">
                            <Badge
                                :variant="bot.enabled ? 'success' : 'secondary'"
                            >
                                {{ bot.enabled ? 'Enabled' : 'Disabled' }}
                            </Badge>
                        </td>
                        <td
                            class="px-4 py-3 align-middle text-muted-foreground"
                        >
                            {{ bot.headless ? 'Headless' : 'Browser' }}
                        </td>
                        <td
                            class="px-4 py-3 align-middle text-muted-foreground"
                        >
                            {{ formatFromNow(bot.processedAt) }}
                        </td>
                        <td
                            class="px-4 py-3 align-middle text-muted-foreground"
                        >
                            {{ formatDate(bot.updatedAt) }}
                        </td>
                        <td
                            class="px-4 py-3 align-middle text-muted-foreground"
                        >
                            {{ formatDate(bot.createdAt) }}
                        </td>

                        <td class="px-4 py-3 align-middle" @click.stop>
                            <div
                                class="flex flex-wrap items-center justify-end gap-2"
                            >
                                <EditButton :href="`/bots/${bot.id}/edit`" />
                                <DeleteBotButton
                                    :bot-id="bot.id"
                                    :bot-query="bot.query"
                                />
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div
            class="flex flex-col gap-4 pt-4 sm:flex-row sm:items-center sm:justify-between"
        >
            <p class="text-sm text-muted-foreground">
                Showing {{ bots.from ?? 0 }}–{{ bots.to ?? 0 }} of
                {{ bots.total }} bots
            </p>

            <div class="flex flex-wrap items-center gap-2">
                <Button
                    as-child
                    variant="outline"
                    size="sm"
                    v-if="bots.prev_page_url"
                >
                    <Link :href="bots.prev_page_url">Previous</Link>
                </Button>
                <Button v-else variant="outline" size="sm" disabled>
                    Previous
                </Button>

                <Button
                    v-for="link in pageLinks"
                    :key="`${link.label}-${link.url}`"
                    as-child
                    :variant="link.active ? 'default' : 'outline'"
                    size="sm"
                    :disabled="link.url === null"
                >
                    <Link v-if="link.url" :href="link.url">{{
                        link.label
                    }}</Link>
                    <span v-else>{{ link.label }}</span>
                </Button>

                <Button
                    as-child
                    variant="outline"
                    size="sm"
                    v-if="bots.next_page_url"
                >
                    <Link :href="bots.next_page_url">Next</Link>
                </Button>
                <Button v-else variant="outline" size="sm" disabled>
                    Next
                </Button>
            </div>
        </div>
    </template>
</template>
