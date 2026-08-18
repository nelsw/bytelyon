<script setup lang="ts">
import { Link } from '@inertiajs/vue3';
import { View } from '@lucide/vue';
import { Button } from '@/components/ui/button';

type Props = {
    href: string;
};

defineProps<Props>();
</script>

<template>
    <Button as-child variant="outline" size="icon-sm">
        <Link :href="href">
            <View style="color: var(--info-strong)" />
        </Link>
    </Button>
</template>
