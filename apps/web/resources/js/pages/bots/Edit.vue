<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import BotForm from '@/components/bots/BotForm.vue';
import DeleteBotButton from '@/components/bots/DeleteBotButton.vue';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import { dashboard } from '@/routes';

type Option = {
    value: string;
    label: string;
};

type Bot = {
    id: number;
    query: string;
    type: string;
    frequency: string;
    blacklist: string;
    enabled: boolean;
    headless: boolean;
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
                title: 'Edit bot',
                href: '#',
            },
        ],
    },
});

defineProps<{
    bot: Bot;
    typeOptions: Option[];
    frequencyOptions: Option[];
}>();
</script>

<template>
    <Head title="Edit bot" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="flex items-center justify-between gap-4">
            <div class="space-y-1">
                <h1 class="text-xl font-bold">Edit bot</h1>
                <p class="text-sm text-muted-foreground">
                    Update this bot's query, schedule, and runtime options.
                </p>
            </div>

            <div class="flex items-center gap-2">
                <DeleteBotButton :bot-id="bot.id" :bot-query="bot.query" />
            </div>
        </div>

        <Card class="max-w-3xl">
            <CardHeader>
                <CardTitle>Bot settings</CardTitle>
                <CardDescription>
                    Save your changes to update how this bot runs.
                </CardDescription>
            </CardHeader>

            <CardContent>
                <BotForm
                    :action="`/bots/${bot.id}`"
                    method="put"
                    submit-label="Save"
                    :bot="bot"
                    :type-options="typeOptions"
                    :frequency-options="frequencyOptions"
                    :cancel-href="`/bots/${bot.id}`"
                />
            </CardContent>
        </Card>
    </div>
</template>
