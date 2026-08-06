<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import DOMPurify from 'dompurify';
import { computed } from 'vue';
import DeleteArticleButton from '@/components/articles/DeleteArticleButton.vue';
import EditButton from '@/components/EditButton.vue';
import { Button } from '@/components/ui/button';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import { normalizeRichText } from '@/lib/richText';
import { formatDate } from '@/lib/utils';
import { dashboard } from '@/routes';

type Bot = {
    id: number;
    query: string;
};

type Article = {
    id: number;
    bot_id: number;
    title: string;
    url: string;
    published_at: string;
    created_at: string;
    updated_at: string;
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
                title: 'View article',
                href: '#',
            },
        ],
    },
});

const props = defineProps<{
    bot: Bot;
    article: Article;
}>();

const sanitizedBody = computed(() =>
    DOMPurify.sanitize(normalizeRichText(props.article.body)),
);
</script>

<template>
    <Head :title="article.title" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex items-center justify-between gap-4">
            <div class="space-y-1">
                <h1 class="text-xl font-bold">{{ article.title }}</h1>
                <p class="text-sm text-muted-foreground">
                    <a
                        :href="article.url"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-sm break-all text-primary hover:underline"
                    >
                        {{ article.url }}
                    </a>
                </p>
            </div>

            <div class="flex items-center gap-2">
                <EditButton
                    :href="`/bots/${bot.id}/articles/${article.id}/edit`"
                />
                <DeleteArticleButton
                    :bot-id="bot.id"
                    :article-id="article.id"
                    :article-title="article.title"
                />
            </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-[2fr_3fr_3fr_2fr]">
            <Card>
                <CardContent class="grid gap-6 text-center md:grid-cols-2">
                    <div class="space-y-2 md:col-span-2">
                        <TooltipProvider :delay-duration="0">
                            <Tooltip>
                                <TooltipContent side="bottom" align="center">
                                    {{ article.img_alt }}
                                </TooltipContent>
                                <TooltipTrigger>
                                    <a
                                        :href="article.img_url"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        class="inline-block"
                                    >
                                        <img
                                            :src="article.img_url"
                                            :alt="article.img_alt"
                                            class="max-h-56 w-auto rounded-lg border object-contain"
                                        />
                                    </a>
                                </TooltipTrigger>
                            </Tooltip>
                        </TooltipProvider>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Description</CardTitle>
                    <CardDescription>
                        Summary text stored for this article.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <p class="text-sm whitespace-pre-wrap">
                        {{ article.description }}
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Keywords</CardTitle>
                    <CardDescription>
                        Meta used to find and categorize this article.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div class="space-y-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Keywords
                        </p>
                        <p class="text-sm">{{ article.keywords.join(', ') }}</p>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Activity</CardTitle>
                    <CardDescription> Article timestamps. </CardDescription>
                </CardHeader>
                <CardContent class="space-y-4 text-sm">
                    <div class="space-y-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Published
                        </p>
                        <p class="text-sm">
                            {{ formatDate(article.published_at) }}
                        </p>
                    </div>
                    <div class="space-y-1">
                        <p class="font-medium text-muted-foreground">Created</p>
                        <p>
                            {{ formatDate(article.created_at) }}
                        </p>
                    </div>
                    <div class="space-y-1">
                        <p class="font-medium text-muted-foreground">Updated</p>
                        <p>
                            {{ formatDate(article.updated_at) }}
                        </p>
                    </div>
                </CardContent>
            </Card>
        </div>

        <Card>
            <CardHeader>
                <CardTitle>Body</CardTitle>
                <CardDescription>
                    Full article content stored in the database.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div
                    class="prose prose-sm max-w-none rounded-lg border bg-muted/20 p-4 dark:prose-invert"
                    v-html="sanitizedBody"
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
