<script setup lang="ts">
import { Form } from '@inertiajs/vue3';
import DeleteButton from '@/components/DeleteButton.vue';
import type { ButtonVariants } from '@/components/ui/button';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';

withDefaults(
    defineProps<{
        botId: number;
        botQuery: string;
        label?: string;
        variant?: ButtonVariants['variant'];
        size?: ButtonVariants['size'];
    }>(),
    {
        label: 'Delete',
        variant: 'destructive',
        size: 'icon-sm',
    },
);
</script>

<template>
    <Dialog>
        <DialogTrigger as-child>
            <DeleteButton />
        </DialogTrigger>

        <DialogContent>
            <Form
                :action="`/bots/${botId}`"
                method="delete"
                class="space-y-6"
                v-slot="{ processing }"
            >
                <DialogHeader class="space-y-3">
                    <DialogTitle>Delete bot?</DialogTitle>
                    <DialogDescription>
                        This will permanently delete
                        <span class="font-medium text-foreground">{{
                            botQuery
                        }}</span
                        >. This action cannot be undone.
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter class="gap-2">
                    <DialogClose as-child>
                        <Button variant="secondary">Cancel</Button>
                    </DialogClose>

                    <Button
                        type="submit"
                        variant="destructive"
                        :disabled="processing"
                    >
                        Delete bot
                    </Button>
                </DialogFooter>
            </Form>
        </DialogContent>
    </Dialog>
</template>
