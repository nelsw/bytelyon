<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import { computed } from 'vue';
import UrlTreeNode from '@/components/sitemaps/UrlTreeNode.vue';
import { Card, CardContent } from '@/components/ui/card';
import { dashboard } from '@/routes';

type Page = {
    id: number;
    url: string;
    title: string;
    meta: Record<string, unknown>;
    screenshotUrl: string | null;
};

type Sitemap = {
    id: number;
    domain: string;
    faviconUrl: string;
    urls: string[];
    pages: Page[];
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

type MutableTreeNode = {
    id: string;
    label: string;
    count: number;
    links: UrlLink[];
    children: Map<string, MutableTreeNode>;
};

const props = defineProps<{
    sitemap: Sitemap;
}>();

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: dashboard(),
            },
            {
                title: 'Sitemaps',
                href: '/sitemaps',
            },
            {
                title: 'Pages',
                href: '/sitemaps',
            },
        ],
    },
});

function toTreeNodes(nodes: Map<string, MutableTreeNode>): UrlTreeNode[] {
    return Array.from(nodes.values())
        .sort((left, right) => left.label.localeCompare(right.label))
        .map((node) => ({
            id: node.id,
            label: node.label,
            count: node.count,
            links: [...node.links],
            children: toTreeNodes(node.children),
        }));
}

const pagesByUrl = computed<Map<string, Page>>(
    () => new Map(props.sitemap.pages.map((page) => [page.url, page])),
);

const treeNodes = computed<UrlTreeNode[]>(() => {
    const roots = new Map<string, MutableTreeNode>();

    for (const rawUrl of props.sitemap.urls) {
        try {
            const parsed = new URL(rawUrl);
            const parts = [
                parsed.host,
                ...parsed.pathname
                    .split('/')
                    .filter((value) => value.length > 0)
                    .map(decodeURIComponent),
            ];

            if (parts.length === 0) {
                continue;
            }

            let currentNodes = roots;
            let currentNode: MutableTreeNode | null = null;
            let currentId = '';

            for (const part of parts) {
                currentId = currentId === '' ? part : `${currentId}/${part}`;

                const existing = currentNodes.get(part);

                if (existing) {
                    existing.count += 1;

                    currentNode = existing;
                } else {
                    const created: MutableTreeNode = {
                        id: currentId,
                        label: part,
                        count: 1,
                        links: [],
                        children: new Map<string, MutableTreeNode>(),
                    };

                    currentNodes.set(part, created);
                    currentNode = created;
                }

                currentNodes = currentNode.children;
            }

            if (currentNode !== null) {
                currentNode.links.push({
                    url: rawUrl,
                    page: pagesByUrl.value.get(rawUrl) ?? null,
                });
            }
        } catch (e) {
            console.error(e);
        }
    }

    return toTreeNodes(roots);
});
</script>

<template>
    <Head :title="`Sitemap tree: ${sitemap.domain}`" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex gap-2">
            <img :src="sitemap.faviconUrl" :alt="sitemap.domain" size="32" />
            <h1 class="text-xl font-bold">
                {{ sitemap.domain }}
            </h1>
        </div>

        <Card class="gap-0">
            <CardContent>
                <div
                    v-if="treeNodes.length === 0"
                    class="rounded-lg border border-dashed p-8 text-center"
                >
                    <h2 class="text-lg font-semibold">
                        No URL nodes available
                    </h2>
                    <p class="mt-2 text-sm text-muted-foreground">
                        This sitemap does not currently contain parsable URLs.
                    </p>
                </div>

                <ul v-else class="space-y-2">
                    <UrlTreeNode
                        v-for="node in treeNodes"
                        :key="node.id"
                        :node="node"
                    />
                </ul>
            </CardContent>
        </Card>
    </div>
</template>
