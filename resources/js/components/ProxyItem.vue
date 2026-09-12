<script setup lang="ts">
import { Globe, Trash2 } from '@lucide/vue';
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
import type { Proxy } from '@/types/auth';

const props = defineProps<{
    proxy: Proxy;
}>();

const emit = defineEmits<{
    delete: [id: number, onError: () => void];
}>();

const isDeleting = ref(false);

const handleDelete = () => {
    isDeleting.value = true;
    emit('delete', props.proxy.id, () => {
        isDeleting.value = false;
    });
};
</script>

<template>
    <tr class="border-b transition-colors last:border-b-0 hover:bg-muted/30">
        <td class="px-4 py-3 align-middle">
            <div class="flex items-center gap-3">
                <div
                    class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-muted"
                >
                    <Globe class="h-4 w-4 text-muted-foreground" />
                </div>
                <span class="font-medium tracking-tight">{{ proxy.name }}</span>
            </div>
        </td>
        <td class="px-4 py-3 align-middle text-muted-foreground uppercase">
            {{ proxy.scheme }}
        </td>
        <td class="px-4 py-3 align-middle text-muted-foreground">
            {{ proxy.host }}
        </td>
        <td class="px-4 py-3 align-middle text-muted-foreground">
            {{ proxy.port ?? '—' }}
        </td>
        <td class="px-4 py-3 align-middle text-muted-foreground">
            {{ proxy.user ?? '—' }}
        </td>
        <td class="px-4 py-3 align-middle text-muted-foreground">
            {{ proxy.bypass ?? '—' }}
        </td>
        <td class="px-4 py-3 align-middle text-muted-foreground">
            {{ proxy.created_at_diff ?? '—' }}
        </td>
        <td class="px-4 py-3 text-right align-middle">
            <Dialog>
                <DialogTrigger as-child>
                    <Button
                        variant="ghost"
                        size="sm"
                        class="text-destructive hover:bg-destructive/10 hover:text-destructive-hover"
                        :data-test="`delete-proxy-${proxy.id}`"
                    >
                        <Trash2 class="h-4 w-4" />
                        <span class="sr-only">Delete</span>
                    </Button>
                </DialogTrigger>

                <DialogContent>
                    <DialogTitle>Delete proxy</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to delete the "{{ proxy.name }}"
                        proxy? Anything using it will immediately lose access.
                    </DialogDescription>
                    <DialogFooter class="gap-2">
                        <DialogClose as-child>
                            <Button variant="secondary">Cancel</Button>
                        </DialogClose>
                        <Button
                            variant="destructive"
                            :disabled="isDeleting"
                            @click="handleDelete"
                        >
                            {{ isDeleting ? 'Deleting...' : 'Delete proxy' }}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </td>
    </tr>
</template>
