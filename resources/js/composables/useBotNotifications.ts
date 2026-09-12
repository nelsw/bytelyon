import { useEcho } from '@laravel/echo-vue';
import { toast } from 'vue-sonner';
import type { BotResultsPersistedEvent } from '@/types/bots';

/**
 * Listens for `BotResultsPersisted` broadcasts on the authenticated user's
 * private channel, and surfaces them as a toast as soon as a botjob
 * persists new articles, pages, serp, or sitemap results.
 */
export function useBotNotifications(userId: number): void {
    useEcho<BotResultsPersistedEvent>(`App.Models.User.${userId}`, '.bot.results.persisted', (event) => {
        toast[event.toast.type](event.toast.message);
    });
}
