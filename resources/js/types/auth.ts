export type User = {
    id: number;
    name: string;
    email: string;
    img_url?: string | null;
    email_verified_at: string | null;
    created_at: string;
    updated_at: string;
    [key: string]: unknown;
};

export type Auth = {
    user: User;
    canViewHorizon: boolean;
    canViewTelescope: boolean;
};

export type ApiToken = {
    id: number;
    name: string;
    created_at_diff: string;
    last_used_at_diff: string | null;
};

export type Proxy = {
    id: number;
    name: string;
    scheme: 'http' | 'https' | 'socks5';
    host: string;
    port: number | null;
    user: string | null;
    bypass: string | null;
    created_at_diff: string | null;
};

/* @chisel-passkeys */
export type Passkey = {
    id: number;
    name: string;
    authenticator: string | null;
    created_at_diff: string;
    last_used_at_diff: string | null;
};
/* @end-chisel-passkeys */
