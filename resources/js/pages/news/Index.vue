<script setup lang="ts">
import { Head, router } from '@inertiajs/vue3';
import BotDrawer from '@/components/BotDrawer.vue';
import DeleteBotButton from '@/components/bots/DeleteBotButton.vue';
import EditBotButton from '@/components/EditBotButton.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { formatDate, formatFromNow } from '@/lib/utils';
import { dashboard } from '@/routes';

type NewsBot = {
    id: number;
    query: string;
    type: string;
    enabled: boolean;
    frequency: string;
    blacklist: string;
    headless: boolean;
    processedAt: string | null;
    createdAt: string;
    updatedAt: string;
    pageCount: number;
};

defineProps<{
    bots: NewsBot[];
}>();

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
        ],
    },
});
function openBotArticles(botId: number): void {
    router.visit(`/bots/${botId}/articles`);
}
</script>

<template>
    <Head title="News" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="space-y-1">
            <h1 class="text-xl font-bold">News</h1>
            <p class="text-sm text-muted-foreground">
                Browse your news bots and the articles each one has collected.
            </p>
        </div>

        <Card class="gap-0 py-0">
            <div
                v-if="bots.length === 0"
                class="rounded-lg border border-dashed p-8 text-center"
            >
                <h2 class="text-lg font-semibold">No news bots yet</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    Create a bot with the "News" type to start collecting
                    articles.
                </p>
                <BotDrawer>
                    <template #trigger>
                        <Button class="mt-4">Create bot</Button>
                    </template>
                </BotDrawer>
            </div>
            <div v-else class="overflow-x-auto rounded-lg">
                <table class="w-full min-w-75 text-left text-sm">
                    <thead class="bg-muted/50 text-muted-foreground">
                        <tr class="border-b">
                            <th class="px-4 py-3 font-medium">Query</th>
                            <th class="px-4 py-3 font-medium">Articles</th>
                            <th class="px-4 py-3 font-medium">Frequency</th>
                            <th class="px-4 py-3 font-medium">Status</th>
                            <th class="px-4 py-3 font-medium">Processed</th>
                            <th class="px-4 py-3 font-medium">Updated</th>
                            <th class="px-4 py-3 font-medium">Created</th>
                            <th class="px-4 py-3 text-right font-medium">
                                <span class="m-3">Actions</span>
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="bot in bots"
                            :key="bot.id"
                            class="cursor-pointer border-b transition-colors last:border-b-0 hover:bg-muted/30"
                            @click="openBotArticles(bot.id)"
                        >
                            <td class="px-4 py-3 align-middle font-medium">
                                {{ bot.query }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ bot.pageCount }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                <span class="capitalize">{{
                                    bot.frequency
                                }}</span>
                            </td>
                            <td class="px-4 py-3 align-middle">
                                <Badge
                                    :variant="
                                        bot.enabled ? 'success' : 'secondary'
                                    "
                                >
                                    {{ bot.enabled ? 'Enabled' : 'Disabled' }}
                                </Badge>
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
                            <td class="px-4 py-3" @click.stop>
                                <div
                                    class="flex flex-wrap items-center justify-end gap-2"
                                >
                                    <EditBotButton :bot="bot" />
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
        </Card>
    </div>
</template>
