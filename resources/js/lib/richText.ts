const HTML_TAG_PATTERN = /<[a-z][\s\S]*>/i;

export function isHtmlContent(value: string): boolean {
    return HTML_TAG_PATTERN.test(value);
}

function escapeHtml(value: string): string {
    return value
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
}

function plainTextToHtml(value: string): string {
    if (value.trim() === '') {
        return '';
    }

    return value
        .split(/\n{2,}/)
        .map(
            (paragraph) =>
                `<p>${escapeHtml(paragraph).replace(/\n/g, '<br>')}</p>`,
        )
        .join('');
}

/**
 * Legacy article bodies were stored as plain text with real newlines. HTML
 * collapses those when rendered or loaded into the editor, so anything
 * without HTML tags gets its line breaks turned into paragraphs/<br> first.
 */
export function normalizeRichText(value: string): string {
    return isHtmlContent(value) ? value : plainTextToHtml(value);
}
