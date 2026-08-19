<script setup lang="ts">
import { Form } from '@inertiajs/vue3';
import { computed, ref } from 'vue';
import InputError from '@/components/InputError.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import type { BotFormData } from '@/types/bots';

type Option = {
    value: string;
    label: string;
};

const props = withDefaults(
    defineProps<{
        typeOptions: Option[];
        frequencyOptions: Option[];
        bot?: Partial<BotFormData>;
        showCancel?: boolean;
    }>(),
    {
        bot: () => ({}),
        showCancel: false,
    },
);

const emit = defineEmits<{
    success: [];
    cancel: [];
}>();

const isEditing = computed(() => props.bot.id !== undefined);
const action = computed(() =>
    isEditing.value ? `/bots/${props.bot.id}` : '/bots',
);
const method = computed<'post' | 'put'>(() =>
    isEditing.value ? 'put' : 'post',
);
const submitLabel = computed(() => (isEditing.value ? 'Save' : 'Create bot'));

const params = new URLSearchParams(window.location.search);
const query = ref(props.bot.query ?? params.get('query') ?? '');
const type = ref(props.bot.type ?? params.get('type') ?? '');
const frequency = ref(props.bot.frequency ?? params.get('frequency') ?? '');
const blacklist = ref(props.bot.blacklist ?? '');
const enabled = ref(props.bot.enabled ?? true);
const headless = ref(props.bot.headless ? '1' : '0');

const textareaClass =
    'dark:bg-input/30 border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex min-h-24 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';

const selectClass =
    'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex h-9 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<template>
    <!--suppress HtmlUnknownTarget -->
    <Form
        :action="action"
        :method="method"
        class="space-y-6"
        v-slot="{ errors, processing }"
        @success="emit('success')"
    >
        <div class="grid gap-6 md:grid-cols-2">
            <div class="grid gap-2 md:col-span-2">
                <Label for="query">Query</Label>
                <Input
                    v-model="query"
                    id="query"
                    name="query"
                    type="text"
                    placeholder="e.g. btc forecast"
                    required
                    :disabled="isEditing"
                    autocomplete="off"
                />
                <p v-if="isEditing" class="text-sm text-muted-foreground">
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
                    :disabled="isEditing"
                    v-model="type"
                >
                    <option value="">Select a type</option>
                    <option
                        v-for="typeOption in typeOptions"
                        :key="typeOption.value"
                        :value="typeOption.value"
                    >
                        {{ typeOption.label }}
                    </option>
                </select>
                <p v-if="isEditing" class="text-sm text-muted-foreground">
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
                    v-model="frequency"
                >
                    <option value="">Select a frequency</option>
                    <option
                        v-for="frequencyOption in frequencyOptions"
                        :key="frequencyOption.value"
                        :value="frequencyOption.value"
                    >
                        {{ frequencyOption.label }}
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
                    v-model="blacklist"
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
                    v-model="enabled"
                    class="mt-1 size-4 rounded border border-input text-primary shadow-xs focus-visible:ring-[3px] focus-visible:ring-ring"
                />
                <div class="space-y-1">
                    <Label for="enabled">Enabled</Label>
                    <p class="text-sm text-muted-foreground">
                        Turn the bot on immediately after saving.
                    </p>
                </div>
            </div>

            <InputError :message="errors.enabled" />

            <div class="space-y-2">
                <Label>Mode</Label>
                <div class="flex items-center gap-6">
                    <label class="flex items-center gap-2 text-sm font-normal">
                        <input
                            type="radio"
                            name="headless"
                            value="0"
                            v-model="headless"
                            class="size-4 rounded-full border border-input text-primary shadow-xs focus-visible:ring-[3px] focus-visible:ring-ring"
                        />
                        Browser
                    </label>
                    <label class="flex items-center gap-2 text-sm font-normal">
                        <input
                            type="radio"
                            name="headless"
                            value="1"
                            v-model="headless"
                            class="size-4 rounded-full border border-input text-primary shadow-xs focus-visible:ring-[3px] focus-visible:ring-ring"
                        />
                        Headless
                    </label>
                </div>
                <p class="text-sm text-muted-foreground">
                    Headless bots run without a visible browser window.
                </p>
                <InputError :message="errors.headless" />
            </div>
        </div>

        <div class="flex items-center gap-3">
            <Button :disabled="processing" data-test="submit-bot-button">
                <Spinner v-if="processing" />
                {{ submitLabel }}
            </Button>

            <Button
                v-if="showCancel"
                type="button"
                variant="outline"
                @click="emit('cancel')"
            >
                Cancel
            </Button>
        </div>
    </Form>
</template>
