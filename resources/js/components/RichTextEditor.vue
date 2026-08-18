<script setup lang="ts">
import {
    Bold,
    Heading2,
    Heading3,
    Italic,
    Link as LinkIcon,
    List,
    ListOrdered,
    Quote,
    Redo,
    Sparkles,
    Strikethrough,
    Underline as UnderlineIcon,
    Undo,
    Unlink,
} from '@lucide/vue';
import Link from '@tiptap/extension-link';
import Placeholder from '@tiptap/extension-placeholder';
import Underline from '@tiptap/extension-underline';
import StarterKit from '@tiptap/starter-kit';
import { EditorContent, useEditor } from '@tiptap/vue-3';
import { computed, ref, watch } from 'vue';
import AlertError from '@/components/AlertError.vue';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Spinner } from '@/components/ui/spinner';
import { getXsrfToken } from '@/lib/csrf';
import { normalizeRichText } from '@/lib/richText';
import { cn } from '@/lib/utils';

defineOptions({
    inheritAttrs: false,
});

const props = withDefaults(
    defineProps<{
        id?: string;
        modelValue: string;
        placeholder?: string;
        class?: string;
        invalid?: boolean;
        assistAction?: string;
    }>(),
    {
        id: undefined,
        placeholder: 'Write something…',
        class: undefined,
        invalid: false,
        assistAction: undefined,
    },
);

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
}>();

const editor = useEditor({
    content: normalizeRichText(props.modelValue),
    extensions: [
        StarterKit,
        Underline,
        Link.configure({
            openOnClick: false,
            autolink: true,
        }),
        Placeholder.configure({
            placeholder: props.placeholder,
        }),
    ],
    editorProps: {
        attributes: {
            ...(props.id ? { id: props.id } : {}),
            class: 'prose prose-sm dark:prose-invert max-w-none min-h-56 max-h-[28rem] overflow-y-auto px-3 py-2 focus:outline-none',
        },
    },
    onUpdate: ({ editor: currentEditor }) => {
        emit('update:modelValue', currentEditor.getHTML());
    },
});

watch(
    () => props.modelValue,
    (value) => {
        if (!editor.value) {
            return;
        }

        if (value !== editor.value.getHTML()) {
            editor.value.commands.setContent(normalizeRichText(value), {
                emitUpdate: false,
            });
        }
    },
);

function setLink() {
    if (!editor.value) {
        return;
    }

    const previousUrl = editor.value.getAttributes('link').href as
        string | undefined;
    const url = window.prompt('URL', previousUrl ?? '');

    if (url === null) {
        return;
    }

    if (url === '') {
        editor.value.chain().focus().extendMarkRange('link').unsetLink().run();

        return;
    }

    editor.value
        .chain()
        .focus()
        .extendMarkRange('link')
        .setLink({ href: url })
        .run();
}

const assistDialogOpen = ref(false);
const reviewDialogOpen = ref(false);
const assistPrompt = ref('');
const assistSystem = ref('');
const assistBody = ref('');
const assistError = ref<string | null>(null);
const assistSubmitting = ref(false);
const reviewContent = ref('');

function openAssistDialog() {
    assistPrompt.value = '';
    assistSystem.value = '';
    assistBody.value = editor.value?.getHTML() ?? props.modelValue;
    assistError.value = null;
    assistDialogOpen.value = true;
}

async function submitAssist() {
    if (!props.assistAction || assistPrompt.value.trim() === '') {
        return;
    }

    assistSubmitting.value = true;
    assistError.value = null;

    try {
        const response = await fetch(props.assistAction, {
            method: 'POST',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                Accept: 'application/json',
                'X-XSRF-TOKEN': getXsrfToken() ?? '',
            },
            body: JSON.stringify({
                prompt: assistPrompt.value,
                system: assistSystem.value,
                body: assistBody.value,
            }),
        });

        const data = (await response.json()) as {
            html?: string;
            message?: string;
        };

        if (!response.ok) {
            assistError.value = data.message ?? 'The AI assist request failed.';

            return;
        }

        reviewContent.value = data.html ?? '';
        assistDialogOpen.value = false;
        reviewDialogOpen.value = true;
    } catch {
        assistError.value = 'The AI assist request failed.';
    } finally {
        assistSubmitting.value = false;
    }
}

function applyReview() {
    if (!editor.value) {
        return;
    }

    editor.value.commands.setContent(reviewContent.value);
    reviewDialogOpen.value = false;
}

const wrapperClass = computed(() =>
    cn(
        'rounded-md border border-input bg-transparent shadow-xs transition-[color,box-shadow] focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:bg-input/30 dark:aria-invalid:ring-destructive/40',
        props.invalid && 'border-destructive',
        props.class,
    ),
);

const toolbarButtonClass = (active: boolean) =>
    cn(
        'inline-flex size-8 items-center justify-center rounded-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50',
        active && 'bg-accent text-accent-foreground',
    );

const dialogTextareaClass =
    'dark:bg-input/30 border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 flex min-h-32 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';
</script>

<template>
    <div :class="wrapperClass" :aria-invalid="props.invalid">
        <div
            v-if="editor"
            class="flex flex-wrap items-center gap-1 border-b border-input px-2 py-1.5"
        >
            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('bold'))"
                title="Bold"
                @click="editor.chain().focus().toggleBold().run()"
            >
                <Bold class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('italic'))"
                title="Italic"
                @click="editor.chain().focus().toggleItalic().run()"
            >
                <Italic class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('underline'))"
                title="Underline"
                @click="editor.chain().focus().toggleUnderline().run()"
            >
                <UnderlineIcon class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('strike'))"
                title="Strikethrough"
                @click="editor.chain().focus().toggleStrike().run()"
            >
                <Strikethrough class="size-4" />
            </button>

            <span class="mx-1 h-5 w-px bg-border" />

            <button
                type="button"
                :class="
                    toolbarButtonClass(editor.isActive('heading', { level: 2 }))
                "
                title="Heading 2"
                @click="
                    editor.chain().focus().toggleHeading({ level: 2 }).run()
                "
            >
                <Heading2 class="size-4" />
            </button>
            <button
                type="button"
                :class="
                    toolbarButtonClass(editor.isActive('heading', { level: 3 }))
                "
                title="Heading 3"
                @click="
                    editor.chain().focus().toggleHeading({ level: 3 }).run()
                "
            >
                <Heading3 class="size-4" />
            </button>

            <span class="mx-1 h-5 w-px bg-border" />

            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('bulletList'))"
                title="Bullet list"
                @click="editor.chain().focus().toggleBulletList().run()"
            >
                <List class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('orderedList'))"
                title="Numbered list"
                @click="editor.chain().focus().toggleOrderedList().run()"
            >
                <ListOrdered class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('blockquote'))"
                title="Quote"
                @click="editor.chain().focus().toggleBlockquote().run()"
            >
                <Quote class="size-4" />
            </button>

            <span class="mx-1 h-5 w-px bg-border" />

            <button
                type="button"
                :class="toolbarButtonClass(editor.isActive('link'))"
                title="Add link"
                @click="setLink"
            >
                <LinkIcon class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(false)"
                title="Remove link"
                :disabled="!editor.isActive('link')"
                @click="editor.chain().focus().unsetLink().run()"
            >
                <Unlink class="size-4" />
            </button>

            <span class="mx-1 h-5 w-px bg-border" />

            <button
                type="button"
                :class="toolbarButtonClass(false)"
                title="Undo"
                :disabled="!editor.can().undo()"
                @click="editor.chain().focus().undo().run()"
            >
                <Undo class="size-4" />
            </button>
            <button
                type="button"
                :class="toolbarButtonClass(false)"
                title="Redo"
                :disabled="!editor.can().redo()"
                @click="editor.chain().focus().redo().run()"
            >
                <Redo class="size-4" />
            </button>

            <template v-if="props.assistAction">
                <span class="mx-1 h-5 w-px bg-border" />

                <button
                    type="button"
                    class="inline-flex h-8 items-center justify-center gap-1.5 rounded-full border border-green-500 bg-transparent px-3 text-sm font-medium text-green-500 shadow-[0_0_10px_rgba(34,197,94,0.7)] transition-shadow hover:shadow-[0_0_18px_rgba(34,197,94,0.9)]"
                    title="AI assist"
                    @click="openAssistDialog"
                >
                    <Sparkles class="size-4" />
                    AI Assist
                </button>
            </template>
        </div>

        <EditorContent :editor="editor" />
    </div>

    <Dialog v-model:open="assistDialogOpen">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Ask AI</DialogTitle>
                <DialogDescription>
                    Describe how you'd like the article body updated. The
                    current body is sent along with your instruction.
                </DialogDescription>
            </DialogHeader>

            <input type="hidden" name="body" :value="assistBody" />
            <textarea
                v-model="assistPrompt"
                name="prompt"
                :class="dialogTextareaClass"
                placeholder="e.g. Make the tone more conversational and add a closing summary."
                :disabled="assistSubmitting"
                autofocus
            />
            <textarea
                v-model="assistSystem"
                name="system"
                :class="dialogTextareaClass"
                placeholder="Optional: custom system instructions (e.g. tone, style, constraints) to guide the AI."
                :disabled="assistSubmitting"
            />

            <AlertError
                v-if="assistError"
                title="AI assist failed"
                :errors="[assistError]"
            />

            <DialogFooter class="gap-2">
                <Button
                    type="button"
                    variant="secondary"
                    :disabled="assistSubmitting"
                    @click="assistDialogOpen = false"
                >
                    Cancel
                </Button>
                <Button
                    type="button"
                    :disabled="assistSubmitting || assistPrompt.trim() === ''"
                    @click="submitAssist"
                >
                    <Spinner v-if="assistSubmitting" />
                    Submit
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>

    <Dialog v-model:open="reviewDialogOpen">
        <DialogContent
            class="flex h-[calc(100vh-2rem)] max-h-[calc(100vh-2rem)] w-[calc(100vw-2rem)] max-w-[calc(100vw-2rem)] flex-col"
        >
            <DialogHeader>
                <DialogTitle>Review AI suggestion</DialogTitle>
                <DialogDescription>
                    Edit the content below if needed, then apply it to replace
                    the article body, or cancel to leave the body unchanged.
                </DialogDescription>
            </DialogHeader>

            <textarea
                v-model="reviewContent"
                :class="cn(dialogTextareaClass, 'min-h-0 flex-1 resize-none')"
            />

            <DialogFooter class="gap-2">
                <Button
                    type="button"
                    variant="secondary"
                    @click="reviewDialogOpen = false"
                >
                    Cancel
                </Button>
                <Button type="button" @click="applyReview">Apply</Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
