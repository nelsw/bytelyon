<script setup lang="ts">
import { usePage } from '@inertiajs/vue3';
import { Bot } from '@lucide/vue';
import { computed, ref } from 'vue';
import BotCreateForm from '@/components/bots/BotCreateForm.vue';
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

const open = ref(false);

const page = usePage();
const typeOptions = computed(() => page.props.typeOptions ?? []);
const frequencyOptions = computed(() => page.props.frequencyOptions ?? []);
</script>

<template>
    <Sheet v-model:open="open">
        <SheetTrigger as-child>
            <Button
                data-slot="bot-drawer-trigger"
                variant="ghost"
                size="icon"
                class="h-7 w-7"
            >
                <Bot />
                <span class="sr-only">Create bot</span>
            </Button>
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
                    Create bot
                </SheetTitle>
                <SheetDescription>
                    Configure the bot type, schedule, and query that should be
                    tracked.
                </SheetDescription>
            </SidebarHeader>

            <SidebarContent class="overflow-y-auto p-4">
                <BotCreateForm
                    :type-options="typeOptions"
                    :frequency-options="frequencyOptions"
                    @success="open = false"
                />
            </SidebarContent>
        </SheetContent>
    </Sheet>
</template>
