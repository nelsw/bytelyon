<script setup lang="ts">
import { Form, Head } from '@inertiajs/vue3';
import IntegrationsController from '@/actions/App/Http/Controllers/Settings/IntegrationsController';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import PasswordInput from '@/components/PasswordInput.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { edit } from '@/routes/integrations';

type Anthropic = {
    api_key: string;
    default_model: string | null;
};

type Shopify = {
    store: string;
    client_id: string;
    client_secret: string;
    default_author_name: string | null;
    default_blog_id: string | null;
};

const props = defineProps<{
    anthropic: Anthropic | null;
    anthropicModels: string[];
    shopify: Shopify | null;
}>();

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Integrations settings',
                href: edit(),
            },
        ],
    },
});

const selectClass =
    'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex h-9 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<template>
    <Head title="Integrations settings" />

    <h1 class="sr-only">Integrations settings</h1>

    <div class="flex flex-col space-y-12">
        <div class="space-y-6">
            <Heading
                variant="small"
                title="Anthropic"
                description="Connect your Anthropic account to enable AI-powered features"
            />

            <Form
                v-bind="IntegrationsController.updateAnthropic.form()"
                :options="{ preserveScroll: true }"
                class="space-y-6"
                v-slot="{ errors, processing }"
            >
                <div class="grid gap-2">
                    <Label for="anthropic_api_key">API key</Label>
                    <PasswordInput
                        id="anthropic_api_key"
                        name="api_key"
                        class="mt-1 block w-full"
                        :default-value="props.anthropic?.api_key"
                        required
                        placeholder="sk-ant-..."
                    />
                    <InputError class="mt-2" :message="errors.api_key" />
                </div>

                <div class="grid gap-2">
                    <Label for="anthropic_default_model">Default model</Label>
                    <select
                        id="anthropic_default_model"
                        name="default_model"
                        :class="selectClass"
                    >
                        <option value="">Select a model</option>
                        <option
                            v-for="model in props.anthropicModels"
                            :key="model"
                            :value="model"
                            :selected="model === props.anthropic?.default_model"
                        >
                            {{ model }}
                        </option>
                        <option
                            v-if="
                                props.anthropic?.default_model &&
                                !props.anthropicModels.includes(
                                    props.anthropic.default_model,
                                )
                            "
                            :value="props.anthropic.default_model"
                            selected
                        >
                            {{ props.anthropic.default_model }}
                        </option>
                    </select>
                    <p
                        v-if="props.anthropicModels.length === 0"
                        class="text-sm text-muted-foreground"
                    >
                        Save a valid API key to load the list of available
                        models.
                    </p>
                    <InputError class="mt-2" :message="errors.default_model" />
                </div>

                <div class="flex items-center gap-4">
                    <Button
                        :disabled="processing"
                        data-test="update-anthropic-button"
                        >Save</Button
                    >
                </div>
            </Form>
        </div>

        <div class="space-y-6">
            <Heading
                variant="small"
                title="Shopify"
                description="Connect your Shopify store to enable publishing features"
            />

            <Form
                v-bind="IntegrationsController.updateShopify.form()"
                :options="{ preserveScroll: true }"
                class="space-y-6"
                v-slot="{ errors, processing }"
            >
                <div class="grid gap-2">
                    <Label for="shopify_store">Store</Label>
                    <Input
                        id="shopify_store"
                        class="mt-1 block w-full"
                        name="store"
                        :default-value="props.shopify?.store"
                        required
                        placeholder="my-store"
                    />
                    <InputError class="mt-2" :message="errors.store" />
                </div>

                <div class="grid gap-2">
                    <Label for="shopify_client_id">Client ID</Label>
                    <Input
                        id="shopify_client_id"
                        class="mt-1 block w-full"
                        name="client_id"
                        :default-value="props.shopify?.client_id"
                        required
                    />
                    <InputError class="mt-2" :message="errors.client_id" />
                </div>

                <div class="grid gap-2">
                    <Label for="shopify_client_secret">Client secret</Label>
                    <PasswordInput
                        id="shopify_client_secret"
                        name="client_secret"
                        class="mt-1 block w-full"
                        :default-value="props.shopify?.client_secret"
                        required
                    />
                    <InputError class="mt-2" :message="errors.client_secret" />
                </div>

                <div class="grid gap-2">
                    <Label for="shopify_default_author_name"
                        >Default author name</Label
                    >
                    <Input
                        id="shopify_default_author_name"
                        class="mt-1 block w-full"
                        name="default_author_name"
                        :default-value="
                            props.shopify?.default_author_name ?? undefined
                        "
                    />
                    <InputError
                        class="mt-2"
                        :message="errors.default_author_name"
                    />
                </div>

                <div class="grid gap-2">
                    <Label for="shopify_default_blog_id">Default blog ID</Label>
                    <Input
                        id="shopify_default_blog_id"
                        class="mt-1 block w-full"
                        name="default_blog_id"
                        :default-value="
                            props.shopify?.default_blog_id ?? undefined
                        "
                    />
                    <InputError
                        class="mt-2"
                        :message="errors.default_blog_id"
                    />
                </div>

                <div class="flex items-center gap-4">
                    <Button
                        :disabled="processing"
                        data-test="update-shopify-button"
                        >Save</Button
                    >
                </div>
            </Form>
        </div>
    </div>
</template>
