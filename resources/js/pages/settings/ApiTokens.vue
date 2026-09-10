<script setup lang="ts">
import { Form, Head, router } from '@inertiajs/vue3';
import { Check, Copy, KeySquare } from '@lucide/vue';
import { ref } from 'vue';
import { toast } from 'vue-sonner';
import ApiTokenController from '@/actions/App/Http/Controllers/Settings/ApiTokenController';
import ApiTokenItem from '@/components/ApiTokenItem.vue';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { destroy, edit } from '@/routes/api-tokens';
import type { ApiToken } from '@/types/auth';

const props = defineProps<{
    tokens: ApiToken[];
    plainTextToken: string | null;
}>();

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Client API settings',
                href: edit(),
            },
        ],
    },
});

const copied = ref(false);

const copyToken = async () => {
    if (!props.plainTextToken) {
        return;
    }

    try {
        await navigator.clipboard.writeText(props.plainTextToken);
        copied.value = true;
        setTimeout(() => (copied.value = false), 2000);
    } catch {
        toast.error('Could not copy the token, select and copy it manually.');
    }
};

const handleRevoke = (id: number, onError: () => void) => {
    router.delete(destroy.url(id), {
        preserveScroll: true,
        onError,
    });
};
</script>

<template>
    <Head title="Client API settings" />

    <h1 class="sr-only">Client API settings</h1>

    <div class="space-y-6">
        <Heading
            variant="small"
            title="Client API"
            description="Issue tokens so your own clients can call the ByteLyon API"
        />

        <Alert v-if="props.plainTextToken">
            <AlertTitle>Copy your new token now</AlertTitle>
            <AlertDescription>
                <p class="text-sm">
                    This is the only time it will be shown. Send it as an
                    <code>Authorization: Bearer</code> header.
                </p>
                <div class="mt-3 flex w-full items-center gap-2">
                    <code
                        class="flex-1 overflow-x-auto rounded-md bg-muted px-3 py-2 font-mono text-xs break-all"
                        data-test="plain-text-token"
                        >{{ props.plainTextToken }}</code
                    >
                    <Button
                        variant="outline"
                        size="icon-sm"
                        type="button"
                        aria-label="Copy token"
                        @click="copyToken"
                    >
                        <Check v-if="copied" class="h-4 w-4" />
                        <Copy v-else class="h-4 w-4" />
                    </Button>
                </div>
            </AlertDescription>
        </Alert>

        <Form
            v-bind="ApiTokenController.store.form()"
            :options="{ preserveScroll: true }"
            reset-on-success
            class="space-y-6"
            v-slot="{ errors, processing }"
        >
            <div class="grid gap-2">
                <Label for="token_name">Token name</Label>
                <Input
                    id="token_name"
                    class="mt-1 block w-full"
                    name="name"
                    required
                    placeholder="My scraper"
                />
                <InputError class="mt-2" :message="errors.name" />
            </div>

            <div class="flex items-center gap-4">
                <Button :disabled="processing" data-test="create-token-button">
                    Issue token
                </Button>
            </div>
        </Form>

        <div class="overflow-hidden rounded-lg border border-border">
            <template v-if="props.tokens.length">
                <ApiTokenItem
                    v-for="token in props.tokens"
                    :key="token.id"
                    :token="token"
                    @revoke="handleRevoke"
                />
            </template>

            <div v-else class="p-8 text-center">
                <div
                    class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-muted"
                >
                    <KeySquare class="h-7 w-7 text-muted-foreground" />
                </div>
                <p class="font-medium">No tokens yet</p>
                <p class="mt-1 text-sm text-muted-foreground">
                    Issue a token to let a client authenticate against the API
                </p>
            </div>
        </div>
    </div>
</template>
