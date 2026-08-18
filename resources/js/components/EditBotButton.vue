<script setup lang="ts">
import EditButton from '@/components/EditButton.vue';

type Props = {
    botId: number;
};

defineProps<Props>();
</script>

<template>
    <EditButton :href="`/bots/${botId}/edit`" />
</template>
