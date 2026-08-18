<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import { Button } from '@/components/ui/button';
import { dashboard, login, register } from '@/routes';
</script>

<template>
    <Head title="Welcome" />
    <div
        class="flex min-h-screen flex-col items-center bg-background p-6 text-foreground lg:justify-center lg:p-8"
    >
        <div
            class="flex w-full items-center justify-center opacity-100 transition-opacity duration-750 lg:grow starting:opacity-0"
        >
            <main
                class="flex w-full max-w-83.75 flex-col overflow-hidden rounded-xl border border-border bg-card shadow-lg lg:max-w-4xl lg:flex-row"
            >
                <div
                    class="lion-panel relative aspect-335/364 w-full shrink-0 overflow-hidden lg:aspect-auto lg:w-99.5"
                >
                    <div class="lion-bg" />
                    <div class="lion-ring lion-ring-1" />
                    <div class="lion-ring lion-ring-2" />
                    <div class="lion-ring lion-ring-3" />
                    <div class="lion-sparkles">
                        <span class="spark spark-1" />
                        <span class="spark spark-2" />
                        <span class="spark spark-3" />
                        <span class="spark spark-4" />
                        <span class="spark spark-5" />
                        <span class="spark spark-6" />
                    </div>
                    <div class="lion-float">
                        <img src="/icon.png" alt="ByteLyon" class="lion-img" />
                        <div class="lion-shadow" />
                    </div>
                </div>

                <div
                    class="flex flex-1 flex-col justify-center p-8 lg:order-first lg:p-16"
                >
                    <h1
                        class="font-display text-3xl leading-tight font-bold tracking-tight"
                    >
                        ByteLyon
                    </h1>
                    <p class="mt-3 text-base text-muted-foreground">
                        Web
                        <span class="inline-flex items-baseline gap-1.5">
                            <span class="line-through opacity-50">Crawler</span>
                            <span class="font-semibold text-foreground"
                                >Prowler</span
                            >
                        </span>
                        &amp;
                        <span class="inline-flex items-baseline gap-1.5">
                            <span class="line-through opacity-50">Scraper</span>
                            <span class="font-semibold text-foreground"
                                >Hunter</span
                            >
                        </span>
                    </p>

                    <div class="mt-8 flex flex-wrap gap-3">
                        <Button v-if="$page.props.auth.user" as-child size="lg">
                            <Link :href="dashboard()">Go to dashboard</Link>
                        </Button>
                        <template v-else>
                            <Button as-child size="lg">
                                <Link :href="register()">Get started</Link>
                            </Button>
                            <Button as-child variant="outline" size="lg">
                                <Link :href="login()">Log in</Link>
                            </Button>
                        </template>
                    </div>
                </div>
            </main>
        </div>
        <div class="hidden h-14.5 lg:block"></div>
    </div>
</template>

<style scoped>
/* The marketing hero is the one place the lion gradient is permitted, so the
   whole panel is built from the spectrum rather than from UI colours. */
.lion-panel {
    background: linear-gradient(135deg, #12131a 0%, #2b0a27 48%, #04203a 100%);
    display: flex;
    align-items: center;
    justify-content: center;
}

/* animated gradient sweep behind the lion */
.lion-bg {
    position: absolute;
    inset: 0;
    background: radial-gradient(
        ellipse at 50% 60%,
        color-mix(in srgb, var(--spectrum-pink) 26%, transparent) 0%,
        color-mix(in srgb, var(--spectrum-blue) 20%, transparent) 45%,
        transparent 75%
    );
    animation: bgPulse 4s ease-in-out infinite;
}

@keyframes bgPulse {
    0%,
    100% {
        opacity: 0.6;
        transform: scale(1);
    }
    50% {
        opacity: 1;
        transform: scale(1.12);
    }
}

/* ── pulsing rings ── */
.lion-ring {
    position: absolute;
    border-radius: 50%;
    border: 1.5px solid rgb(255 255 255 / 0.12);
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%) scale(0.5);
    animation: ringExpand 3.6s ease-out infinite;
}
.lion-ring-1 {
    width: 280px;
    height: 280px;
    animation-delay: 0s;
}
.lion-ring-2 {
    width: 380px;
    height: 380px;
    animation-delay: 1.2s;
}
.lion-ring-3 {
    width: 480px;
    height: 480px;
    animation-delay: 2.4s;
}

@keyframes ringExpand {
    0% {
        transform: translate(-50%, -50%) scale(0.5);
        opacity: 0.7;
    }
    100% {
        transform: translate(-50%, -50%) scale(1.4);
        opacity: 0;
    }
}

/* ── sparkle particles ── */
.lion-sparkles {
    position: absolute;
    inset: 0;
    pointer-events: none;
}

.spark {
    position: absolute;
    border-radius: 50%;
    animation: sparkFloat 5s ease-in-out infinite;
}

.spark-1 {
    width: 6px;
    height: 6px;
    background: var(--spectrum-yellow);
    left: 18%;
    top: 72%;
    animation-delay: 0s;
}
.spark-2 {
    width: 4px;
    height: 4px;
    background: var(--spectrum-orchid);
    left: 35%;
    top: 80%;
    animation-delay: 0.8s;
}
.spark-3 {
    width: 5px;
    height: 5px;
    background: var(--spectrum-cyan);
    left: 55%;
    top: 75%;
    animation-delay: 1.5s;
}
.spark-4 {
    width: 3px;
    height: 3px;
    background: var(--spectrum-yellow);
    left: 70%;
    top: 68%;
    animation-delay: 2.2s;
}
.spark-5 {
    width: 5px;
    height: 5px;
    background: var(--spectrum-orange);
    left: 25%;
    top: 20%;
    animation-delay: 0.4s;
}
.spark-6 {
    width: 4px;
    height: 4px;
    background: var(--spectrum-purple);
    left: 75%;
    top: 25%;
    animation-delay: 1.1s;
}

@keyframes sparkFloat {
    0% {
        transform: translateY(0) scale(1);
        opacity: 0;
    }
    20% {
        opacity: 1;
    }
    80% {
        opacity: 0.8;
    }
    100% {
        transform: translateY(-80px) scale(0.3);
        opacity: 0;
    }
}

/* ── floating lion ── */
.lion-float {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    animation: lionFloat 3.8s ease-in-out infinite;
    z-index: 1;
}

@keyframes lionFloat {
    0%,
    100% {
        transform: translateY(0px) rotate(-1deg);
    }
    50% {
        transform: translateY(-18px) rotate(1deg);
    }
}

.lion-img {
    width: 220px;
    max-width: 70%;
    filter: drop-shadow(0 0 18px rgb(228 5 111 / 0.7))
        drop-shadow(0 0 40px rgb(0 91 169 / 0.5));
    animation: glowPulse 3.8s ease-in-out infinite;
}

@keyframes glowPulse {
    0%,
    100% {
        filter: drop-shadow(0 0 14px rgb(228 5 111 / 0.6))
            drop-shadow(0 0 35px rgb(0 91 169 / 0.4));
    }
    50% {
        filter: drop-shadow(0 0 28px rgb(251 170 38 / 0.8))
            drop-shadow(0 0 55px rgb(228 5 111 / 0.6))
            drop-shadow(0 0 80px rgb(89 201 230 / 0.4));
    }
}

/* elliptical shadow on the floor */
.lion-shadow {
    width: 140px;
    height: 16px;
    border-radius: 50%;
    background: rgb(0 0 0 / 0.45);
    filter: blur(8px);
    margin-top: 8px;
    animation: shadowPulse 3.8s ease-in-out infinite;
}

@keyframes shadowPulse {
    0%,
    100% {
        transform: scaleX(1);
        opacity: 0.5;
    }
    50% {
        transform: scaleX(0.7);
        opacity: 0.25;
    }
}

/* The design system asks for no ambient motion; honour the OS setting rather
   than removing the hero animation outright. */
@media (prefers-reduced-motion: reduce) {
    .lion-bg,
    .lion-ring,
    .spark,
    .lion-float,
    .lion-img,
    .lion-shadow {
        animation: none;
    }

    .lion-ring {
        opacity: 0;
    }
}
</style>
