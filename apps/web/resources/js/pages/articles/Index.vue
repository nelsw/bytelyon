<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import { computed, ref } from 'vue';
import DeleteArticleButton from '@/components/articles/DeleteArticleButton.vue';
import EditButton from '@/components/EditButton.vue';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { formatDate } from '@/lib/utils';
import { dashboard } from '@/routes';

type Bot = {
    id: number;
    query: string;
};

type Article = {
    id: number;
    title: string;
    url: string;
    source: string;
    publishedAt: string;
    createdAt: string;
    imgUrl: string | null;
    favicon: string;
};

const props = defineProps<{
    bot: Bot;
    articles: Article[];
}>();

const titleFilter = ref('');

const filteredArticles = computed(() => {
    const search = titleFilter.value.trim().toLowerCase();

    if (search === '') {
        return props.articles;
    }

    return props.articles.filter(
        (article) =>
            article.title.toLowerCase().includes(search) ||
            article.source.toLowerCase().includes(search),
    );
});

const brokenThumbnails = ref<Set<number>>(new Set());

function handleThumbnailError(articleId: number) {
    brokenThumbnails.value = new Set(brokenThumbnails.value).add(articleId);
}

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: dashboard(),
            },
            {
                title: 'News',
                href: '/news',
            },
            {
                title: 'Articles',
                href: '/bots',
            },
        ],
    },
});
const filterPlaceholder = 'Filter Articles ...';
</script>

<template>
    <Head title="Articles" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex justify-between space-y-1">
            <div>
                <h1 class="text-xl font-bold">{{ bot.query }}</h1>
            </div>
            <div>
                <Label for="title-filter" class="sr-only">
                    {{ filterPlaceholder }}
                </Label>
                <Input
                    id="title-filter"
                    v-model="titleFilter"
                    type="text"
                    :placeholder="filterPlaceholder"
                    class="w-56"
                />
            </div>
        </div>

        <Card class="gap-0 py-0">
            <div
                v-if="articles.length === 0"
                class="rounded-lg border border-dashed p-8 text-center"
            >
                <h2 class="text-lg font-semibold">No articles yet</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    This bot has not collected any articles yet.
                </p>
            </div>

            <template v-else>
                <div
                    v-if="filteredArticles.length === 0"
                    class="rounded-lg border border-dashed p-8 text-center"
                >
                    <h2 class="text-lg font-semibold">No matching articles</h2>
                    <p class="mt-2 text-sm text-muted-foreground">
                        No articles match “{{ titleFilter }}”.
                    </p>
                </div>

                <div v-else class="overflow-x-auto rounded-lg">
                    <table class="w-full min-w-200 text-left text-sm">
                        <thead class="bg-muted/50 text-muted-foreground">
                            <tr class="border-b">
                                <th class="px-4 py-3 font-medium">Title</th>
                                <th class="px-4 py-3 font-medium">Source</th>
                                <th class="px-4 py-3 font-medium">Published</th>
                                <th class="px-4 py-3 text-center font-medium">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                v-for="article in filteredArticles"
                                :key="article.id"
                                class="border-b last:border-b-0"
                            >
                                <td class="px-4 py-3 align-middle font-medium">
                                    <div
                                        class="flex justify-items-start gap-3 align-middle"
                                    >
                                        <div class="mt-1">
                                            <img
                                                :src="article.favicon"
                                                :alt="article.title"
                                                class="size-4 rounded-lg object-cover"
                                                @error="
                                                    handleThumbnailError(
                                                        article.id,
                                                    )
                                                "
                                            />
                                        </div>
                                        <div>
                                            <Link
                                                :href="`/bots/${bot.id}/articles/${article.id}`"
                                                class="line-clamp-2 hover:underline"
                                            >
                                                {{ article.title }}
                                            </Link>
                                        </div>
                                    </div>
                                </td>
                                <td class="px-4 py-3 align-middle">
                                    {{ article.source }}
                                </td>
                                <td
                                    class="px-4 py-3 align-middle text-muted-foreground"
                                >
                                    {{ formatDate(article.publishedAt) }}
                                </td>
                                <td class="px-4 py-3 text-right align-middle">
                                    <div class="flex justify-end gap-2">
                                        <EditButton
                                            :href="`/bots/${bot.id}/articles/${article.id}/edit`"
                                        />
                                        <DeleteArticleButton
                                            :bot-id="bot.id"
                                            :article-id="article.id"
                                            :article-title="article.title"
                                        />
                                    </div>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </template>
        </Card>
    </div>
</template>
