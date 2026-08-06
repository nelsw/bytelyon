<script setup lang="ts">
import { Newspaper, Search, Network } from '@lucide/vue';
defineProps<{
    botType: string;
}>()
</script>

<template>
    <Newspaper v-if="botType === 'news'"  />
    <Search v-else-if="botType === 'search'"  />
    <Network v-else />
</template>

