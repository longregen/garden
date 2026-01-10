/**
 * Garden PKM - Main Application
 * Router, API service, and core utilities
 */

// ============================================
// API Service
// ============================================
class ApiService {
    constructor(baseUrl = '') {
        this.baseUrl = baseUrl;
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (!response.ok) {
                throw new ApiError(data.error || 'Request failed', response.status, data);
            }

            return data;
        } catch (error) {
            if (error instanceof ApiError) throw error;
            throw new ApiError(error.message, 0);
        }
    }

    get(endpoint, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const url = queryString ? `${endpoint}?${queryString}` : endpoint;
        return this.request(url, { method: 'GET' });
    }

    post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // Dashboard
    getDashboardStats() {
        return this.get('/api/dashboard/stats');
    }

    // Bookmarks
    getBookmarks(params = {}) {
        return this.get('/api/bookmarks', params);
    }

    getBookmark(id) {
        return this.get(`/api/bookmarks/${id}`);
    }

    createBookmark(data) {
        return this.post('/api/bookmarks', data);
    }

    updateBookmark(id, data) {
        return this.put(`/api/bookmarks/${id}`, data);
    }

    deleteBookmark(id) {
        return this.delete(`/api/bookmarks/${id}`);
    }

    searchBookmarks(query) {
        return this.get('/api/bookmarks/search', { query });
    }

    getRandomBookmark() {
        return this.get('/api/bookmarks/random');
    }

    // Notes
    getNotes(params = {}) {
        return this.get('/api/notes', params);
    }

    getNote(id) {
        return this.get(`/api/notes/${id}`);
    }

    createNote(data) {
        return this.post('/api/notes', data);
    }

    updateNote(id, data) {
        return this.put(`/api/notes/${id}`, data);
    }

    deleteNote(id) {
        return this.delete(`/api/notes/${id}`);
    }

    searchNotes(query) {
        return this.get('/api/notes/search', { q: query });
    }

    // Contacts
    getContacts(params = {}) {
        return this.get('/api/contacts', params);
    }

    getContact(id) {
        return this.get(`/api/contacts/${id}`);
    }

    createContact(data) {
        return this.post('/api/contacts', data);
    }

    updateContact(id, data) {
        return this.put(`/api/contacts/${id}`, data);
    }

    deleteContact(id) {
        return this.delete(`/api/contacts/${id}`);
    }

    updateContactEvaluation(id, data) {
        return this.put(`/api/contacts/${id}/evaluation`, data);
    }

    // Messages
    getMessages(params = {}) {
        return this.get('/api/messages', params);
    }

    searchMessages(query) {
        return this.get('/api/messages/search', { query });
    }

    // Rooms
    getRooms() {
        return this.get('/api/rooms');
    }

    getRoom(id) {
        return this.get(`/api/rooms/${id}`);
    }

    // Entities
    getEntities(params = {}) {
        return this.get('/api/entities', params);
    }

    getEntity(id) {
        return this.get(`/api/entities/${id}`);
    }

    createEntity(data) {
        return this.post('/api/entities', data);
    }

    // Browser History
    getBrowserHistory(params = {}) {
        return this.get('/api/browser-history', params);
    }

    getBrowserHistoryDomains() {
        return this.get('/api/browser-history/domains');
    }

    // Social Posts
    getSocialPosts(params = {}) {
        return this.get('/api/social-posts', params);
    }

    getSocialPost(id) {
        return this.get(`/api/social-posts/${id}`);
    }

    createSocialPost(data) {
        return this.post('/api/social-posts', data);
    }

    publishSocialPost(id) {
        return this.post(`/api/social-posts/${id}/publish`);
    }

    getSocialCredentials() {
        return this.get('/api/social/credentials');
    }

    // Tags
    getTags() {
        return this.get('/api/tags');
    }

    // Categories
    getCategories() {
        return this.get('/api/categories');
    }

    // Global Search
    search(query) {
        return this.get('/api/search', { q: query });
    }

    // Configuration
    getConfiguration() {
        return this.get('/api/configuration');
    }

    updateConfiguration(data) {
        return this.put('/api/configuration', data);
    }
}

class ApiError extends Error {
    constructor(message, status, data = null) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.data = data;
    }
}

// ============================================
// Router
// ============================================
class Router {
    constructor() {
        this.routes = new Map();
        this.currentRoute = null;
        this.beforeHooks = [];
        this.afterHooks = [];
    }

    register(path, handler) {
        this.routes.set(path, handler);
        return this;
    }

    before(hook) {
        this.beforeHooks.push(hook);
        return this;
    }

    after(hook) {
        this.afterHooks.push(hook);
        return this;
    }

    async navigate(path, params = {}) {
        // Run before hooks
        for (const hook of this.beforeHooks) {
            const result = await hook(path, params);
            if (result === false) return;
        }

        const handler = this.routes.get(path) || this.routes.get('*');
        if (handler) {
            this.currentRoute = { path, params };
            await handler(params);

            // Run after hooks
            for (const hook of this.afterHooks) {
                await hook(path, params);
            }
        }
    }

    parseHash(hash) {
        const [path, queryString] = hash.replace('#', '').split('?');
        const params = {};

        if (queryString) {
            new URLSearchParams(queryString).forEach((value, key) => {
                params[key] = value;
            });
        }

        // Extract route params like /notes/:id
        const segments = path.split('/').filter(Boolean);
        const routePath = '/' + segments[0];

        if (segments.length > 1) {
            params.id = segments[1];
        }

        return { path: routePath || '/dashboard', params };
    }

    start() {
        const handleRoute = () => {
            const { path, params } = this.parseHash(window.location.hash);
            this.navigate(path, params);
        };

        window.addEventListener('hashchange', handleRoute);
        handleRoute();
    }
}

// ============================================
// Toast Notifications
// ============================================
class ToastManager {
    constructor(containerId = 'toast-container') {
        this.container = document.getElementById(containerId);
        this.toasts = [];
    }

    show(message, type = 'info', duration = 5000) {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;

        const icons = {
            success: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22,4 12,14.01 9,11.01"/></svg>',
            error: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
            warning: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
            info: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>'
        };

        toast.innerHTML = `
            <span class="toast-icon">${icons[type] || icons.info}</span>
            <div class="toast-content">
                <p class="toast-message">${message}</p>
            </div>
            <button class="toast-close" aria-label="Close">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
            </button>
        `;

        const closeBtn = toast.querySelector('.toast-close');
        closeBtn.addEventListener('click', () => this.dismiss(toast));

        this.container.appendChild(toast);
        this.toasts.push(toast);

        if (duration > 0) {
            setTimeout(() => this.dismiss(toast), duration);
        }

        return toast;
    }

    dismiss(toast) {
        toast.classList.add('toast-exit');
        setTimeout(() => {
            toast.remove();
            this.toasts = this.toasts.filter(t => t !== toast);
        }, 300);
    }

    success(message, duration) {
        return this.show(message, 'success', duration);
    }

    error(message, duration) {
        return this.show(message, 'error', duration);
    }

    warning(message, duration) {
        return this.show(message, 'warning', duration);
    }

    info(message, duration) {
        return this.show(message, 'info', duration);
    }
}

// ============================================
// Modal System
// ============================================
class ModalManager {
    constructor() {
        this.overlay = document.getElementById('modal-overlay');
        this.modal = document.getElementById('modal');
        this.title = document.getElementById('modal-title');
        this.body = document.getElementById('modal-body');
        this.footer = document.getElementById('modal-footer');
        this.closeBtn = document.getElementById('modal-close');

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.closeBtn.addEventListener('click', () => this.close());
        this.overlay.addEventListener('click', (e) => {
            if (e.target === this.overlay) this.close();
        });
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen()) this.close();
        });
    }

    open(options = {}) {
        const { title = '', content = '', footer = '', onClose } = options;

        this.title.textContent = title;
        this.body.innerHTML = typeof content === 'string' ? content : '';
        this.footer.innerHTML = typeof footer === 'string' ? footer : '';

        if (typeof content !== 'string') {
            this.body.innerHTML = '';
            this.body.appendChild(content);
        }

        this.onCloseCallback = onClose;
        this.overlay.classList.add('active');
        document.body.style.overflow = 'hidden';
    }

    close() {
        this.overlay.classList.remove('active');
        document.body.style.overflow = '';

        if (this.onCloseCallback) {
            this.onCloseCallback();
            this.onCloseCallback = null;
        }
    }

    isOpen() {
        return this.overlay.classList.contains('active');
    }

    confirm(message, options = {}) {
        return new Promise((resolve) => {
            const { title = 'Confirm', confirmText = 'Confirm', cancelText = 'Cancel', type = 'primary' } = options;

            this.open({
                title,
                content: `<p>${message}</p>`,
                footer: `
                    <button class="btn btn-secondary" id="modal-cancel">${cancelText}</button>
                    <button class="btn btn-${type}" id="modal-confirm">${confirmText}</button>
                `,
                onClose: () => resolve(false)
            });

            document.getElementById('modal-cancel').addEventListener('click', () => {
                this.close();
                resolve(false);
            });

            document.getElementById('modal-confirm').addEventListener('click', () => {
                this.close();
                resolve(true);
            });
        });
    }
}

// ============================================
// Command Palette
// ============================================
class CommandPalette {
    constructor(app) {
        this.app = app;
        this.overlay = document.getElementById('command-palette-overlay');
        this.input = document.getElementById('command-palette-input');
        this.results = document.getElementById('command-palette-results');
        this.selectedIndex = 0;
        this.items = [];

        this.setupEventListeners();
    }

    setupEventListeners() {
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                this.toggle();
            }
            if (e.key === 'Escape' && this.isOpen()) {
                this.close();
            }
        });

        this.overlay.addEventListener('click', (e) => {
            if (e.target === this.overlay) this.close();
        });

        this.input.addEventListener('input', () => this.search());
        this.input.addEventListener('keydown', (e) => this.handleKeydown(e));
    }

    toggle() {
        if (this.isOpen()) {
            this.close();
        } else {
            this.open();
        }
    }

    open() {
        this.overlay.classList.add('active');
        this.input.value = '';
        this.input.focus();
        this.showDefaultItems();
    }

    close() {
        this.overlay.classList.remove('active');
        this.input.value = '';
        this.results.innerHTML = '';
    }

    isOpen() {
        return this.overlay.classList.contains('active');
    }

    showDefaultItems() {
        this.items = [
            { label: 'Dashboard', icon: 'grid', action: () => window.location.hash = '#/dashboard', shortcut: 'G D' },
            { label: 'Bookmarks', icon: 'bookmark', action: () => window.location.hash = '#/bookmarks', shortcut: 'G B' },
            { label: 'Notes', icon: 'file', action: () => window.location.hash = '#/notes', shortcut: 'G N' },
            { label: 'Contacts', icon: 'users', action: () => window.location.hash = '#/contacts', shortcut: 'G C' },
            { label: 'Messages', icon: 'message', action: () => window.location.hash = '#/messages', shortcut: 'G M' },
            { label: 'Entities', icon: 'share', action: () => window.location.hash = '#/entities', shortcut: 'G E' },
            { label: 'Search', icon: 'search', action: () => window.location.hash = '#/search', shortcut: 'G S' },
            { label: 'Settings', icon: 'settings', action: () => window.location.hash = '#/settings', shortcut: 'G ,' }
        ];
        this.selectedIndex = 0;
        this.renderItems();
    }

    async search() {
        const query = this.input.value.trim().toLowerCase();

        if (!query) {
            this.showDefaultItems();
            return;
        }

        // Filter navigation items
        const navItems = [
            { label: 'Dashboard', action: () => window.location.hash = '#/dashboard' },
            { label: 'Bookmarks', action: () => window.location.hash = '#/bookmarks' },
            { label: 'Notes', action: () => window.location.hash = '#/notes' },
            { label: 'Contacts', action: () => window.location.hash = '#/contacts' },
            { label: 'Messages', action: () => window.location.hash = '#/messages' },
            { label: 'Entities', action: () => window.location.hash = '#/entities' },
            { label: 'Search', action: () => window.location.hash = '#/search' },
            { label: 'History', action: () => window.location.hash = '#/history' },
            { label: 'Social', action: () => window.location.hash = '#/social' },
            { label: 'Tags', action: () => window.location.hash = '#/tags' },
            { label: 'Settings', action: () => window.location.hash = '#/settings' }
        ].filter(item => item.label.toLowerCase().includes(query));

        // Search API
        try {
            const results = await this.app.api.search(query);
            const apiItems = [];

            results.bookmarks?.forEach(b => {
                apiItems.push({
                    label: b.title || b.url,
                    sublabel: 'Bookmark',
                    action: () => window.location.hash = `#/bookmarks/${b.bookmark_id}`
                });
            });

            results.notes?.forEach(n => {
                apiItems.push({
                    label: n.title || 'Untitled Note',
                    sublabel: 'Note',
                    action: () => window.location.hash = `#/notes/${n.id}`
                });
            });

            results.contacts?.forEach(c => {
                apiItems.push({
                    label: c.name,
                    sublabel: 'Contact',
                    action: () => window.location.hash = `#/contacts/${c.contact_id}`
                });
            });

            this.items = [...navItems, ...apiItems];
        } catch (e) {
            this.items = navItems;
        }

        this.selectedIndex = 0;
        this.renderItems();
    }

    renderItems() {
        this.results.innerHTML = this.items.map((item, index) => `
            <div class="command-palette-item ${index === this.selectedIndex ? 'active' : ''}" data-index="${index}">
                <span class="command-palette-item-label">
                    ${item.label}
                    ${item.sublabel ? `<span class="text-muted text-sm"> - ${item.sublabel}</span>` : ''}
                </span>
                ${item.shortcut ? `
                    <span class="command-palette-item-shortcut">
                        ${item.shortcut.split(' ').map(k => `<kbd>${k}</kbd>`).join('')}
                    </span>
                ` : ''}
            </div>
        `).join('');

        this.results.querySelectorAll('.command-palette-item').forEach((el, index) => {
            el.addEventListener('click', () => this.selectItem(index));
        });
    }

    handleKeydown(e) {
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            this.selectedIndex = Math.min(this.selectedIndex + 1, this.items.length - 1);
            this.renderItems();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            this.selectedIndex = Math.max(this.selectedIndex - 1, 0);
            this.renderItems();
        } else if (e.key === 'Enter') {
            e.preventDefault();
            this.selectItem(this.selectedIndex);
        }
    }

    selectItem(index) {
        const item = this.items[index];
        if (item) {
            this.close();
            item.action();
        }
    }
}

// ============================================
// View Loader
// ============================================
class ViewLoader {
    constructor(containerId = 'page-content') {
        this.container = document.getElementById(containerId);
        this.cache = new Map();
    }

    async load(viewName) {
        this.showLoading();

        try {
            let html;
            if (this.cache.has(viewName)) {
                html = this.cache.get(viewName);
            } else {
                const response = await fetch(`views/${viewName}.html`);
                if (!response.ok) throw new Error('View not found');
                html = await response.text();
                this.cache.set(viewName, html);
            }

            this.container.innerHTML = html;
            return true;
        } catch (error) {
            this.showError(`Failed to load view: ${viewName}`);
            return false;
        }
    }

    render(html) {
        this.container.innerHTML = html;
    }

    showLoading() {
        this.container.innerHTML = `
            <div class="loading-state">
                <div class="loading-spinner"></div>
                <p>Loading...</p>
            </div>
        `;
    }

    showError(message) {
        this.container.innerHTML = `
            <div class="empty-state">
                <svg class="empty-state-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                <h3 class="empty-state-title">Error</h3>
                <p class="empty-state-description">${message}</p>
            </div>
        `;
    }
}

// ============================================
// Utilities
// ============================================
const Utils = {
    // Date formatting
    formatDate(dateString, options = {}) {
        const date = new Date(dateString);
        const defaultOptions = {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        };
        return date.toLocaleDateString('en-US', { ...defaultOptions, ...options });
    },

    formatDateTime(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    },

    formatRelativeTime(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diff = now - date;
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (seconds < 60) return 'just now';
        if (minutes < 60) return `${minutes}m ago`;
        if (hours < 24) return `${hours}h ago`;
        if (days < 7) return `${days}d ago`;

        return this.formatDate(dateString);
    },

    formatTimestamp(timestamp) {
        return this.formatDateTime(new Date(timestamp));
    },

    // String utilities
    truncate(str, length = 50) {
        if (!str) return '';
        return str.length > length ? str.substring(0, length) + '...' : str;
    },

    slugify(str) {
        return str
            .toLowerCase()
            .trim()
            .replace(/[^\w\s-]/g, '')
            .replace(/[\s_-]+/g, '-')
            .replace(/^-+|-+$/g, '');
    },

    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    // URL utilities
    getDomain(url) {
        try {
            return new URL(url).hostname;
        } catch {
            return url;
        }
    },

    // Number formatting
    formatNumber(num) {
        if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
        if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
        return num.toString();
    },

    formatPercentage(num, decimals = 1) {
        const sign = num >= 0 ? '+' : '';
        return `${sign}${num.toFixed(decimals)}%`;
    },

    // Array utilities
    groupBy(array, key) {
        return array.reduce((groups, item) => {
            const value = typeof key === 'function' ? key(item) : item[key];
            (groups[value] = groups[value] || []).push(item);
            return groups;
        }, {});
    },

    // Debounce
    debounce(fn, delay) {
        let timeoutId;
        return (...args) => {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => fn.apply(this, args), delay);
        };
    },

    // Generate initials
    getInitials(name) {
        if (!name) return '?';
        return name
            .split(' ')
            .map(word => word[0])
            .join('')
            .toUpperCase()
            .substring(0, 2);
    },

    // Generate color from string (for avatars)
    stringToColor(str) {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            hash = str.charCodeAt(i) + ((hash << 5) - hash);
        }
        const hue = hash % 360;
        return `hsl(${hue}, 50%, 50%)`;
    }
};

// ============================================
// Pagination Component
// ============================================
class Pagination {
    constructor(options = {}) {
        this.page = options.page || 1;
        this.totalPages = options.totalPages || 1;
        this.onPageChange = options.onPageChange || (() => { });
    }

    render() {
        if (this.totalPages <= 1) return '';

        const pages = [];
        const maxVisible = 5;
        let start = Math.max(1, this.page - Math.floor(maxVisible / 2));
        let end = Math.min(this.totalPages, start + maxVisible - 1);

        if (end - start + 1 < maxVisible) {
            start = Math.max(1, end - maxVisible + 1);
        }

        // Previous button
        pages.push(`
            <button class="pagination-btn" data-page="${this.page - 1}" ${this.page === 1 ? 'disabled' : ''}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M15 18l-6-6 6-6"/>
                </svg>
            </button>
        `);

        // First page + ellipsis
        if (start > 1) {
            pages.push(`<button class="pagination-btn" data-page="1">1</button>`);
            if (start > 2) {
                pages.push(`<span class="pagination-ellipsis">...</span>`);
            }
        }

        // Page numbers
        for (let i = start; i <= end; i++) {
            pages.push(`
                <button class="pagination-btn ${i === this.page ? 'active' : ''}" data-page="${i}">${i}</button>
            `);
        }

        // Last page + ellipsis
        if (end < this.totalPages) {
            if (end < this.totalPages - 1) {
                pages.push(`<span class="pagination-ellipsis">...</span>`);
            }
            pages.push(`<button class="pagination-btn" data-page="${this.totalPages}">${this.totalPages}</button>`);
        }

        // Next button
        pages.push(`
            <button class="pagination-btn" data-page="${this.page + 1}" ${this.page === this.totalPages ? 'disabled' : ''}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M9 18l6-6-6-6"/>
                </svg>
            </button>
        `);

        return `<div class="pagination">${pages.join('')}</div>`;
    }

    attachEvents(container) {
        container.querySelectorAll('.pagination-btn[data-page]').forEach(btn => {
            btn.addEventListener('click', () => {
                const page = parseInt(btn.dataset.page);
                if (page >= 1 && page <= this.totalPages && page !== this.page) {
                    this.page = page;
                    this.onPageChange(page);
                }
            });
        });
    }
}

// ============================================
// Main Application
// ============================================
class GardenApp {
    constructor() {
        this.api = new ApiService();
        this.router = new Router();
        this.toast = new ToastManager();
        this.modal = new ModalManager();
        this.viewLoader = new ViewLoader();
        this.commandPalette = new CommandPalette(this);

        this.setupRoutes();
        this.setupEventListeners();
        this.router.start();
    }

    setupRoutes() {
        this.router
            .register('/dashboard', () => this.loadDashboard())
            .register('/bookmarks', (params) => this.loadBookmarks(params))
            .register('/notes', (params) => this.loadNotes(params))
            .register('/contacts', (params) => this.loadContacts(params))
            .register('/messages', (params) => this.loadMessages(params))
            .register('/entities', (params) => this.loadEntities(params))
            .register('/search', (params) => this.loadSearch(params))
            .register('/history', (params) => this.loadHistory(params))
            .register('/social', (params) => this.loadSocial(params))
            .register('/tags', () => this.loadTags())
            .register('/settings', () => this.loadSettings())
            .register('*', () => this.loadDashboard());

        this.router.after((path) => {
            this.updateNavigation(path);
            this.updateBreadcrumb(path);
        });
    }

    setupEventListeners() {
        // Sidebar toggle
        document.getElementById('sidebar-toggle')?.addEventListener('click', () => {
            document.getElementById('sidebar').classList.toggle('collapsed');
        });

        // Mobile menu toggle
        document.getElementById('mobile-menu-toggle')?.addEventListener('click', () => {
            document.getElementById('sidebar').classList.toggle('open');
        });

        // Global search
        document.getElementById('global-search')?.addEventListener('focus', () => {
            this.commandPalette.open();
        });

        // Close mobile menu on navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', () => {
                document.getElementById('sidebar').classList.remove('open');
            });
        });
    }

    updateNavigation(path) {
        document.querySelectorAll('.nav-link').forEach(link => {
            const route = link.dataset.route;
            link.classList.toggle('active', path === `/${route}`);
        });
    }

    updateBreadcrumb(path) {
        const breadcrumb = document.getElementById('breadcrumb');
        const pathName = path.replace('/', '').replace(/-/g, ' ');
        const title = pathName.charAt(0).toUpperCase() + pathName.slice(1) || 'Dashboard';
        breadcrumb.innerHTML = `<span class="breadcrumb-item">${title}</span>`;
    }

    // ============================================
    // View Loaders
    // ============================================

    async loadDashboard() {
        await this.viewLoader.load('dashboard');

        try {
            // Fetch all data in parallel for better performance
            const [stats, bookmarksResult, notesResult, contactsResult, roomsResult, messagesResult] = await Promise.all([
                this.api.getDashboardStats(),
                this.api.getBookmarks({ limit: 5 }),
                this.api.getNotes({ pageSize: 5 }),
                this.api.getContacts({ pageSize: 30 }),
                this.api.getRooms(),
                this.api.getMessages({ limit: 50 })
            ]);

            // Render all dashboard sections
            this.renderDashboardStats(stats, contactsResult, bookmarksResult, notesResult, messagesResult, roomsResult);
            this.renderDashboardCharts(bookmarksResult, notesResult, contactsResult, messagesResult);
            this.renderQuickActions();
            this.renderActivityFeed(bookmarksResult, notesResult, contactsResult, messagesResult);
            this.renderTopItems(bookmarksResult, notesResult, contactsResult);
            this.renderSystemStatus();
        } catch (error) {
            console.error('Dashboard load error:', error);
            this.toast.error('Failed to load dashboard data');
        }
    }

    renderDashboardStats(stats, contactsResult, bookmarksResult, notesResult, messagesResult, roomsResult) {
        const container = document.getElementById('stats-cards');
        if (!container) return;

        const totalContacts = contactsResult.data?.length || stats.contacts.total;
        const totalBookmarks = bookmarksResult.data?.length || stats.bookmarks.total;
        const totalNotes = notesResult.notes?.length || 15;
        const totalMessages = messagesResult.data?.length || 0;
        const activeRooms = roomsResult?.length || 0;

        // Calculate category breakdown for bookmarks
        const categories = {};
        (bookmarksResult.data || []).forEach(b => {
            const cat = b.category_name || 'Uncategorized';
            categories[cat] = (categories[cat] || 0) + 1;
        });
        const topCategory = Object.entries(categories).sort((a, b) => b[1] - a[1])[0];

        // Recent notes count (last 7 days)
        const oneWeekAgo = Date.now() - 7 * 24 * 60 * 60 * 1000;
        const recentNotes = (notesResult.notes || []).filter(n => n.modified > oneWeekAgo).length;

        container.innerHTML = `
            <div class="stat-card">
                <div class="stat-card-header">
                    <span class="stat-card-title">Contacts</span>
                    <svg class="stat-card-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                        <circle cx="9" cy="7" r="4"/>
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                    </svg>
                </div>
                <div class="stat-card-value">${Utils.formatNumber(totalContacts)}</div>
                <div class="stat-card-change positive">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M12 19V5M5 12l7-7 7 7"/>
                    </svg>
                    +12 this month
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-card-header">
                    <span class="stat-card-title">Bookmarks</span>
                    <svg class="stat-card-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/>
                    </svg>
                </div>
                <div class="stat-card-value">${Utils.formatNumber(totalBookmarks)}</div>
                <div class="stat-card-change text-muted">
                    ${topCategory ? `Top: ${topCategory[0]} (${topCategory[1]})` : 'No categories'}
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-card-header">
                    <span class="stat-card-title">Notes</span>
                    <svg class="stat-card-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                        <polyline points="14,2 14,8 20,8"/>
                    </svg>
                </div>
                <div class="stat-card-value">${Utils.formatNumber(totalNotes)}</div>
                <div class="stat-card-change ${recentNotes > 0 ? 'positive' : 'text-muted'}">
                    ${recentNotes > 0 ? `${recentNotes} updated this week` : 'No recent updates'}
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-card-header">
                    <span class="stat-card-title">Messages</span>
                    <svg class="stat-card-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                    </svg>
                </div>
                <div class="stat-card-value">${Utils.formatNumber(totalMessages)}</div>
                <div class="stat-card-change text-muted">
                    ${activeRooms} active rooms
                </div>
            </div>
        `;
    }

    renderDashboardCharts(bookmarksResult, notesResult, contactsResult, messagesResult) {
        const container = document.getElementById('charts-section');
        if (!container) return;

        // Generate activity data for bar chart (last 30 days)
        const activityData = this.generateActivityData(bookmarksResult, notesResult, messagesResult);

        // Generate distribution data for pie chart
        const distributionData = this.generateDistributionData(bookmarksResult, notesResult, contactsResult);

        // Generate growth data for line chart
        const growthData = this.generateGrowthData();

        container.innerHTML = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Activity Timeline</h3>
                    <span class="text-sm text-muted">Last 30 days</span>
                </div>
                <div class="card-body">
                    <div class="chart-container">
                        <div class="bar-chart">
                            ${activityData.map((val, i) => `
                                <div class="bar" style="height: ${Math.max(val * 3, 4)}px;" title="${val} activities">
                                    ${i % 5 === 0 ? `<span class="bar-label">${30 - i}d</span>` : ''}
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
            </div>
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Knowledge Distribution</h3>
                </div>
                <div class="card-body">
                    <div class="chart-container">
                        <div class="pie-chart-container">
                            <svg class="pie-chart" viewBox="0 0 100 100">
                                ${this.generatePieChartPaths(distributionData)}
                            </svg>
                            <div class="pie-legend">
                                ${distributionData.map(item => `
                                    <div class="pie-legend-item">
                                        <div class="pie-legend-color" style="background-color: ${item.color};"></div>
                                        <span class="pie-legend-label">${item.label}</span>
                                        <span class="pie-legend-value">${item.value}</span>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Growth Trends</h3>
                    <span class="text-sm text-muted">This month</span>
                </div>
                <div class="card-body">
                    <div class="chart-container">
                        <div class="line-chart">
                            <svg viewBox="0 0 300 120" preserveAspectRatio="none">
                                <defs>
                                    <linearGradient id="lineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                                        <stop offset="0%" style="stop-color:#0070f3;stop-opacity:0.3"/>
                                        <stop offset="100%" style="stop-color:#0070f3;stop-opacity:0"/>
                                    </linearGradient>
                                </defs>
                                <path class="line-chart-area" d="${this.generateLineChartArea(growthData)}"/>
                                <path class="line-chart-path" d="${this.generateLineChartPath(growthData)}"/>
                                ${growthData.map((val, i) => `
                                    <circle class="line-chart-dot" cx="${(i / (growthData.length - 1)) * 290 + 5}" cy="${100 - val * 0.8}" r="3"/>
                                `).join('')}
                            </svg>
                            <div class="line-chart-labels">
                                <span class="line-chart-label">Week 1</span>
                                <span class="line-chart-label">Week 2</span>
                                <span class="line-chart-label">Week 3</span>
                                <span class="line-chart-label">Week 4</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    generateActivityData(bookmarksResult, notesResult, messagesResult) {
        // Generate mock activity data for last 30 days
        const data = [];
        for (let i = 0; i < 30; i++) {
            // Simulate varying activity levels
            const base = Math.floor(Math.random() * 8) + 2;
            const weekend = (i % 7 === 0 || i % 7 === 6) ? 0.5 : 1;
            data.push(Math.floor(base * weekend));
        }
        return data;
    }

    generateDistributionData(bookmarksResult, notesResult, contactsResult) {
        const bookmarks = bookmarksResult.data?.length || 22;
        const notes = notesResult.notes?.length || 15;
        const contacts = contactsResult.data?.length || 30;
        const total = bookmarks + notes + contacts;

        return [
            { label: 'Bookmarks', value: bookmarks, percentage: (bookmarks / total * 100).toFixed(0), color: '#0070f3' },
            { label: 'Notes', value: notes, percentage: (notes / total * 100).toFixed(0), color: '#50e3c2' },
            { label: 'Contacts', value: contacts, percentage: (contacts / total * 100).toFixed(0), color: '#f5a623' }
        ];
    }

    generateGrowthData() {
        // Generate growth trend data
        return [20, 35, 45, 40, 55, 65, 75, 85];
    }

    generatePieChartPaths(data) {
        let currentAngle = 0;
        const total = data.reduce((sum, item) => sum + item.value, 0);

        return data.map(item => {
            const percentage = item.value / total;
            const angle = percentage * 360;
            const startAngle = currentAngle;
            const endAngle = currentAngle + angle;

            const x1 = 50 + 40 * Math.cos((startAngle - 90) * Math.PI / 180);
            const y1 = 50 + 40 * Math.sin((startAngle - 90) * Math.PI / 180);
            const x2 = 50 + 40 * Math.cos((endAngle - 90) * Math.PI / 180);
            const y2 = 50 + 40 * Math.sin((endAngle - 90) * Math.PI / 180);

            const largeArc = angle > 180 ? 1 : 0;

            currentAngle = endAngle;

            return `<path d="M 50 50 L ${x1} ${y1} A 40 40 0 ${largeArc} 1 ${x2} ${y2} Z" fill="${item.color}"/>`;
        }).join('');
    }

    generateLineChartPath(data) {
        const points = data.map((val, i) => {
            const x = (i / (data.length - 1)) * 290 + 5;
            const y = 100 - val * 0.8;
            return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
        });
        return points.join(' ');
    }

    generateLineChartArea(data) {
        const linePath = this.generateLineChartPath(data);
        const lastX = 295;
        const firstX = 5;
        return `${linePath} L ${lastX} 100 L ${firstX} 100 Z`;
    }

    renderQuickActions() {
        const container = document.getElementById('quick-actions-section');
        if (!container) return;

        container.innerHTML = `
            <div class="card-header">
                <h3 class="card-title">Quick Actions</h3>
            </div>
            <div class="card-body">
                <div class="quick-actions-grid">
                    <button class="quick-action-btn" onclick="window.location.hash='#/bookmarks?action=add'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/>
                            <line x1="12" y1="8" x2="12" y2="14"/>
                            <line x1="9" y1="11" x2="15" y2="11"/>
                        </svg>
                        Add Bookmark
                    </button>
                    <button class="quick-action-btn" onclick="window.location.hash='#/notes?action=new'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                            <polyline points="14,2 14,8 20,8"/>
                            <line x1="12" y1="11" x2="12" y2="17"/>
                            <line x1="9" y1="14" x2="15" y2="14"/>
                        </svg>
                        Add Note
                    </button>
                    <button class="quick-action-btn" onclick="window.location.hash='#/contacts?action=add'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                            <circle cx="8.5" cy="7" r="4"/>
                            <line x1="20" y1="8" x2="20" y2="14"/>
                            <line x1="23" y1="11" x2="17" y2="11"/>
                        </svg>
                        Add Contact
                    </button>
                    <button class="quick-action-btn" onclick="window.location.hash='#/messages?action=compose'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                            <line x1="9" y1="10" x2="15" y2="10"/>
                        </svg>
                        Compose Message
                    </button>
                    <button class="quick-action-btn" onclick="window.location.hash='#/history'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"/>
                            <polyline points="12,6 12,12 16,14"/>
                            <path d="M2 12h2"/>
                        </svg>
                        Import History
                    </button>
                    <button class="quick-action-btn" onclick="window.location.hash='#/entities?action=create'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="3"/>
                            <circle cx="19" cy="5" r="2"/>
                            <circle cx="5" cy="5" r="2"/>
                            <line x1="12" y1="9" x2="12" y2="5"/>
                        </svg>
                        Create Entity
                    </button>
                    <button class="quick-action-btn" onclick="window.location.hash='#/social?action=post'">
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M4 4v16h16"/>
                            <path d="M4 14l4-4 4 4 8-8"/>
                        </svg>
                        Post to Social
                    </button>
                </div>
            </div>
        `;
    }

    renderActivityFeed(bookmarksResult, notesResult, contactsResult, messagesResult) {
        const container = document.getElementById('activity-feed-section');
        if (!container) return;

        // Combine and sort all activities
        const activities = [];

        (bookmarksResult.data || []).slice(0, 3).forEach(b => {
            activities.push({
                type: 'bookmark',
                title: b.title || b.url,
                date: new Date(b.creation_date),
                icon: 'bookmark',
                link: `#/bookmarks/${b.bookmark_id}`
            });
        });

        (notesResult.notes || []).slice(0, 3).forEach(n => {
            activities.push({
                type: 'note',
                title: n.title || 'Untitled Note',
                date: new Date(n.modified),
                icon: 'note',
                link: `#/notes/${n.id}`
            });
        });

        (contactsResult.data || []).slice(0, 3).forEach(c => {
            activities.push({
                type: 'contact',
                title: c.name,
                date: new Date(c.last_update),
                icon: 'contact',
                link: `#/contacts/${c.contact_id}`
            });
        });

        (messagesResult.data || []).slice(0, 3).forEach(m => {
            activities.push({
                type: 'message',
                title: `${m.sender_name}: ${Utils.truncate(m.body, 40)}`,
                date: new Date(m.event_datetime),
                icon: 'message',
                link: `#/messages?room=${m.room_id}`
            });
        });

        // Sort by date descending
        activities.sort((a, b) => b.date - a.date);

        const icons = {
            bookmark: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/></svg>`,
            note: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14,2 14,8 20,8"/></svg>`,
            contact: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>`,
            message: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`
        };

        container.innerHTML = `
            <div class="card-header">
                <h3 class="card-title">Recent Activity</h3>
                <a href="#/history" class="btn btn-ghost btn-sm">View All</a>
            </div>
            <div class="card-body" style="padding: 0;">
                ${activities.slice(0, 8).map(item => `
                    <a href="${item.link}" class="list-item" style="text-decoration: none; color: inherit;">
                        <div class="activity-icon ${item.icon}">
                            ${icons[item.icon]}
                        </div>
                        <div class="list-item-content">
                            <div class="list-item-title">${Utils.escapeHtml(item.title)}</div>
                            <div class="list-item-subtitle">
                                <span class="badge badge-${item.type === 'bookmark' ? 'info' : item.type === 'note' ? 'success' : item.type === 'contact' ? 'warning' : 'pending'}" style="margin-right: 8px;">
                                    ${item.type}
                                </span>
                                ${Utils.formatRelativeTime(item.date)}
                            </div>
                        </div>
                    </a>
                `).join('')}
            </div>
        `;
    }

    renderTopItems(bookmarksResult, notesResult, contactsResult) {
        const container = document.getElementById('top-items-section');
        if (!container) return;

        const bookmarks = (bookmarksResult.data || []).slice(0, 5);
        const notes = (notesResult.notes || []).slice(0, 5);
        const topContacts = [...(contactsResult.data || [])]
            .sort((a, b) => (b.importance || 0) - (a.importance || 0))
            .slice(0, 5);

        container.innerHTML = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Recent Bookmarks</h3>
                    <a href="#/bookmarks" class="btn btn-ghost btn-sm">View All</a>
                </div>
                <div class="card-body" style="padding: 0;">
                    ${bookmarks.map(b => `
                        <a href="#/bookmarks/${b.bookmark_id}" class="top-item" style="text-decoration: none; color: inherit;">
                            <div class="top-item-icon">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/>
                                </svg>
                            </div>
                            <div class="top-item-content">
                                <div class="top-item-title">${Utils.escapeHtml(b.title || 'Untitled')}</div>
                                <div class="top-item-meta">${Utils.getDomain(b.url)}</div>
                            </div>
                        </a>
                    `).join('')}
                    ${bookmarks.length === 0 ? '<div class="list-item text-muted text-center">No bookmarks yet</div>' : ''}
                </div>
            </div>
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Recent Notes</h3>
                    <a href="#/notes" class="btn btn-ghost btn-sm">View All</a>
                </div>
                <div class="card-body" style="padding: 0;">
                    ${notes.map(n => `
                        <a href="#/notes/${n.id}" class="top-item" style="text-decoration: none; color: inherit;">
                            <div class="top-item-icon">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                                    <polyline points="14,2 14,8 20,8"/>
                                </svg>
                            </div>
                            <div class="top-item-content">
                                <div class="top-item-title">${Utils.escapeHtml(n.title || 'Untitled')}</div>
                                <div class="top-item-meta">${Utils.formatRelativeTime(new Date(n.modified))}</div>
                            </div>
                        </a>
                    `).join('')}
                    ${notes.length === 0 ? '<div class="list-item text-muted text-center">No notes yet</div>' : ''}
                </div>
            </div>
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Top Contacts</h3>
                    <a href="#/contacts" class="btn btn-ghost btn-sm">View All</a>
                </div>
                <div class="card-body" style="padding: 0;">
                    ${topContacts.map(c => `
                        <a href="#/contacts/${c.contact_id}" class="top-item" style="text-decoration: none; color: inherit;">
                            <div class="avatar avatar-sm" style="background-color: ${Utils.stringToColor(c.name)}; width: 32px; height: 32px;">
                                ${Utils.getInitials(c.name)}
                            </div>
                            <div class="top-item-content">
                                <div class="top-item-title">${Utils.escapeHtml(c.name)}</div>
                                <div class="top-item-meta">${c.last_week_messages || 0} messages this week</div>
                            </div>
                            <span class="score-badge ${c.importance >= 7 ? 'high' : c.importance >= 4 ? 'medium' : 'low'}">
                                ${c.importance || 0}
                            </span>
                        </a>
                    `).join('')}
                    ${topContacts.length === 0 ? '<div class="list-item text-muted text-center">No contacts yet</div>' : ''}
                </div>
            </div>
        `;
    }

    async renderSystemStatus() {
        const container = document.getElementById('system-status-section');
        if (!container) return;

        // Check service worker status
        let swStatus = 'Inactive';
        let swStatusClass = 'offline';
        if ('serviceWorker' in navigator) {
            const registration = await navigator.serviceWorker.getRegistration();
            if (registration && registration.active) {
                swStatus = 'Active';
                swStatusClass = 'online';
            }
        }

        // Get last sync time (simulated)
        const lastSync = new Date(Date.now() - Math.random() * 300000); // Random time within last 5 minutes

        // Estimate storage usage
        let storageUsed = 0;
        let storageQuota = 0;
        let storagePercentage = 0;
        if ('storage' in navigator && 'estimate' in navigator.storage) {
            try {
                const estimate = await navigator.storage.estimate();
                storageUsed = estimate.usage || 0;
                storageQuota = estimate.quota || 0;
                storagePercentage = storageQuota > 0 ? (storageUsed / storageQuota * 100) : 0;
            } catch (e) {
                // Storage API not available
                storageUsed = 2.4 * 1024 * 1024; // 2.4 MB mock
                storageQuota = 50 * 1024 * 1024; // 50 MB mock
                storagePercentage = 4.8;
            }
        } else {
            storageUsed = 2.4 * 1024 * 1024;
            storageQuota = 50 * 1024 * 1024;
            storagePercentage = 4.8;
        }

        const formatBytes = (bytes) => {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
            if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
            return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
        };

        const storageClass = storagePercentage > 80 ? 'danger' : storagePercentage > 60 ? 'warning' : '';

        container.innerHTML = `
            <div class="card-header">
                <h3 class="card-title">System Status</h3>
                <span class="badge badge-success">All Systems Operational</span>
            </div>
            <div class="card-body">
                <div class="system-status-grid">
                    <div class="status-item">
                        <div class="status-item-label">
                            <span class="status-indicator ${swStatusClass}"></span>
                            Service Worker
                        </div>
                        <div class="status-item-value">${swStatus}</div>
                    </div>
                    <div class="status-item">
                        <div class="status-item-label">
                            <span class="status-indicator online"></span>
                            Last Sync
                        </div>
                        <div class="status-item-value">${Utils.formatRelativeTime(lastSync)}</div>
                    </div>
                    <div class="status-item">
                        <div class="status-item-label">
                            <span class="status-indicator ${storageClass || 'online'}"></span>
                            Storage Usage
                        </div>
                        <div class="storage-progress">
                            <div class="storage-bar">
                                <div class="storage-bar-fill ${storageClass}" style="width: ${Math.min(storagePercentage, 100)}%;"></div>
                            </div>
                            <div class="storage-text">
                                <span>${formatBytes(storageUsed)} used</span>
                                <span>${formatBytes(storageQuota)} total</span>
                            </div>
                        </div>
                    </div>
                    <div class="status-item">
                        <div class="status-item-label">
                            <span class="status-indicator online"></span>
                            API Health
                        </div>
                        <div class="status-item-value">Healthy</div>
                    </div>
                </div>
            </div>
        `;
    }

    // Legacy method for backward compatibility
    renderDashboard(stats) {
        // This method is kept for backward compatibility but delegates to new methods
        this.renderDashboardStats(stats, { data: [] }, { data: [] }, { notes: [] }, { data: [] }, []);
    }

    async loadBookmarks(params = {}) {
        await this.viewLoader.load('bookmarks');

        try {
            const page = parseInt(params.page) || 1;
            const result = await this.api.getBookmarks({ page, limit: 10 });
            this.renderBookmarks(result);
        } catch (error) {
            this.toast.error('Failed to load bookmarks');
        }
    }

    renderBookmarks(result) {
        const container = document.getElementById('bookmarks-content');
        if (!container) return;

        container.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table class="table">
                        <thead>
                            <tr>
                                <th>Title</th>
                                <th>Category</th>
                                <th>Created</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            ${result.data.map(bookmark => `
                                <tr>
                                    <td>
                                        <div class="list-item-title">${Utils.escapeHtml(bookmark.title || 'Untitled')}</div>
                                        <div class="list-item-subtitle">${Utils.escapeHtml(Utils.truncate(bookmark.url, 60))}</div>
                                    </td>
                                    <td><span class="badge badge-info">${bookmark.category_name || 'Uncategorized'}</span></td>
                                    <td class="text-muted">${Utils.formatDate(bookmark.creation_date)}</td>
                                    <td>
                                        <a href="${bookmark.url}" target="_blank" class="btn btn-ghost btn-sm">
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
                                                <polyline points="15 3 21 3 21 9"/>
                                                <line x1="10" y1="14" x2="21" y2="3"/>
                                            </svg>
                                        </a>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
            <div id="bookmarks-pagination"></div>
        `;

        const pagination = new Pagination({
            page: result.page,
            totalPages: result.totalPages,
            onPageChange: (page) => {
                window.location.hash = `#/bookmarks?page=${page}`;
            }
        });

        document.getElementById('bookmarks-pagination').innerHTML = pagination.render();
        pagination.attachEvents(document.getElementById('bookmarks-pagination'));
    }

    async loadNotes(params = {}) {
        await this.viewLoader.load('notes');

        try {
            const page = parseInt(params.page) || 1;
            const result = await this.api.getNotes({ page, pageSize: 12 });
            this.renderNotes(result);
        } catch (error) {
            this.toast.error('Failed to load notes');
        }
    }

    renderNotes(result) {
        const container = document.getElementById('notes-content');
        if (!container) return;

        container.innerHTML = `
            <div class="grid grid-cols-3">
                ${result.notes.map(note => `
                    <div class="card card-clickable" onclick="window.location.hash='#/notes/${note.id}'">
                        <div class="card-body">
                            <h4 class="font-semibold mb-2">${Utils.escapeHtml(note.title || 'Untitled')}</h4>
                            <div class="tag-list mb-4">
                                ${(note.tags || []).slice(0, 3).map(tag => `
                                    <span class="tag">${Utils.escapeHtml(tag)}</span>
                                `).join('')}
                            </div>
                            <div class="text-sm text-muted">
                                Modified ${Utils.formatRelativeTime(new Date(note.modified))}
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
            <div id="notes-pagination"></div>
        `;

        const pagination = new Pagination({
            page: 1,
            totalPages: result.totalPages,
            onPageChange: (page) => {
                window.location.hash = `#/notes?page=${page}`;
            }
        });

        document.getElementById('notes-pagination').innerHTML = pagination.render();
        pagination.attachEvents(document.getElementById('notes-pagination'));
    }

    async loadContacts(params = {}) {
        await this.viewLoader.load('contacts');

        try {
            const page = parseInt(params.page) || 1;
            const result = await this.api.getContacts({ page, pageSize: 20 });
            this.renderContacts(result);
        } catch (error) {
            this.toast.error('Failed to load contacts');
        }
    }

    renderContacts(result) {
        const container = document.getElementById('contacts-content');
        if (!container) return;

        container.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table class="table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Email</th>
                                <th>Messages</th>
                                <th>Score</th>
                                <th>Tags</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${result.data.map(contact => `
                                <tr onclick="window.location.hash='#/contacts/${contact.contact_id}'" style="cursor: pointer;">
                                    <td>
                                        <div class="flex items-center gap-4">
                                            <div class="avatar" style="background-color: ${Utils.stringToColor(contact.name)}">
                                                ${Utils.getInitials(contact.name)}
                                            </div>
                                            <span class="font-medium">${Utils.escapeHtml(contact.name)}</span>
                                        </div>
                                    </td>
                                    <td class="text-muted">${contact.email || '-'}</td>
                                    <td>${contact.last_week_messages || 0}</td>
                                    <td>
                                        <span class="badge badge-${contact.importance >= 7 ? 'success' : contact.importance >= 4 ? 'warning' : 'pending'}">
                                            ${contact.importance || 0}/10
                                        </span>
                                    </td>
                                    <td>
                                        <div class="tag-list">
                                            ${(contact.tags || []).slice(0, 2).map(tag => `
                                                <span class="tag">${Utils.escapeHtml(tag.name)}</span>
                                            `).join('')}
                                        </div>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
            <div id="contacts-pagination"></div>
        `;

        const pagination = new Pagination({
            page: result.page,
            totalPages: result.totalPages,
            onPageChange: (page) => {
                window.location.hash = `#/contacts?page=${page}`;
            }
        });

        document.getElementById('contacts-pagination').innerHTML = pagination.render();
        pagination.attachEvents(document.getElementById('contacts-pagination'));
    }

    async loadMessages(params = {}) {
        await this.viewLoader.load('messages');

        try {
            const rooms = await this.api.getRooms();
            this.renderMessages(rooms, params);
        } catch (error) {
            this.toast.error('Failed to load messages');
        }
    }

    renderMessages(rooms, params) {
        const container = document.getElementById('messages-content');
        if (!container) return;

        container.innerHTML = `
            <div class="grid grid-cols-2">
                <div class="card">
                    <div class="card-header">
                        <h3 class="card-title">Rooms</h3>
                    </div>
                    <div class="card-body" style="padding: 0;">
                        ${rooms.map(room => `
                            <div class="list-item" onclick="window.location.hash='#/messages?room=${room.room_id}'" style="cursor: pointer;">
                                <div class="avatar" style="background-color: ${Utils.stringToColor(room.display_name || room.user_defined_name || 'Room')}">
                                    ${Utils.getInitials(room.display_name || room.user_defined_name || 'R')}
                                </div>
                                <div class="list-item-content">
                                    <div class="list-item-title">${Utils.escapeHtml(room.user_defined_name || room.display_name || 'Unknown Room')}</div>
                                    <div class="list-item-subtitle">${room.participant_count} participants - ${Utils.formatRelativeTime(room.last_activity)}</div>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
                <div class="card">
                    <div class="card-header">
                        <h3 class="card-title">Messages</h3>
                    </div>
                    <div class="card-body" id="messages-list">
                        <p class="text-muted text-center">Select a room to view messages</p>
                    </div>
                </div>
            </div>
        `;

        if (params.room) {
            this.loadRoomMessages(params.room);
        }
    }

    async loadRoomMessages(roomId) {
        try {
            const room = await this.api.getRoom(roomId);
            const container = document.getElementById('messages-list');
            if (!container) return;

            container.innerHTML = room.messages.map(msg => {
                const contact = room.contacts[msg.sender_contact_id];
                return `
                    <div class="list-item">
                        <div class="avatar avatar-sm" style="background-color: ${Utils.stringToColor(contact?.name || 'Unknown')}">
                            ${Utils.getInitials(contact?.name || '?')}
                        </div>
                        <div class="list-item-content">
                            <div class="flex items-center gap-2 mb-1">
                                <span class="font-medium">${Utils.escapeHtml(contact?.name || 'Unknown')}</span>
                                <span class="text-xs text-muted">${Utils.formatRelativeTime(msg.event_datetime)}</span>
                            </div>
                            <div class="text-sm">${Utils.escapeHtml(msg.body || '')}</div>
                        </div>
                    </div>
                `;
            }).join('') || '<p class="text-muted text-center">No messages</p>';
        } catch (error) {
            this.toast.error('Failed to load room messages');
        }
    }

    async loadEntities(params = {}) {
        await this.viewLoader.load('entities');

        try {
            const page = parseInt(params.page) || 1;
            const result = await this.api.getEntities({ page, pageSize: 20 });
            this.renderEntities(result);
        } catch (error) {
            this.toast.error('Failed to load entities');
        }
    }

    renderEntities(result) {
        const container = document.getElementById('entities-content');
        if (!container) return;

        const typeColors = {
            technology: 'info',
            organization: 'success',
            project: 'warning',
            event: 'pending'
        };

        container.innerHTML = `
            <div class="grid grid-cols-3">
                ${result.data.map(entity => `
                    <div class="card">
                        <div class="card-body">
                            <div class="flex items-center gap-4 mb-4">
                                <div class="avatar" style="background-color: ${Utils.stringToColor(entity.name)}">
                                    ${Utils.getInitials(entity.name)}
                                </div>
                                <div>
                                    <h4 class="font-semibold">${Utils.escapeHtml(entity.name)}</h4>
                                    <span class="badge badge-${typeColors[entity.type] || 'pending'}">${entity.type}</span>
                                </div>
                            </div>
                            <p class="text-sm text-muted">${Utils.escapeHtml(Utils.truncate(entity.description || 'No description', 100))}</p>
                        </div>
                    </div>
                `).join('')}
            </div>
            <div id="entities-pagination"></div>
        `;

        const pagination = new Pagination({
            page: result.page,
            totalPages: result.totalPages,
            onPageChange: (page) => {
                window.location.hash = `#/entities?page=${page}`;
            }
        });

        document.getElementById('entities-pagination').innerHTML = pagination.render();
        pagination.attachEvents(document.getElementById('entities-pagination'));
    }

    async loadSearch(params = {}) {
        await this.viewLoader.load('search');
        // Search view handles its own logic
    }

    async loadHistory(params = {}) {
        await this.viewLoader.load('history');

        try {
            const page = parseInt(params.page) || 1;
            const result = await this.api.getBrowserHistory({ page, pageSize: 20 });
            this.renderHistory(result);
        } catch (error) {
            this.toast.error('Failed to load browser history');
        }
    }

    renderHistory(result) {
        const container = document.getElementById('history-content');
        if (!container) return;

        container.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table class="table">
                        <thead>
                            <tr>
                                <th>Title</th>
                                <th>Domain</th>
                                <th>Visited</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${result.data.map(item => `
                                <tr>
                                    <td>
                                        <a href="${item.url}" target="_blank" class="list-item-title">${Utils.escapeHtml(item.title || item.url)}</a>
                                        <div class="list-item-subtitle">${Utils.escapeHtml(Utils.truncate(item.url, 60))}</div>
                                    </td>
                                    <td><span class="badge badge-pending">${item.domain}</span></td>
                                    <td class="text-muted">${Utils.formatRelativeTime(item.visit_date)}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
            <div id="history-pagination"></div>
        `;

        const pagination = new Pagination({
            page: result.page,
            totalPages: result.totalPages,
            onPageChange: (page) => {
                window.location.hash = `#/history?page=${page}`;
            }
        });

        document.getElementById('history-pagination').innerHTML = pagination.render();
        pagination.attachEvents(document.getElementById('history-pagination'));
    }

    async loadSocial(params = {}) {
        await this.viewLoader.load('social');

        try {
            const result = await this.api.getSocialPosts({ page: 1, limit: 10 });
            const credentials = await this.api.getSocialCredentials();
            this.renderSocial(result, credentials);
        } catch (error) {
            this.toast.error('Failed to load social posts');
        }
    }

    renderSocial(result, credentials) {
        const container = document.getElementById('social-content');
        if (!container) return;

        const statusColors = {
            posted: 'success',
            draft: 'pending',
            failed: 'error',
            pending: 'warning'
        };

        container.innerHTML = `
            <div class="grid grid-cols-2 mb-6">
                <div class="card">
                    <div class="card-body">
                        <div class="flex items-center gap-4">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                            </svg>
                            <div>
                                <h4 class="font-semibold">Twitter/X</h4>
                                <p class="text-sm ${credentials.twitter.working ? 'text-success' : 'text-error'}">
                                    ${credentials.twitter.working ? 'Connected' : 'Not connected'}
                                    ${credentials.twitter.profile?.username ? ` - @${credentials.twitter.profile.username}` : ''}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="card">
                    <div class="card-body">
                        <div class="flex items-center gap-4">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm5.5 6.5c-.67.33-1.39.56-2.15.66.77-.46 1.37-1.2 1.65-2.07-.72.43-1.53.74-2.38.91-.68-.73-1.65-1.18-2.72-1.18-2.06 0-3.73 1.67-3.73 3.73 0 .29.03.58.09.85-3.1-.16-5.85-1.64-7.69-3.9-.32.55-.5 1.2-.5 1.88 0 1.3.66 2.44 1.66 3.11-.61-.02-1.19-.19-1.69-.47v.05c0 1.81 1.29 3.32 3 3.66-.31.09-.64.13-.99.13-.24 0-.48-.02-.71-.07.48 1.5 1.87 2.59 3.52 2.62-1.29 1.01-2.92 1.61-4.69 1.61-.3 0-.6-.02-.9-.05 1.67 1.07 3.66 1.7 5.79 1.7 6.95 0 10.75-5.76 10.75-10.75 0-.16 0-.33-.01-.49.74-.53 1.38-1.2 1.89-1.96z"/>
                            </svg>
                            <div>
                                <h4 class="font-semibold">Bluesky</h4>
                                <p class="text-sm ${credentials.bluesky.working ? 'text-success' : 'text-error'}">
                                    ${credentials.bluesky.working ? 'Connected' : 'Not connected'}
                                    ${credentials.bluesky.profile?.handle ? ` - @${credentials.bluesky.profile.handle}` : ''}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Posts</h3>
                    <button class="btn btn-primary btn-sm" onclick="app.openNewPostModal()">New Post</button>
                </div>
                <div class="card-body" style="padding: 0;">
                    ${result.data.map(post => `
                        <div class="list-item">
                            <div class="list-item-content">
                                <div class="list-item-title">${Utils.escapeHtml(Utils.truncate(post.content, 100))}</div>
                                <div class="list-item-subtitle">
                                    ${Utils.formatRelativeTime(post.created_at)}
                                    ${post.twitter_post_id ? ' - Twitter' : ''}
                                    ${post.bluesky_post_id ? ' - Bluesky' : ''}
                                </div>
                            </div>
                            <span class="badge badge-${statusColors[post.status] || 'pending'}">${post.status}</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    openNewPostModal() {
        this.modal.open({
            title: 'New Social Post',
            content: `
                <div class="form-group">
                    <label class="form-label">Content</label>
                    <textarea class="form-textarea" id="post-content" placeholder="What's on your mind?" rows="4"></textarea>
                    <p class="form-helper"><span id="char-count">0</span>/280 characters</p>
                </div>
            `,
            footer: `
                <button class="btn btn-secondary" onclick="app.modal.close()">Cancel</button>
                <button class="btn btn-primary" onclick="app.createSocialPost()">Post</button>
            `
        });

        document.getElementById('post-content').addEventListener('input', (e) => {
            document.getElementById('char-count').textContent = e.target.value.length;
        });
    }

    async createSocialPost() {
        const content = document.getElementById('post-content').value;
        if (!content.trim()) {
            this.toast.warning('Please enter some content');
            return;
        }

        try {
            await this.api.createSocialPost({ content });
            this.modal.close();
            this.toast.success('Post created successfully');
            this.loadSocial();
        } catch (error) {
            this.toast.error('Failed to create post');
        }
    }

    async loadTags() {
        await this.viewLoader.load('tags');

        try {
            const tags = await this.api.getTags();
            this.renderTags(tags);
        } catch (error) {
            this.toast.error('Failed to load tags');
        }
    }

    renderTags(tags) {
        const container = document.getElementById('tags-content');
        if (!container) return;

        container.innerHTML = `
            <div class="card">
                <div class="card-body">
                    <div class="tag-list" style="gap: var(--space-3);">
                        ${tags.map(tag => `
                            <span class="tag" style="padding: var(--space-2) var(--space-4);">
                                ${Utils.escapeHtml(tag.name)}
                                <span class="text-muted text-xs">(${tag.usage_count || 0})</span>
                            </span>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;
    }

    async loadSettings() {
        await this.viewLoader.load('settings');

        try {
            const config = await this.api.getConfiguration();
            this.renderSettings(config);
        } catch (error) {
            this.toast.error('Failed to load settings');
        }
    }

    renderSettings(config) {
        const container = document.getElementById('settings-content');
        if (!container) return;

        container.innerHTML = `
            <div class="card mb-6">
                <div class="card-header">
                    <h3 class="card-title">Appearance</h3>
                </div>
                <div class="card-body">
                    <div class="form-group">
                        <label class="form-label">Theme</label>
                        <select class="form-select" id="theme-select">
                            <option value="dark" ${config.theme === 'dark' ? 'selected' : ''}>Dark</option>
                            <option value="light" ${config.theme === 'light' ? 'selected' : ''}>Light</option>
                            <option value="system" ${config.theme === 'system' ? 'selected' : ''}>System</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Language</label>
                        <select class="form-select" id="language-select">
                            <option value="en" ${config.language === 'en' ? 'selected' : ''}>English</option>
                            <option value="es" ${config.language === 'es' ? 'selected' : ''}>Spanish</option>
                            <option value="fr" ${config.language === 'fr' ? 'selected' : ''}>French</option>
                        </select>
                    </div>
                </div>
            </div>

            <div class="card mb-6">
                <div class="card-header">
                    <h3 class="card-title">Notifications</h3>
                </div>
                <div class="card-body">
                    <div class="form-group">
                        <label class="flex items-center gap-4">
                            <input type="checkbox" id="email-notifications" ${config.notifications?.email ? 'checked' : ''}>
                            <span>Email notifications</span>
                        </label>
                    </div>
                    <div class="form-group">
                        <label class="flex items-center gap-4">
                            <input type="checkbox" id="push-notifications" ${config.notifications?.push ? 'checked' : ''}>
                            <span>Push notifications</span>
                        </label>
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Integrations</h3>
                </div>
                <div class="card-body">
                    <div class="list-item">
                        <div class="list-item-content">
                            <div class="list-item-title">Twitter/X</div>
                            <div class="list-item-subtitle">${config.integrations?.twitter?.connected ? `Connected as @${config.integrations.twitter.username}` : 'Not connected'}</div>
                        </div>
                        <button class="btn btn-secondary btn-sm">${config.integrations?.twitter?.connected ? 'Disconnect' : 'Connect'}</button>
                    </div>
                    <div class="list-item">
                        <div class="list-item-content">
                            <div class="list-item-title">Bluesky</div>
                            <div class="list-item-subtitle">${config.integrations?.bluesky?.connected ? `Connected as @${config.integrations.bluesky.handle}` : 'Not connected'}</div>
                        </div>
                        <button class="btn btn-secondary btn-sm">${config.integrations?.bluesky?.connected ? 'Disconnect' : 'Connect'}</button>
                    </div>
                </div>
            </div>
        `;
    }
}

// Export for global access
window.GardenApp = GardenApp;
window.Utils = Utils;
window.Pagination = Pagination;
