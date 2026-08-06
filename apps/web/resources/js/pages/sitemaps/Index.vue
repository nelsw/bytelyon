<script setup lang="ts">
import { Head, router } from '@inertiajs/vue3';
import DeleteBotButton from '@/components/bots/DeleteBotButton.vue';
import EditBotButton from '@/components/EditBotButton.vue';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { formatDate, formatFromNow } from '@/lib/utils';
import { dashboard } from '@/routes';

type Sitemap = {
    id: number;
    domain: string;
    urls: string[];
    bot_id: number;
    bot: {
        enabled: boolean;
        frequency: string;
        processedAt: string | null;
        createdAt: string;
        updatedAt: string;
    };
};

defineProps<{
    sitemaps: Sitemap[];
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
        ],
    },
});
function openSitemap(sitemapId: number): void {
    router.visit(`/sitemaps/${sitemapId}`);
}
</script>

<template>
    <Head title="Sitemaps" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="space-y-1">
            <h1 class="text-xl font-bold">Sitemaps</h1>
            <p class="text-sm text-muted-foreground">
                Browse every available sitemap and its stored metadata.
            </p>
        </div>

        <Card class="gap-0 py-0">
            <div
                v-if="sitemaps.length === 0"
                class="rounded-lg border border-dashed p-8 text-center"
            >
                <h2 class="text-lg font-semibold">No sitemaps yet</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    There are no sitemap records available right now.
                </p>
            </div>

            <div v-else class="overflow-x-auto rounded-lg">
                <table class="w-full min-w-3xl text-left text-sm">
                    <thead class="bg-muted/50 text-muted-foreground">
                        <tr class="border-b">
                            <th class="px-4 py-3 font-medium">Domain</th>
                            <th class="px-4 py-3 font-medium">URLs</th>
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
                            v-for="sitemap in sitemaps"
                            :key="sitemap.id"
                            class="cursor-pointer border-b transition-colors last:border-b-0 hover:bg-muted/30"
                            @click="openSitemap(sitemap.id)"
                        >
                            <td class="px-4 py-3 align-middle">
                                {{ sitemap.domain }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ sitemap.urls.length }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                <span class="capitalize">{{
                                    sitemap.bot.frequency
                                }}</span>
                            </td>
                            <td class="px-4 py-3 align-middle">
                                <Badge
                                    :variant="
                                        sitemap.bot.enabled
                                            ? 'success'
                                            : 'secondary'
                                    "
                                >
                                    {{
                                        sitemap.bot.enabled
                                            ? 'Enabled'
                                            : 'Disabled'
                                    }}
                                </Badge>
                            </td>

                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ formatFromNow(sitemap.bot.processedAt) }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ formatDate(sitemap.bot.updatedAt) }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ formatDate(sitemap.bot.createdAt) }}
                            </td>
                            <td class="px-4 py-3" @click.stop>
                                <div
                                    class="flex flex-wrap items-center justify-end gap-2"
                                >
                                    <EditBotButton :bot-id="sitemap.bot_id" />
                                    <DeleteBotButton
                                        :bot-id="sitemap.bot_id"
                                        :bot-query="sitemap.domain"
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
