<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import { computed } from 'vue';
import BotsTable from '@/components/bots/BotsTable.vue';
import { Card, CardContent } from '@/components/ui/card';
import { dashboard } from '@/routes';
import type { BotFilters, PaginatedBots } from '@/types/bots';

type Metrics = {
    bots: number;
    articles: number;
    searches: number;
    sitemaps: number;
};

type RecentBot = {
    id: number;
    query: string;
    enabled: boolean;
    articles: number;
    lastRunAt: string | null;
};

const props = defineProps<{
    metrics: Metrics;
    recentBots: RecentBot[];
    bots: PaginatedBots;
    filters: BotFilters;
}>();

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: dashboard(),
            },
        ],
    },
});

const numberFormat = new Intl.NumberFormat();

const tiles = computed(() => [
    { label: 'Bots', value: props.metrics.bots },
    { label: 'Articles', value: props.metrics.articles },
    { label: 'Searches', value: props.metrics.searches },
    { label: 'Sitemaps', value: props.metrics.sitemaps },
]);

// The spectrum lifted from the lion mark is reserved for data visualisation,
// so the bars cycle through it rather than through any UI colour.
// const spectrum = [
//     'var(--chart-1)',
//     'var(--chart-2)',
//     'var(--chart-3)',
//     'var(--chart-4)',
//     'var(--chart-5)',
//     'var(--chart-6)',
//     'var(--chart-7)',
// ];
//
// const peakArticles = computed(() =>
//     Math.max(1, ...props.recentBots.map((bot) => bot.articles)),
// );
</script>

<template>
    <Head title="Dashboard" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="grid grid-cols-4 gap-4">
            <Card v-for="tile in tiles" :key="tile.label">
                <CardContent>
                    <div class="text-sm text-muted-foreground">
                        {{ tile.label }}
                    </div>
                    <div
                        class="mt-1.5 font-mono text-2xl font-semibold tabular-nums"
                    >
                        {{ numberFormat.format(tile.value) }}
                    </div>
                </CardContent>
            </Card>
        </div>

        <!--        <Card>-->
        <!--            <CardContent class="space-y-4">-->
        <!--                <div class="flex items-baseline justify-between gap-4">-->
        <!--                    <h2 class="text-base font-semibold">Articles by bot</h2>-->
        <!--                    <span class="text-xs text-muted-foreground">-->
        <!--                        Five most recently run-->
        <!--                    </span>-->
        <!--                </div>-->

        <!--                <div-->
        <!--                    v-if="recentBots.length"-->
        <!--                    class="flex h-32 items-end gap-2.5"-->
        <!--                >-->
        <!--                    <div-->
        <!--                        v-for="(bot, index) in recentBots"-->
        <!--                        :key="bot.id"-->
        <!--                        class="flex flex-1 flex-col items-center gap-2"-->
        <!--                        :title="`${bot.query} — ${bot.articles} articles`"-->
        <!--                    >-->
        <!--                        <span-->
        <!--                            class="font-mono text-xs text-muted-foreground tabular-nums"-->
        <!--                        >-->
        <!--                            {{ bot.articles }}-->
        <!--                        </span>-->
        <!--                        <div-->
        <!--                            class="w-full rounded-t-sm"-->
        <!--                            :style="{-->
        <!--                                height: `${Math.max(4, (bot.articles / peakArticles) * 100)}px`,-->
        <!--                                background: spectrum[index % spectrum.length],-->
        <!--                            }"-->
        <!--                        />-->
        <!--                    </div>-->
        <!--                </div>-->

        <!--                <p-->
        <!--                    v-else-->
        <!--                    class="py-8 text-center text-sm text-muted-foreground"-->
        <!--                >-->
        <!--                    No bots yet. Create one to start collecting articles.-->
        <!--                </p>-->
        <!--            </CardContent>-->
        <!--        </Card>-->

        <Card class="gap-0 py-0">
            <div class="border-b px-5 py-4">
                <h2 class="text-base font-semibold">All bots</h2>
            </div>

            <div class="p-5">
                <BotsTable
                    :bots="bots"
                    :filters="filters"
                    base-path="/dashboard"
                />
            </div>
        </Card>
    </div>
</template>
