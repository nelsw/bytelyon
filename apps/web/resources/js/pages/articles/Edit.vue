<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import ArticleForm from '@/components/articles/ArticleForm.vue';
import DeleteArticleButton from '@/components/articles/DeleteArticleButton.vue';
import PublishArticleButton from '@/components/articles/PublishArticleButton.vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { dashboard } from '@/routes';

type Bot = {
    id: number;
    query: string;
};

type Option = {
    value: string;
    label: string;
};

type Article = {
    id: number;
    bot_id: number;
    title: string;
    url: string;
    published_at: string;
    img_alt: string;
    img_url: string;
    source: string;
    keywords: string[];
    description: string;
    body: string;
};

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: dashboard(),
            },
            {
                title: 'Bots',
                href: '/bots',
            },
            {
                title: 'Articles',
                href: '#',
            },
            {
                title: 'Edit article',
                href: '#',
            },
        ],
    },
});

defineProps<{
    bot: Bot;
    article: Article;
    botOptions: Option[];
}>();
</script>

<template>
    <Head title="Edit article" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex items-center justify-between gap-4">
            <div class="space-y-1">
                <h1 class="text-xl font-bold">Edit article</h1>
                <p class="text-sm text-muted-foreground">
                    Update the article saved under bot: {{ bot.query }}
                </p>
            </div>

            <div class="flex items-center gap-2">
                <PublishArticleButton
                    :bot-id="bot.id"
                    :article-id="article.id"
                    :article-title="article.title"
                />
                <Button as-child variant="outline">
                    <Link :href="`/bots/${bot.id}/articles/${article.id}`">
                        View
                    </Link>
                </Button>
                <DeleteArticleButton
                    :bot-id="bot.id"
                    :article-id="article.id"
                    :article-title="article.title"
                />
            </div>
        </div>

        <Card class="max-w-8xl">
            <CardContent>
                <ArticleForm
                    :action="`/bots/${bot.id}/articles/${article.id}`"
                    method="put"
                    submit-label="Save"
                    :article="article"
                    :bot-options="botOptions"
                    :cancel-href="`/bots/${bot.id}/articles`"
                />
            </CardContent>
        </Card>
        <div>
            <Button as-child variant="outline">
                <Link :href="`/bots/${bot.id}/articles`">Back to articles</Link>
            </Button>
        </div>
    </div>
</template>
