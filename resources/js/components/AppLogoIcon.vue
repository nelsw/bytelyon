<script setup lang="ts">
import { usePage } from '@inertiajs/vue3';
import { computed } from 'vue';
import type { HTMLAttributes } from 'vue';
import AppLogoIconAlt from '@/components/AppLogoIconAlt.vue';
import AppLogoIconDefault from '@/components/AppLogoIconDefault.vue';

type Props = {
    className?: HTMLAttributes['class'];
};

defineProps<Props>();

const page = usePage();
const isProduction = computed(() => page.props.isProduction);
</script>

<template>
    <AppLogoIconDefault v-if="isProduction" v-bind="$attrs" />
    <AppLogoIconAlt v-else v-bind="$attrs" />
</template>
