<script setup lang="ts">
import { Form } from '@inertiajs/vue3';
import { ref } from 'vue';
import InputError from '@/components/InputError.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';

type Option = {
    value: string;
    label: string;
};

defineProps<{
    typeOptions: Option[];
    frequencyOptions: Option[];
}>();

const emit = defineEmits<{
    success: [];
}>();

const params = new URLSearchParams(window.location.search);
const query = ref(params.get('query') || '');
const type = ref(params.get('type') || '');
const frequency = ref(params.get('frequency') || '');

const textareaClass =
    'dark:bg-input/30 border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex min-h-24 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';

const selectClass =
    'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex h-9 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<template>
    <!--suppress HtmlUnknownTarget -->
    <Form
        action="/bots"
        method="post"
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
                    autocomplete="off"
                />
                <InputError :message="errors.query" />
            </div>

            <div class="grid gap-2">
                <Label for="type">Bot type</Label>
                <select
                    id="type"
                    name="type"
                    :class="selectClass"
                    required
                    v-model="type"
                >
                    <option value="">Select a type</option>
                    <option
                        v-for="type in typeOptions"
                        :key="type.value"
                        :value="type.value"
                    >
                        {{ type.label }}
                    </option>
                </select>
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
                        v-for="frequency in frequencyOptions"
                        :key="frequency.value"
                        :value="frequency.value"
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
                    checked
                    class="mt-1 size-4 rounded border border-input text-primary shadow-xs focus-visible:ring-[3px] focus-visible:ring-ring"
                />
                <div class="space-y-1">
                    <Label for="enabled">Enabled</Label>
                    <p class="text-sm text-muted-foreground">
                        Turn the bot on immediately after creation.
                    </p>
                </div>
            </div>

            <input type="hidden" name="headless" value="0" />
        </div>

        <div>
            <InputError :message="errors.enabled" />
        </div>

        <div class="flex items-center gap-3">
            <Button :disabled="processing" data-test="create-bot-button">
                <Spinner v-if="processing" />
                Create bot
            </Button>

            <p class="text-sm text-muted-foreground">
                You can change these settings later if bot editing is added.
            </p>
        </div>
    </Form>
</template>
