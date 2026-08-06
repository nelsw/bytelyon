<script setup lang="ts">
import { Form } from '@inertiajs/vue3';
import { ref } from 'vue';
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
import { Spinner } from '@/components/ui/spinner';

withDefaults(
    defineProps<{
        botId: number;
        articleId: number;
        articleTitle: string;
        label?: string;
        variant?: ButtonVariants['variant'];
        size?: ButtonVariants['size'];
    }>(),
    {
        label: 'Publish',
        variant: 'outline',
        size: 'default',
    },
);

const open = ref(false);
</script>

<template>
    <Dialog v-model:open="open">
        <DialogTrigger as-child>
            <Button :variant="variant" :size="size">
                {{ label }}
            </Button>
        </DialogTrigger>

        <DialogContent>
            <Form
                :action="`/bots/${botId}/articles/${articleId}/publish`"
                method="post"
                class="space-y-6"
                @success="open = false"
                v-slot="{ processing }"
            >
                <DialogHeader class="space-y-3">
                    <DialogTitle>Publish article?</DialogTitle>
                    <DialogDescription>
                        This will publish
                        <span class="font-medium text-foreground">{{
                            articleTitle
                        }}</span>
                        to your connected Shopify store.
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter class="gap-2">
                    <DialogClose as-child>
                        <Button variant="secondary">Cancel</Button>
                    </DialogClose>

                    <Button type="submit" :disabled="processing">
                        <Spinner v-if="processing" />
                        Publish article
                    </Button>
                </DialogFooter>
            </Form>
        </DialogContent>
    </Dialog>
</template>
