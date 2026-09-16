<script setup lang="ts">
import { Link, usePage } from '@inertiajs/vue3';
import {
    Sunrise,
    Network,
    LayoutDashboard,
    Newspaper,
    Search,
    Telescope,
} from '@lucide/vue';
import AppLogo from '@/components/AppLogo.vue';
import NavFooter from '@/components/NavFooter.vue';
import NavMain from '@/components/NavMain.vue';
import NavUser from '@/components/NavUser.vue';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    useSidebar,
} from '@/components/ui/sidebar';
import { dashboard } from '@/routes';
import type { NavItem } from '@/types';

const { isMobile, setOpenMobile } = useSidebar();

function handleNavigate() {
    if (isMobile.value) {
        setOpenMobile(false);
    }
}

const mainNavItems: NavItem[] = [
    {
        title: 'Dashboard',
        href: dashboard(),
        icon: LayoutDashboard,
    },
    {
        title: 'News',
        href: '/news',
        icon: Newspaper,
    },
    {
        title: 'Sitemaps',
        href: '/sitemaps',
        icon: Network,
    },
    {
        title: 'Searches',
        href: '/serps',
        icon: Search,
    },
];

const footerNavItems: NavItem[] = [
    {
        title: 'Telescope',
        href: '/telescope',
        icon: Telescope,
        show: usePage().props.auth.canViewTelescope,
    },
].filter((item) => item.show);
</script>

<template>
    <Sidebar collapsible="icon" variant="inset">
        <SidebarHeader>
            <SidebarMenu>
                <SidebarMenuItem>
                    <SidebarMenuButton size="lg" as-child>
                        <Link :href="dashboard()" @click="handleNavigate">
                            <AppLogo />
                        </Link>
                    </SidebarMenuButton>
                </SidebarMenuItem>
            </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
            <NavMain :items="mainNavItems" />
        </SidebarContent>

        <SidebarFooter>
            <NavFooter :items="footerNavItems" />
            <NavUser />
        </SidebarFooter>
    </Sidebar>
    <slot />
</template>
