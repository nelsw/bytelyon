<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import { ArrowDown, ArrowUp, ArrowUpDown, Code, Wallpaper } from '@lucide/vue';
import { computed, ref } from 'vue';
import BotDrawer from '@/components/BotDrawer.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { dashboard } from '@/routes';

type Page = {
    id: number;
    kind: string | null;
    index: number | null;
    title: string;
    url: string;
    domain: string;
    created_at: string | null;
    meta: Record<string, unknown>;
    faviconUrl: string;
    screenshotUrl: string | null;
};

type Serp = {
    id: number;
    query: string;
    data: Record<string, unknown>;
    screenshotUrl: string | null;
    pages: Page[];
    similarQueries: string[];
};

const props = defineProps<{
    serp: Serp;
}>();

type SortKey = 'kind' | 'index';
type SortDir = 'asc' | 'desc';

const sortKey = ref<SortKey>('index');
const sortDir = ref<SortDir>('asc');

function toggleSort(key: SortKey): void {
    if (sortKey.value === key) {
        sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc';
    } else {
        sortKey.value = key;
        sortDir.value = 'asc';
    }
}

function compareValues(
    left: string | number | null,
    right: string | number | null,
): number {
    if (left === null && right === null) {
        return 0;
    }

    if (left === null) {
        return 1;
    }

    if (right === null) {
        return -1;
    }

    if (typeof left === 'number' && typeof right === 'number') {
        return left - right;
    }

    return String(left).localeCompare(String(right));
}

const sortedPages = computed<Page[]>(() => {
    const direction = sortDir.value === 'asc' ? 1 : -1;

    return [...props.serp.pages].sort(
        (left, right) =>
            compareValues(left[sortKey.value], right[sortKey.value]) *
            direction,
    );
});

const screenshotDialogOpen = ref(false);
const activeScreenshotUrl = ref<string | null>(null);

function openScreenshot(url: string | null): void {
    if (url === null) {
        return;
    }

    activeScreenshotUrl.value = url;
    screenshotDialogOpen.value = true;
}

const metaDialogOpen = ref(false);
const activeMetaPage = ref<Page | null>(null);

function openMeta(page: Page): void {
    activeMetaPage.value = page;
    metaDialogOpen.value = true;
}

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: dashboard(),
            },
            {
                title: 'Searches',
                href: '/serps',
            },
            {
                title: 'Results',
                href: '/serps',
            },
        ],
    },
});

function formatDataValue(value: unknown): string {
    if (value === null || value === undefined) {
        return '—';
    }

    if (Array.isArray(value)) {
        return value.map((entry) => formatDataValue(entry)).join(', ');
    }

    if (typeof value === 'object') {
        return JSON.stringify(value);
    }

    return String(value);
}

function dataEntries(
    data: Record<string, unknown>,
): { key: string; value: string }[] {
    return Object.entries(data ?? {})
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, value]) => ({ key, value: formatDataValue(value) }));
}
</script>

<template>
    <Head :title="`Search: ${serp.query}`" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex justify-between space-y-1">
            <div>
                <h1 class="text-xl font-bold">{{ serp.query }}</h1>
            </div>
            <div>
                <Button
                    v-if="serp.screenshotUrl"
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="openScreenshot(serp.screenshotUrl)"
                    class="cursor-pointer"
                >
                    <Wallpaper class="size-4" style="color: var(--success)" />
                    SERP
                </Button>
            </div>
        </div>

        <hr />
        <div v-if="serp.similarQueries.length > 0">
            <div class="flex flex-col align-middle md:flex-row">
                <div class="flex w-full items-center text-center md:w-1/12">
                    <h3>Similar Queries</h3>
                </div>
                <div class="w-full text-center md:w-11/12">
                    <span v-for="q in serp.similarQueries" :key="q" class="p-1">
                        <BotDrawer
                            :bot="{
                                query: q,
                                type: 'search',
                                frequency: 'daily',
                            }"
                        >
                            <template #trigger>
                                <Badge
                                    class="my-2 cursor-pointer capitalize"
                                    variant="outline"
                                >
                                    {{ q }}
                                </Badge>
                            </template>
                        </BotDrawer>
                    </span>
                </div>
            </div>
            <div class="flex justify-between gap-2 space-y-1 align-middle">
                <div></div>
            </div>
        </div>

        <Card class="gap-0 py-0">
            <div
                v-if="serp.pages.length === 0"
                class="rounded-lg border border-dashed p-8 text-center"
            >
                <h2 class="text-lg font-semibold">No pages yet</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    There are no pages recorded for this search.
                </p>
            </div>

            <div v-else class="overflow-x-auto rounded-lg">
                <table class="w-full min-w-3xl text-left text-sm">
                    <thead class="bg-muted/50 text-muted-foreground">
                        <tr class="border-b">
                            <th class="px-4 py-3 font-medium">
                                <button
                                    type="button"
                                    class="inline-flex cursor-pointer items-center gap-1 hover:text-foreground"
                                    @click="toggleSort('kind')"
                                >
                                    Kind
                                    <ArrowUp
                                        v-if="
                                            sortKey === 'kind' &&
                                            sortDir === 'asc'
                                        "
                                        class="size-3.5"
                                    />
                                    <ArrowDown
                                        v-else-if="
                                            sortKey === 'kind' &&
                                            sortDir === 'desc'
                                        "
                                        class="size-3.5"
                                    />
                                    <ArrowUpDown
                                        v-else
                                        class="size-3.5 opacity-50"
                                    />
                                </button>
                            </th>
                            <th class="px-4 py-3 font-medium">
                                <button
                                    type="button"
                                    class="inline-flex cursor-pointer items-center gap-1 hover:text-foreground"
                                    @click="toggleSort('index')"
                                >
                                    Rank
                                    <ArrowUp
                                        v-if="
                                            sortKey === 'index' &&
                                            sortDir === 'asc'
                                        "
                                        class="size-3.5"
                                    />
                                    <ArrowDown
                                        v-else-if="
                                            sortKey === 'index' &&
                                            sortDir === 'desc'
                                        "
                                        class="size-3.5"
                                    />
                                    <ArrowUpDown
                                        v-else
                                        class="size-3.5 opacity-50"
                                    />
                                </button>
                            </th>
                            <th class="px-4 py-3 font-medium">Source</th>
                            <th class="px-4 py-3 font-medium">Title</th>
                            <th class="px-4 py-3 text-right font-medium">
                                <span class="mr-1">Meta</span>
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="page in sortedPages"
                            :key="page.id"
                            class="cursor-pointer border-b transition-colors last:border-b-0 hover:bg-muted/30"
                            @click="openScreenshot(page.screenshotUrl)"
                        >
                            <td class="px-4 py-3 align-middle font-medium">
                                {{ page.kind ?? '—' }}
                            </td>
                            <td class="px-4 py-3">
                                {{ page.index ?? '—' }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                <div class="flex items-center gap-2">
                                    <img
                                        :src="page.faviconUrl"
                                        :alt="page.domain"
                                        class="size-5 object-cover"
                                    />
                                    <span class="">
                                        {{ page.domain }}
                                    </span>
                                </div>
                            </td>
                            <td class="px-4 py-3 align-middle">
                                <a
                                    :href="page.url"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="hover:underline"
                                >
                                    {{ page.title }}
                                </a>
                            </td>
                            <td
                                class="px-4 py-3 text-right align-middle"
                                @click.stop
                            >
                                <div class="flex justify-end gap-2">
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="icon-sm"
                                        @click="openMeta(page)"
                                        class="cursor-pointer"
                                    >
                                        <Code
                                            class="size-4"
                                            style="color: var(--info)"
                                        />
                                        <span class="sr-only"
                                            >View page metadata</span
                                        >
                                    </Button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </Card>
    </div>

    <Dialog v-model:open="screenshotDialogOpen">
        <DialogContent
            class="max-h-[95vh] max-w-[95vw] overflow-auto p-2 sm:max-w-[95vw]"
        >
            <DialogTitle class="sr-only">
                Screenshot: {{ serp.query }}
            </DialogTitle>
            <img
                v-if="activeScreenshotUrl"
                :src="activeScreenshotUrl"
                :alt="`Screenshot of search: ${serp.query}`"
                class="w-full rounded-md"
            />
        </DialogContent>
    </Dialog>

    <Dialog v-model:open="metaDialogOpen">
        <DialogContent
            class="max-h-[90vh] max-w-[90vw] overflow-auto sm:max-w-[90vw]"
        >
            <DialogHeader>
                <DialogTitle>{{ activeMetaPage?.title ?? 'Page' }}</DialogTitle>
                <DialogDescription>
                    Metadata captured for this page.
                </DialogDescription>
            </DialogHeader>

            <dl
                v-if="
                    activeMetaPage &&
                    dataEntries(activeMetaPage.meta).length > 0
                "
                class="grid gap-x-4 gap-y-1 sm:grid-cols-[minmax(0,15rem)_1fr]"
            >
                <template
                    v-for="entry in dataEntries(activeMetaPage.meta)"
                    :key="entry.key"
                >
                    <dt class="font-medium break-all text-muted-foreground">
                        {{ entry.key }}
                    </dt>
                    <dd class="wrap-break-word">{{ entry.value }}</dd>
                </template>
            </dl>

            <p v-else class="text-xs text-muted-foreground">
                No metadata recorded for this page.
            </p>
        </DialogContent>
    </Dialog>
</template>
