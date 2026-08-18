<script setup lang="ts">
import { Head } from '@inertiajs/vue3';
import BotCreateForm from '@/components/bots/BotCreateForm.vue';
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

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Dashboard',
                href: dashboard(),
            },
            {
                title: 'Create bot',
                href: '/bots/create',
            },
        ],
    },
});

defineProps<{
    typeOptions: Option[];
    frequencyOptions: Option[];
}>();
</script>

<template>
    <Head title="Create bot" />

    <div class="flex h-full flex-1 flex-col gap-5 overflow-x-auto p-6">
        <div class="space-y-1">
            <h1 class="text-xl font-bold">Create a new bot</h1>
            <p class="text-sm text-muted-foreground">
                Configure the bot type, schedule, and query that should be
                tracked.
            </p>
        </div>

        <Card class="max-w-3xl">
            <CardHeader>
                <CardTitle>Bot details</CardTitle>
                <CardDescription>
                    Create a bot that can be scheduled and managed from your
                    account.
                </CardDescription>
            </CardHeader>

            <CardContent>
                <BotCreateForm
                    :type-options="typeOptions"
                    :frequency-options="frequencyOptions"
                />
            </CardContent>
        </Card>
    </div>
</template>
