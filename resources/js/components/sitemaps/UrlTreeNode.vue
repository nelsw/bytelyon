<script setup lang="ts">
import { ChevronRight, ExternalLink, Image, Info } from '@lucide/vue';
import { inject, ref, watch } from 'vue';
import type { Ref } from 'vue';

type Page = {
    id: number;
    url: string;
    title: string;
    meta: Record<string, unknown>;
    screenshotUrl: string | null;
};

type UrlLink = {
    url: string;
    page: Page | null;
};

type UrlTreeNode = {
    id: string;
    label: string;
    count: number;
    links: UrlLink[];
    children: UrlTreeNode[];
};

type ExpandSignal = { open: boolean; version: number };

const props = defineProps<{
    node: UrlTreeNode;
}>();

const open = ref(props.node.id === props.node.label);
const openScreenshots = ref(new Set<string>());
const openMeta = ref(new Set<string>());

const expandSignal = inject<Ref<ExpandSignal> | null>(
    'sitemapExpandSignal',
    null,
);

if (expandSignal) {
    if (expandSignal.value.version > 0) {
        open.value = expandSignal.value.open;
    }

    watch(
        () => expandSignal.value.version,
        () => (open.value = expandSignal.value.open),
    );
}

function toggle(set: Set<string>, url: string): void {
    if (set.has(url)) {
        set.delete(url);
    } else {
        set.add(url);
    }
}

function formatMetaValue(value: unknown): string {
    if (value === null || value === undefined) {
        return '—';
    }

    if (Array.isArray(value)) {
        return value.map((entry) => formatMetaValue(entry)).join(', ');
    }

    if (typeof value === 'object') {
        return JSON.stringify(value);
    }

    return String(value);
}

function metaEntries(page: Page): { key: string; value: string }[] {
    return Object.entries(page.meta ?? {})
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, value]) => ({ key, value: formatMetaValue(value) }));
}
</script>

<template>
    <li>
        <div
            class="group flex h-7 items-center gap-1 rounded px-1 hover:bg-muted/60"
        >
            <button
                type="button"
                class="flex min-w-0 flex-1 items-center gap-1.5 text-left"
                :class="node.children.length ? '' : 'cursor-default'"
                @click="node.children.length && (open = !open)"
            >
                <ChevronRight
                    class="size-3.5 shrink-0 text-muted-foreground transition-transform duration-150"
                    :class="[
                        open ? 'rotate-90' : '',
                        node.children.length ? '' : 'invisible',
                    ]"
                />

                <span class="shrink-0 font-medium">{{ node.label }}</span>

                <span
                    v-if="node.children.length"
                    class="shrink-0 rounded-full bg-muted px-1.5 text-[11px] leading-4 text-muted-foreground"
                >
                    {{ node.count }}
                </span>

                <span
                    v-if="node.links.length === 1 && node.links[0].page"
                    class="truncate text-xs text-muted-foreground"
                    :title="node.links[0].page.title"
                >
                    {{ node.links[0].page.title }}
                </span>
            </button>

            <template v-if="node.links.length === 1">
                <div
                    class="flex shrink-0 items-center gap-0.5 text-muted-foreground"
                >
                    <button
                        v-if="node.links[0].page?.screenshotUrl"
                        type="button"
                        class="rounded p-1 hover:bg-background hover:text-foreground"
                        :class="
                            openScreenshots.has(node.links[0].url)
                                ? 'text-primary'
                                : ''
                        "
                        title="Toggle screenshot"
                        @click="toggle(openScreenshots, node.links[0].url)"
                    >
                        <Image class="size-3.5" />
                    </button>
                    <button
                        v-if="node.links[0].page"
                        type="button"
                        class="rounded p-1 hover:bg-background hover:text-foreground"
                        :class="
                            openMeta.has(node.links[0].url)
                                ? 'text-primary'
                                : ''
                        "
                        title="Toggle meta"
                        @click="toggle(openMeta, node.links[0].url)"
                    >
                        <Info class="size-3.5" />
                    </button>
                    <span
                        v-else
                        class="px-1 text-[11px]"
                        title="This URL has not been crawled yet."
                    >
                        not crawled
                    </span>
                    <a
                        :href="node.links[0].url"
                        target="_blank"
                        rel="noreferrer"
                        class="rounded p-1 hover:bg-background hover:text-foreground"
                        :title="node.links[0].url"
                    >
                        <ExternalLink class="size-3.5" />
                    </a>
                </div>
            </template>
        </div>

        <ul
            v-if="
                node.links.length > 1 ||
                node.links.some(
                    (link) =>
                        openScreenshots.has(link.url) || openMeta.has(link.url),
                )
            "
            class="ml-[11px] border-l pl-3 text-xs"
        >
            <li v-for="link in node.links" :key="link.url">
                <div
                    v-if="node.links.length > 1"
                    class="flex h-6 items-center gap-1 rounded px-1 hover:bg-muted/60"
                >
                    <a
                        :href="link.url"
                        target="_blank"
                        rel="noreferrer"
                        class="min-w-0 flex-1 truncate text-primary hover:underline"
                        :title="link.url"
                    >
                        {{ link.url }}
                    </a>
                    <span
                        v-if="link.page"
                        class="max-w-[40%] truncate text-muted-foreground"
                        :title="link.page.title"
                    >
                        {{ link.page.title }}
                    </span>
                    <button
                        v-if="link.page?.screenshotUrl"
                        type="button"
                        class="rounded p-1 text-muted-foreground hover:bg-background hover:text-foreground"
                        :class="
                            openScreenshots.has(link.url) ? 'text-primary' : ''
                        "
                        title="Toggle screenshot"
                        @click="toggle(openScreenshots, link.url)"
                    >
                        <Image class="size-3.5" />
                    </button>
                    <button
                        v-if="link.page"
                        type="button"
                        class="rounded p-1 text-muted-foreground hover:bg-background hover:text-foreground"
                        :class="openMeta.has(link.url) ? 'text-primary' : ''"
                        title="Toggle meta"
                        @click="toggle(openMeta, link.url)"
                    >
                        <Info class="size-3.5" />
                    </button>
                    <span v-else class="px-1 text-[11px] text-muted-foreground">
                        not crawled
                    </span>
                </div>

                <div
                    v-if="
                        link.page &&
                        (openScreenshots.has(link.url) ||
                            openMeta.has(link.url))
                    "
                    class="my-1 space-y-2 rounded-md border bg-muted/30 p-2"
                >
                    <img
                        v-if="
                            link.page.screenshotUrl &&
                            openScreenshots.has(link.url)
                        "
                        :src="link.page.screenshotUrl"
                        :alt="`Screenshot of ${link.page.url}`"
                        loading="lazy"
                        class="w-full max-w-xl rounded border bg-muted"
                    />

                    <template v-if="openMeta.has(link.url)">
                        <dl
                            v-if="metaEntries(link.page).length > 0"
                            class="grid grid-cols-[fit-content(14rem)_minmax(0,1fr)] divide-y divide-border/60"
                        >
                            <div
                                v-for="entry in metaEntries(link.page)"
                                :key="entry.key"
                                class="col-span-2 grid grid-cols-subgrid gap-x-4 py-1"
                            >
                                <dt
                                    class="font-medium [overflow-wrap:anywhere] text-muted-foreground"
                                    :title="entry.key"
                                >
                                    {{ entry.key }}
                                </dt>
                                <dd class="[overflow-wrap:anywhere]">
                                    {{ entry.value }}
                                </dd>
                            </div>
                        </dl>
                        <p v-else class="text-muted-foreground">
                            No meta recorded for this page.
                        </p>
                    </template>
                </div>
            </li>
        </ul>

        <ul v-if="open && node.children.length" class="ml-[11px] border-l pl-2">
            <UrlTreeNode
                v-for="child in node.children"
                :key="child.id"
                :node="child"
            />
        </ul>
    </li>
</template>
