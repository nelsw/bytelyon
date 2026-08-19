<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import DeleteBotButton from '@/components/bots/DeleteBotButton.vue';
import EditBotButton from '@/components/EditBotButton.vue';
import { Badge } from '@/components/ui/badge';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import { dashboard } from '@/routes';

type Bot = {
    id: number;
    query: string;
    type: string;
    frequency: string;
    enabled: boolean;
    headless: boolean;
    lastRunAt: string | null;
    createdAt: string;
    updatedAt: string;
    blacklist: string;
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
                title: 'View bot',
                href: '#',
            },
        ],
    },
});

defineProps<{
    bot: Bot;
}>();
</script>

<template>
    <Head :title="bot.query" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex items-center justify-between gap-4">
            <div class="space-y-1">
                <h1 class="text-xl font-bold">{{ bot.query }}</h1>
                <p class="text-sm text-muted-foreground">
                    Review this bot's configuration and recent status.
                </p>
            </div>

            <div class="flex items-center gap-2">
                <EditBotButton :bot="bot" />
                <DeleteBotButton :bot-id="bot.id" :bot-query="bot.query" />
            </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-[2fr_1fr]">
            <Card>
                <CardHeader>
                    <CardTitle>Details</CardTitle>
                    <CardDescription>
                        Core information and runtime settings for this bot.
                    </CardDescription>
                </CardHeader>
                <CardContent class="grid gap-6 md:grid-cols-2">
                    <div class="space-y-2 md:col-span-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Query
                        </p>
                        <p class="text-sm">{{ bot.query }}</p>
                    </div>

                    <div class="space-y-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Type
                        </p>
                        <Badge class="capitalize" variant="outline">{{
                            bot.type
                        }}</Badge>
                    </div>

                    <div class="space-y-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Frequency
                        </p>
                        <p class="text-sm capitalize">{{ bot.frequency }}</p>
                    </div>

                    <div class="space-y-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Status
                        </p>
                        <Badge :variant="bot.enabled ? 'success' : 'secondary'">
                            {{ bot.enabled ? 'Enabled' : 'Disabled' }}
                        </Badge>
                    </div>

                    <div class="space-y-2">
                        <p class="text-sm font-medium text-muted-foreground">
                            Mode
                        </p>
                        <p class="text-sm">
                            {{ bot.headless ? 'Headless' : 'Browser' }}
                        </p>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Activity</CardTitle>
                    <CardDescription
                        >Recent timestamps for this bot.</CardDescription
                    >
                </CardHeader>
                <CardContent class="space-y-4 text-sm">
                    <div class="space-y-1">
                        <p class="font-medium text-muted-foreground">
                            Last run
                        </p>
                        <p>{{ bot.lastRunAt ?? 'Never' }}</p>
                    </div>
                    <div class="space-y-1">
                        <p class="font-medium text-muted-foreground">Created</p>
                        <p>{{ new Date(bot.createdAt).toLocaleString() }}</p>
                    </div>
                    <div class="space-y-1">
                        <p class="font-medium text-muted-foreground">Updated</p>
                        <p>{{ new Date(bot.updatedAt).toLocaleString() }}</p>
                    </div>
                </CardContent>
            </Card>
        </div>

        <Card>
            <CardHeader>
                <CardTitle>Blacklist</CardTitle>
                <CardDescription>
                    Terms that should be excluded when this bot runs.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div
                    v-if="bot.blacklist.length === 0"
                    class="text-sm text-muted-foreground"
                >
                    No blacklist terms have been configured.
                </div>
                <ul v-else class="list-disc space-y-2 pl-5 text-sm">
                    <li
                        v-for="term in bot.blacklist.split(/\r|\n|\r\n/)"
                        :key="term"
                    >
                        {{ term }}
                    </li>
                </ul>
            </CardContent>
        </Card>
    </div>
</template>
