<script setup lang="ts">
import { Form, Link } from '@inertiajs/vue3';
import { computed, ref, watch } from 'vue';
import InputError from '@/components/InputError.vue';
import RichTextEditor from '@/components/RichTextEditor.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';

type Option = {
    value: string;
    label: string;
};

type ArticleFormData = {
    id?: number;
    bot_id: number;
    title: string;
    url: string;
    published_at: string;
    img_alt: string;
    img_url: string;
    source: string;
    keywords: string[];
    description: string;
    body: string;
};

const props = withDefaults(
    defineProps<{
        action: string;
        method: 'post' | 'put';
        submitLabel: string;
        botOptions: Option[];
        article?: ArticleFormData;
        cancelHref?: string;
    }>(),
    {
        article: () => ({
            bot_id: 0,
            title: '',
            url: '',
            published_at: '',
            img_alt: '',
            img_url: '',
            source: '',
            keywords: [],
            description: '',
            body: '',
        }),
        cancelHref: undefined,
    },
);

const textareaClass =
    'dark:bg-input/30 border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex min-h-24 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50';

const bodyContent = ref(props.article.body);

const imgUrl = ref(props.article.img_url);
const imgPreviewError = ref(false);

watch(imgUrl, () => {
    imgPreviewError.value = false;
});

const keywordsText = ref(props.article.keywords.join('\n'));

const keywordsList = computed(() =>
    keywordsText.value
        .split('\n')
        .map((keyword) => keyword.trim())
        .filter(Boolean),
);

function formatDateTimeLocal(value: string): string {
    if (value === '') {
        return '';
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
        return value;
    }

    const localDate = new Date(
        date.getTime() - date.getTimezoneOffset() * 60_000,
    );

    return localDate.toISOString().slice(0, 16);
}
</script>

<template>
    <Form
        :action="props.action"
        :method="props.method"
        class="space-y-6"
        v-slot="{ errors, processing }"
    >
        <div class="grid gap-6 lg:grid-cols-2 lg:items-start">
            <div class="grid gap-6 md:grid-cols-2">
                <div class="grid gap-2 md:col-span-2">
                    <Label for="title">Title</Label>
                    <Input
                        id="title"
                        name="title"
                        type="text"
                        required
                        :default-value="props.article.title"
                    />
                    <InputError :message="errors.title" />
                </div>

                <div class="grid gap-2 md:col-span-2">
                    <Label for="url">Link</Label>
                    <Input
                        id="url"
                        name="url"
                        type="url"
                        required
                        :default-value="props.article.url"
                    />
                    <InputError :message="errors.url" />
                </div>

                <div class="grid gap-2 md:col-span-2">
                    <Label for="img_url">Image URL</Label>
                    <Input
                        id="img_url"
                        name="img_url"
                        type="url"
                        required
                        v-model="imgUrl"
                    />
                    <InputError :message="errors.img_url" />
                </div>

                <div class="grid gap-2">
                    <Label for="img_alt">Image alt text</Label>
                    <Input
                        id="img_alt"
                        name="img_alt"
                        type="text"
                        required
                        :default-value="props.article.img_alt"
                    />
                    <InputError :message="errors.img_alt" />
                </div>

                <div class="grid gap-2">
                    <Label for="published_at">Published at</Label>
                    <Input
                        id="published_at"
                        name="published_at"
                        type="datetime-local"
                        required
                        :default-value="
                            formatDateTimeLocal(props.article.published_at)
                        "
                    />
                    <InputError :message="errors.published_at" />
                </div>

                <div class="grid gap-2">
                    <Label for="description">Description</Label>
                    <textarea
                        id="description"
                        name="description"
                        :class="textareaClass"
                        :rows="7"
                        required
                        v-text="props.article.description"
                    />
                    <InputError :message="errors.description" />
                </div>

                <div class="grid gap-2">
                    <Label for="keywords">Keywords</Label>
                    <textarea
                        id="keywords"
                        :class="textareaClass"
                        :rows="7"
                        placeholder="One keyword per line"
                        v-model="keywordsText"
                    />
                    <input
                        v-for="(keyword, index) in keywordsList"
                        :key="index"
                        type="hidden"
                        name="keywords[]"
                        :value="keyword"
                    />
                    <InputError :message="errors.keywords" />
                </div>
            </div>

            <div class="grid gap-2">
                <Label for="body">Body</Label>
                <input type="hidden" name="body" :value="bodyContent" />
                <RichTextEditor
                    id="body"
                    v-model="bodyContent"
                    placeholder="Write the article body…"
                    :invalid="Boolean(errors.body)"
                    :assist-action="`${props.action}/assist`"
                />
                <InputError :message="errors.body" />
            </div>
        </div>

        <div class="flex items-center gap-3">
            <Button :disabled="processing" data-test="submit-article-button">
                <Spinner v-if="processing" />
                {{ props.submitLabel }}
            </Button>

            <Button v-if="props.cancelHref" as-child variant="outline">
                <Link :href="props.cancelHref">Cancel</Link>
            </Button>
        </div>
    </Form>
</template>
