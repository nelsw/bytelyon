<script setup lang="ts">
import { KeySquare, Trash2 } from '@lucide/vue';
import { ref } from 'vue';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import type { ApiToken } from '@/types/auth';

const props = defineProps<{
    token: ApiToken;
}>();

const emit = defineEmits<{
    revoke: [id: number, onError: () => void];
}>();

const isRevoking = ref(false);

const handleRevoke = () => {
    isRevoking.value = true;
    emit('revoke', props.token.id, () => {
        isRevoking.value = false;
    });
};
</script>

<template>
    <div class="flex items-center justify-between border-b p-4 last:border-b-0">
        <div class="flex items-center gap-4">
            <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-muted"
            >
                <KeySquare class="h-5 w-5 text-muted-foreground" />
            </div>
            <div class="space-y-1">
                <p class="font-medium tracking-tight">{{ token.name }}</p>
                <p class="text-sm text-muted-foreground">
                    Created {{ token.created_at_diff }}
                    <template v-if="token.last_used_at_diff">
                        <span class="mx-1 text-muted-foreground/50">/</span>
                        Last used {{ token.last_used_at_diff }}
                    </template>
                    <template v-else>
                        <span class="mx-1 text-muted-foreground/50">/</span>
                        Never used
                    </template>
                </p>
            </div>
        </div>

        <Dialog>
            <DialogTrigger as-child>
                <Button
                    variant="ghost"
                    size="sm"
                    class="text-destructive hover:bg-destructive/10 hover:text-destructive-hover"
                    :data-test="`revoke-token-${token.id}`"
                >
                    <Trash2 class="h-4 w-4" />
                    <span class="sr-only">Revoke</span>
                </Button>
            </DialogTrigger>

            <DialogContent>
                <DialogTitle>Revoke token</DialogTitle>
                <DialogDescription>
                    Are you sure you want to revoke the "{{ token.name }}"
                    token? Any client using it will immediately lose access to
                    the API.
                </DialogDescription>
                <DialogFooter class="gap-2">
                    <DialogClose as-child>
                        <Button variant="secondary">Cancel</Button>
                    </DialogClose>
                    <Button
                        variant="destructive"
                        :disabled="isRevoking"
                        @click="handleRevoke"
                    >
                        {{ isRevoking ? 'Revoking...' : 'Revoke token' }}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    </div>
</template>
