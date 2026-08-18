<script setup lang="ts">
import { ChevronRight } from '@lucide/vue';
import { computed } from 'vue';
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from '@/components/ui/collapsible';

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

const props = defineProps<{
    node: UrlTreeNode;
}>();

const pageCount = computed(
    () => props.node.links.filter((link) => link.page !== null).length,
);

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
        <Collapsible v-slot="{ open }" :default-open="node.id === node.label">
            <CollapsibleTrigger
                class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-muted/60"
            >
                <ChevronRight
                    class="size-4 shrink-0 text-muted-foreground transition-transform duration-200"
                    :class="open ? 'rotate-90' : ''"
                />

                <span class="font-medium break-all">{{ node.label }}</span>

                <span
                    class="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
                >
                    {{ node.count }}
                </span>

                <span
                    v-if="pageCount > 0"
                    class="rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary"
                >
                    {{ pageCount }} captured
                </span>
            </CollapsibleTrigger>

            <CollapsibleContent
                class="mt-2 ml-4 space-y-3 border-l pl-4 text-sm"
            >
                <div
                    v-for="link in node.links"
                    :key="link.url"
                    class="space-y-3 rounded-lg border p-3"
                >
                    <a
                        :href="link.url"
                        class="block break-all text-primary hover:underline"
                        target="_blank"
                        rel="noreferrer"
                    >
                        {{ link.url }}
                    </a>

                    <template v-if="link.page">
                        <p class="font-medium">{{ link.page.title }}</p>

                        <img
                            v-if="link.page.screenshotUrl"
                            :src="link.page.screenshotUrl"
                            :alt="`Screenshot of ${link.page.url}`"
                            loading="lazy"
                            class="w-full max-w-2xl rounded-md border bg-muted"
                        />

                        <p v-else class="text-xs text-muted-foreground">
                            No screenshot stored for this page.
                        </p>

                        <dl
                            v-if="metaEntries(link.page).length > 0"
                            class="grid gap-x-4 gap-y-1 sm:grid-cols-[minmax(0,12rem)_1fr]"
                        >
                            <template
                                v-for="entry in metaEntries(link.page)"
                                :key="entry.key"
                            >
                                <dt
                                    class="font-medium break-all text-muted-foreground"
                                >
                                    {{ entry.key }}
                                </dt>
                                <dd class="break-words">{{ entry.value }}</dd>
                            </template>
                        </dl>

                        <p v-else class="text-xs text-muted-foreground">
                            No meta recorded for this page.
                        </p>
                    </template>

                    <p v-else class="text-xs text-muted-foreground">
                        This URL has not been crawled yet.
                    </p>
                </div>

                <ul v-if="node.children.length" class="space-y-2">
                    <UrlTreeNode
                        v-for="child in node.children"
                        :key="child.id"
                        :node="child"
                    />
                </ul>
            </CollapsibleContent>
        </Collapsible>
    </li>
</template>
