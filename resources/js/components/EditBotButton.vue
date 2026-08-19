<script setup lang="ts">
import { Link } from '@inertiajs/vue3';
import { Pencil } from '@lucide/vue';
import BotDrawer from '@/components/BotDrawer.vue';
import { Button } from '@/components/ui/button';
import type { BotFormData } from '@/types/bots';

type Props = {
    bot?: BotFormData & { id: number };
    botId?: number;
};

defineProps<Props>();
</script>

<template>
    <BotDrawer v-if="bot" :bot="bot">
        <template #trigger>
            <Button variant="outline" size="icon-sm">
                <Pencil style="color: var(--warning)" />
            </Button>
        </template>
    </BotDrawer>

    <Button v-else-if="botId" as-child variant="outline" size="icon-sm">
        <Link :href="`/bots/${botId}/edit`">
            <Pencil style="color: var(--warning)" />
        </Link>
    </Button>
</template>
