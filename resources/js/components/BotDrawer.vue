<script setup lang="ts">
import { usePage } from '@inertiajs/vue3';
import { Bot } from '@lucide/vue';
import { computed, ref } from 'vue';
import BotForm from '@/components/bots/BotForm.vue';
import { Button } from '@/components/ui/button';
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetTitle,
    SheetTrigger,
} from '@/components/ui/sheet';
import { SidebarContent, SidebarHeader } from '@/components/ui/sidebar';
import { SIDEBAR_WIDTH_MOBILE } from '@/components/ui/sidebar/utils';
import type { BotFormData } from '@/types/bots';

const props = withDefaults(
    defineProps<{
        bot?: Partial<BotFormData>;
    }>(),
    {
        bot: () => ({}),
    },
);

const open = ref(false);

const page = usePage();
const typeOptions = computed(() => page.props.typeOptions ?? []);
const frequencyOptions = computed(() => page.props.frequencyOptions ?? []);

const isEditing = computed(() => props.bot.id !== undefined);
const title = computed(() => (isEditing.value ? 'Edit bot' : 'Create bot'));
const description = computed(() =>
    isEditing.value
        ? "Update this bot's query, schedule, and runtime options."
        : 'Configure the bot type, schedule, and query that should be tracked.',
);

defineExpose({ open });
</script>

<template>
    <Sheet v-model:open="open">
        <SheetTrigger as-child>
            <slot name="trigger">
                <Button
                    data-slot="bot-drawer-trigger"
                    variant="ghost"
                    size="icon"
                    class="h-7 w-7"
                >
                    <Bot />
                    <span class="sr-only">Create bot</span>
                </Button>
            </slot>
        </SheetTrigger>

        <SheetContent
            side="right"
            class="w-(--sidebar-width) gap-0 border-sidebar-border bg-sidebar p-0 text-sidebar-foreground"
            :style="{
                '--sidebar-width': SIDEBAR_WIDTH_MOBILE,
            }"
        >
            <SidebarHeader class="gap-1.5 border-b border-sidebar-border p-4">
                <SheetTitle class="flex items-center gap-2 text-base">
                    <Bot class="size-5" />
                    {{ title }}
                </SheetTitle>
                <SheetDescription>{{ description }}</SheetDescription>
            </SidebarHeader>

            <SidebarContent class="overflow-y-auto p-4">
                <BotForm
                    :bot="bot"
                    :type-options="typeOptions"
                    :frequency-options="frequencyOptions"
                    show-cancel
                    @success="open = false"
                    @cancel="open = false"
                />
            </SidebarContent>
        </SheetContent>
    </Sheet>
</template>
