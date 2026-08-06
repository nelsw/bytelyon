<script setup lang="ts">
import { Head, router } from '@inertiajs/vue3';
import DeleteBotButton from '@/components/bots/DeleteBotButton.vue';
import EditBotButton from '@/components/EditBotButton.vue';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { formatDate, formatFromNow } from '@/lib/utils';
import { dashboard } from '@/routes';

type Serp = {
    id: number;
    query: string;
    bot_id: number;
    pages_count: number;
    bot: {
        enabled: boolean;
        processedAt: string | null;
        createdAt: string;
        updatedAt: string;
        frequency: string;
    };
};

defineProps<{
    serps: Serp[];
}>();

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
        ],
    },
});
function openSearch(searchId: number): void {
    router.visit(`/serps/${searchId}`);
}
</script>

<template>
    <Head title="Searches" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="space-y-1">
            <h1 class="text-xl font-bold">Searches</h1>
            <p class="text-sm text-muted-foreground">
                Browse every captured search result and its stored screenshot.
            </p>
        </div>

        <Card class="gap-0 py-0">
            <div
                v-if="serps.length === 0"
                class="rounded-lg border border-dashed p-8 text-center"
            >
                <h2 class="text-lg font-semibold">No searches yet</h2>
                <p class="mt-2 text-sm text-muted-foreground">
                    There are no serp records available right now.
                </p>
            </div>

            <div v-else class="overflow-x-auto rounded-lg">
                <table class="w-full min-w-3xl text-left text-sm">
                    <thead class="bg-muted/50 text-muted-foreground">
                        <tr class="border-b">
                            <th class="px-4 py-3 font-medium">Domain</th>
                            <th class="px-4 py-3 font-medium">Pages</th>
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
                            v-for="serp in serps"
                            :key="serp.id"
                            class="cursor-pointer border-b transition-colors last:border-b-0 hover:bg-muted/30"
                            @click="openSearch(serp.id)"
                        >
                            <td class="px-4 py-3 align-middle">
                                {{ serp.query }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ serp.pages_count }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                <span class="capitalize">{{
                                    serp.bot.frequency
                                }}</span>
                            </td>
                            <td class="px-4 py-3 align-middle">
                                <Badge
                                    :variant="
                                        serp.bot.enabled
                                            ? 'success'
                                            : 'secondary'
                                    "
                                >
                                    {{
                                        serp.bot.enabled
                                            ? 'Enabled'
                                            : 'Disabled'
                                    }}
                                </Badge>
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ formatFromNow(serp.bot.processedAt) }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ formatDate(serp.bot.updatedAt) }}
                            </td>
                            <td
                                class="px-4 py-3 align-middle text-muted-foreground"
                            >
                                {{ formatDate(serp.bot.createdAt) }}
                            </td>

                            <td class="px-4 py-3" @click.stop>
                                <div
                                    class="flex flex-wrap items-center justify-end gap-2"
                                >
                                    <EditBotButton :bot-id="serp.bot_id" />
                                    <DeleteBotButton
                                        :bot-id="serp.bot_id"
                                        :bot-query="serp.query"
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
