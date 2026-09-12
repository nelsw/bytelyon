<script setup lang="ts">
import { usePage } from '@inertiajs/vue3';
import { useBotNotifications } from '@/composables/useBotNotifications';
import AppLayout from '@/layouts/app/AppSidebarLayout.vue';
import type { BreadcrumbItem } from '@/types';

const { breadcrumbs = [] } = defineProps<{
    breadcrumbs?: BreadcrumbItem[];
}>();

useBotNotifications(usePage().props.auth.user.id);
</script>

<template>
    <AppLayout :breadcrumbs="breadcrumbs">
        <slot />
    </AppLayout>
</template>
