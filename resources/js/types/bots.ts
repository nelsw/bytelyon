export type BotRow = {
    id: number;
    childId: number;
    query: string;
    type: string;
    frequency: string;
    blacklist: string;
    enabled: boolean;
    headless: boolean;
    processedAt: string | null;
    createdAt: string;
    updatedAt: string;
    pageCount: number;
};

export type BotFormData = {
    id?: number;
    query: string;
    type: string;
    frequency: string;
    blacklist: string;
    enabled: boolean;
    headless: boolean;
};

export type BotFilters = {
    query: string;
    type: string;
    status: string;
    mode: string;
    sort: string;
    perPage: number;
};

export type PaginationLink = {
    url: string | null;
    label: string;
    active: boolean;
};

export type PaginatedBots = {
    data: BotRow[];
    current_page: number;
    last_page: number;
    per_page: number;
    total: number;
    from: number | null;
    to: number | null;
    prev_page_url: string | null;
    next_page_url: string | null;
    links: PaginationLink[];
};
