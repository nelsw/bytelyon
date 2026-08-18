<script setup lang="ts">
import { Trash } from '@lucide/vue';
import { Button } from '@/components/ui/button';

type Props = {
    disabled?: boolean;
};

defineProps<Props>();
</script>

<template>
    <Button
        class="cursor-pointer"
        type="submit"
        variant="outline"
        size="icon-sm"
    >
        <Trash style="color: var(--destructive)" />
    </Button>
</template>
