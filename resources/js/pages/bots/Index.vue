<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import BotDrawer from '@/components/BotDrawer.vue';
import BotsTable from '@/components/bots/BotsTable.vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { dashboard } from '@/routes';
import type { BotFilters, PaginatedBots } from '@/types/bots';

type Option = {
    value: string;
    label: string;
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
        ],
    },
});

defineProps<{
    bots: PaginatedBots;
    filters: BotFilters;
    typeOptions: Option[];
}>();
</script>

<template>
    <Head title="Bots" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex items-center justify-between gap-4">
            <div class="space-y-1">
                <h1 class="text-xl font-bold">My bots</h1>
                <p class="text-sm text-muted-foreground">
                    View every bot in your account and its current
                    configuration.
                </p>
            </div>

            <BotDrawer>
                <template #trigger>
                    <Button size="sm">Create bot</Button>
                </template>
            </BotDrawer>
        </div>

        <Card>
            <CardContent class="space-y-6">
                <BotsTable :bots="bots" :filters="filters" base-path="/bots" />
            </CardContent>
        </Card>
    </div>
</template>
