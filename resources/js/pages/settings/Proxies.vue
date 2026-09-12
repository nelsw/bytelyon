<script setup lang="ts">
import { Form, Head, router } from '@inertiajs/vue3';
import { ShieldOff } from '@lucide/vue';
import ProxyController from '@/actions/App/Http/Controllers/Settings/ProxyController';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import PasswordInput from '@/components/PasswordInput.vue';
import ProxyItem from '@/components/ProxyItem.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { destroy, edit } from '@/routes/proxies';
import type { Proxy } from '@/types/auth';

const props = defineProps<{
    proxies: Proxy[];
}>();

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Proxies settings',
                href: edit(),
            },
        ],
    },
});

const handleProxyDelete = (id: number, onError: () => void) => {
    router.delete(destroy.url(id), {
        preserveScroll: true,
        onError,
    });
};

const selectClass =
    'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex h-9 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<template>
    <Head title="Proxies settings" />

    <h1 class="sr-only">Proxies settings</h1>

    <div class="space-y-6">
        <Heading
            variant="small"
            title="Proxies"
            description="Manage the proxies your bots use to make outbound requests"
        />

        <Form
            v-bind="ProxyController.store.form()"
            :options="{ preserveScroll: true }"
            reset-on-success
            class="space-y-6"
            v-slot="{ errors, processing }"
        >
            <div class="grid gap-4">
                <div class="grid gap-2">
                    <Label for="proxy_name">Name</Label>
                    <Input
                        id="proxy_name"
                        class="mt-1 block w-full"
                        name="name"
                        required
                        placeholder="Residential US"
                    />
                    <InputError class="mt-2" :message="errors.name" />
                </div>

                <div class="grid gap-4 sm:grid-cols-12">
                    <div class="grid gap-2 sm:col-span-3">
                        <Label for="proxy_protocol">Protocol</Label>
                        <select
                            id="proxy_protocol"
                            name="protocol"
                            :class="selectClass"
                        >
                            <option value="http">http</option>
                            <option value="https">https</option>
                            <option value="socks5">socks5</option>
                        </select>
                        <InputError class="mt-2" :message="errors.protocol" />
                    </div>

                    <div class="grid gap-2 sm:col-span-6">
                        <Label for="proxy_server">Server</Label>
                        <Input
                            id="proxy_server"
                            class="mt-1 block w-full"
                            name="server"
                            required
                            placeholder="proxy.example.com"
                        />
                        <InputError class="mt-2" :message="errors.server" />
                    </div>

                    <div class="grid gap-2 sm:col-span-3">
                        <Label for="proxy_port">Port (optional)</Label>
                        <Input
                            id="proxy_port"
                            type="number"
                            min="1"
                            max="65535"
                            class="mt-1 block w-full"
                            name="port"
                            placeholder="8080"
                        />
                        <InputError class="mt-2" :message="errors.port" />
                    </div>
                </div>

                <div class="grid gap-4 sm:grid-cols-2">
                    <div class="grid gap-2">
                        <Label for="proxy_username">Username (optional)</Label>
                        <Input
                            id="proxy_username"
                            class="mt-1 block w-full"
                            name="username"
                        />
                        <InputError class="mt-2" :message="errors.username" />
                    </div>

                    <div class="grid gap-2">
                        <Label for="proxy_password">Password (optional)</Label>
                        <PasswordInput
                            id="proxy_password"
                            name="password"
                            class="mt-1 block w-full"
                        />
                        <InputError class="mt-2" :message="errors.password" />
                    </div>
                </div>

                <div class="grid gap-2">
                    <Label for="proxy_bypass">Bypass (optional)</Label>
                    <Input
                        id="proxy_bypass"
                        class="mt-1 block w-full"
                        name="bypass"
                        placeholder="localhost,127.0.0.1"
                    />
                    <InputError class="mt-2" :message="errors.bypass" />
                </div>
            </div>

            <div class="flex items-center gap-4">
                <Button :disabled="processing" data-test="create-proxy-button">
                    Add proxy
                </Button>
            </div>
        </Form>

        <div
            v-if="props.proxies.length"
            class="overflow-x-auto rounded-lg border border-border"
        >
            <table class="w-full min-w-200 text-left text-sm">
                <thead class="bg-muted/50 text-muted-foreground">
                    <tr class="border-b">
                        <th class="px-4 py-3 font-medium">Name</th>
                        <th class="px-4 py-3 font-medium">Protocol</th>
                        <th class="px-4 py-3 font-medium">Server</th>
                        <th class="px-4 py-3 font-medium">Port</th>
                        <th class="px-4 py-3 font-medium">Username</th>
                        <th class="px-4 py-3 font-medium">Bypass</th>
                        <th class="px-4 py-3 font-medium">Added</th>
                        <th class="px-4 py-3 text-right font-medium">
                            Actions
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <ProxyItem
                        v-for="proxy in props.proxies"
                        :key="proxy.id"
                        :proxy="proxy"
                        @delete="handleProxyDelete"
                    />
                </tbody>
            </table>
        </div>

        <div v-else class="rounded-lg border border-border p-8 text-center">
            <div
                class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-muted"
            >
                <ShieldOff class="h-7 w-7 text-muted-foreground" />
            </div>
            <p class="font-medium">No proxies yet</p>
            <p class="mt-1 text-sm text-muted-foreground">
                Add a proxy so your bots can make outbound requests through it
            </p>
        </div>
    </div>
</template>
