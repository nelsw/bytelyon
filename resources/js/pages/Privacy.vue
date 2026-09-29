<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import { home } from '@/routes';

const updated = 'September 29, 2026';
</script>

<template>
    <Head title="Privacy Policy" />
    <div class="min-h-screen bg-background p-6 text-foreground lg:p-8">
        <main
            class="mx-auto w-full max-w-3xl rounded-xl border border-border bg-card p-8 shadow-lg lg:p-12"
        >
            <Link
                :href="home()"
                class="text-sm text-muted-foreground hover:text-foreground"
            >
                &larr; Back to ByteLyon
            </Link>

            <h1
                class="mt-6 font-display text-3xl leading-tight font-bold tracking-tight"
            >
                Privacy Policy
            </h1>
            <p class="mt-2 text-sm text-muted-foreground">
                Last updated {{ updated }}
            </p>

            <div class="mt-8 space-y-8 text-base leading-relaxed">
                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">Our aim</h2>
                    <p>
                        ByteLyon exists to help build a better world through
                        holistic information paradigms. We believe information
                        is most valuable when it is understood in context:
                        connected, complete, and put to work for the people it
                        serves. That belief shapes how we treat the information
                        you entrust to us.
                    </p>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">
                        We do not sell your information
                    </h2>
                    <p>
                        We do not sell, rent, or trade any of the information we
                        collect, full stop. Your data is never a product, and we
                        do not share it with advertisers or data brokers.
                    </p>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">What we collect</h2>
                    <ul class="list-disc space-y-2 pl-6">
                        <li>
                            <strong>Account information:</strong> your name,
                            email address, and the credentials you use to sign
                            in, such as hashed passwords, passkeys, and
                            two-factor settings.
                        </li>
                        <li>
                            <strong>Content you create:</strong> bots, articles,
                            sitemaps, search results, proxies, and integration
                            settings you configure in ByteLyon.
                        </li>
                        <li>
                            <strong>Technical data:</strong> basic logs such as
                            IP address, browser type, and request details. We
                            use these to keep the service secure and working.
                        </li>
                    </ul>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">How we use it</h2>
                    <p>
                        We use your information only to provide, secure, and
                        improve ByteLyon. That covers authenticating you,
                        running the bots you configure, generating and
                        publishing content at your request, and communicating
                        with you about your account.
                    </p>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">Service providers</h2>
                    <p>
                        Some features rely on trusted third parties, such as AI
                        providers for writing assistance or platforms you choose
                        to connect for publishing. We share only what each
                        feature needs to work, and only when you use it. This is
                        never a sale of your data.
                    </p>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">Your choices</h2>
                    <p>
                        You can view and update your profile at any time, and
                        you can delete your account from your settings. When you
                        delete your account, we remove your associated data
                        except where we are legally required to keep it.
                    </p>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">Security</h2>
                    <p>
                        We protect your information with industry-standard
                        safeguards, including encrypted connections and hashed
                        credentials. No system is perfectly secure, but we work
                        to keep your data safe.
                    </p>
                </section>

                <section class="space-y-3">
                    <h2 class="text-xl font-semibold">Changes and contact</h2>
                    <p>
                        If we update this policy, we will revise the date above.
                        Questions? Reach us at
                        <a
                            href="mailto:hello@bytelyon.com"
                            class="font-medium underline underline-offset-4"
                            >hello@bytelyon.com</a
                        >.
                    </p>
                </section>
            </div>
        </main>
    </div>
</template>
