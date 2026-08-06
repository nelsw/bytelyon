<script setup lang="ts">
import { Link } from '@inertiajs/vue3';
import { Pencil } from '@lucide/vue';
import { Button } from '@/components/ui/button';

type Props = {
    href: string;
};

defineProps<Props>();
</script>

<template>
    <Button as-child variant="outline" size="icon-sm">
        <Link :href="href">
            <Pencil style="color: var(--warning)" />
        </Link>
    </Button>
</template>
