<script setup lang="ts">
import { Form, Link } from '@inertiajs/vue3';
import { computed } from 'vue';
import InputError from '@/components/InputError.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';

type Option = {
    value: string;
    label: string;
};

type BotFormData = {
    id?: number;
    query: string;
    type: string;
    frequency: string;
    blacklist: string;
    enabled: boolean;
    headless: boolean;
};

const props = withDefaults(
    defineProps<{
        action: string;
        method: 'post' | 'put';
        submitLabel: string;
        typeOptions: Option[];
        frequencyOptions: Option[];
        bot?: BotFormData;
        cancelHref?: string;
    }>(),
    {
        bot: () => ({
            query: '',
            type: '',
            frequency: '',
            blacklist: '',
            enabled: true,
            headless: false,
        }),
        cancelHref: undefined,
    },
);

const cancelHREF = computed(() => {
    if (props.bot === undefined) {
        return undefined;
    }

    if (props.bot.type === 'news') {
        return '/news';
    }

    if (props.bot.type === 'search') {
        return `/serps`;
    }

    return `/sitemaps`;
});

const textareaClass =
    'dark:bg-input/30 border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex min-h-24 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';

const selectClass =
    'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex h-9 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<template>
    <Form
        :action="props.action"
        :method="props.method"
        class="space-y-6"
        v-slot="{ errors, processing }"
    >
        <div class="grid gap-6 md:grid-cols-2">
            <div class="grid gap-2 md:col-span-2">
                <Label for="query">Query</Label>
                <Input
                    id="query"
                    name="query"
                    type="text"
                    placeholder="e.g. btc forecast"
                    required
                    disabled
                    :default-value="props.bot.query"
                />
                <p class="text-sm text-muted-foreground">
                    The query can't be changed after a bot is created.
                </p>
                <InputError :message="errors.query" />
            </div>

            <div class="grid gap-2">
                <Label for="type">Bot type</Label>
                <select
                    id="type"
                    name="type"
                    :class="selectClass"
                    required
                    disabled
                >
                    <option value="">Select a type</option>
                    <option
                        v-for="type in props.typeOptions"
                        :key="type.value"
                        :value="type.value"
                        :selected="type.value === props.bot.type"
                    >
                        {{ type.label }}
                    </option>
                </select>
                <p class="text-sm text-muted-foreground">
                    The bot type can't be changed after a bot is created.
                </p>
                <InputError :message="errors.type" />
            </div>

            <div class="grid gap-2">
                <Label for="frequency">Frequency</Label>
                <select
                    id="frequency"
                    name="frequency"
                    :class="selectClass"
                    required
                >
                    <option value="">Select a frequency</option>
                    <option
                        v-for="frequency in props.frequencyOptions"
                        :key="frequency.value"
                        :value="frequency.value"
                        :selected="frequency.value === props.bot.frequency"
                    >
                        {{ frequency.label }}
                    </option>
                </select>
                <InputError :message="errors.frequency" />
            </div>

            <div class="grid gap-2 md:col-span-2">
                <Label for="blacklist">Blacklist</Label>
                <textarea
                    id="blacklist"
                    name="blacklist"
                    :class="textareaClass"
                    placeholder="Enter one blocked term per line"
                    v-text="props.bot.blacklist"
                />
                <p class="text-sm text-muted-foreground">
                    Optional. Each line will be stored as a blocked term for
                    this bot.
                </p>
                <InputError :message="errors.blacklist" />
            </div>
        </div>

        <div class="grid gap-4 rounded-lg border p-4">
            <div class="flex items-start gap-3">
                <input type="hidden" name="enabled" value="0" />
                <input
                    id="enabled"
                    name="enabled"
                    type="checkbox"
                    value="1"
                    :checked="props.bot.enabled"
                    class="mt-1 size-4 rounded border border-input text-primary shadow-xs focus-visible:ring-[3px] focus-visible:ring-ring"
                />
                <div class="space-y-1">
                    <Label for="enabled">Enabled</Label>
                    <p class="text-sm text-muted-foreground">
                        Turn the bot on immediately after saving.
                    </p>
                </div>
            </div>

            <input
                type="hidden"
                name="headless"
                :value="props.bot.headless ? '1' : '0'"
            />
        </div>

        <div>
            <InputError :message="errors.enabled" />
        </div>

        <div class="flex items-center gap-3">
            <Button :disabled="processing" data-test="submit-bot-button">
                <Spinner v-if="processing" />
                {{ props.submitLabel }}
            </Button>

            <Button v-if="cancelHREF" as-child variant="outline">
                <Link :href="cancelHREF">Cancel</Link>
            </Button>
        </div>
    </Form>
</template>
