/**
 * Garden PKM - Service Worker
 * Intercepts API requests and returns mock data
 */

const CACHE_NAME = 'garden-pkm-v1';
const DB_NAME = 'garden-pkm-db';
const DB_VERSION = 1;

// Import mock data
importScripts('mock-data.js');

// IndexedDB setup
let db = null;

function openDatabase() {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, DB_VERSION);

        request.onerror = () => reject(request.error);
        request.onsuccess = () => {
            db = request.result;
            resolve(db);
        };

        request.onupgradeneeded = (event) => {
            const database = event.target.result;

            // Create object stores
            const stores = [
                'bookmarks', 'notes', 'contacts', 'messages',
                'rooms', 'entities', 'browserHistory', 'socialPosts',
                'tags', 'categories', 'configuration'
            ];

            stores.forEach(storeName => {
                if (!database.objectStoreNames.contains(storeName)) {
                    const keyPath = storeName === 'browserHistory' ? 'id' :
                        storeName === 'bookmarks' ? 'bookmark_id' :
                            storeName === 'notes' ? 'id' :
                                storeName === 'contacts' ? 'contact_id' :
                                    storeName === 'messages' ? 'message_id' :
                                        storeName === 'rooms' ? 'room_id' :
                                            storeName === 'entities' ? 'entity_id' :
                                                storeName === 'socialPosts' ? 'post_id' :
                                                    storeName === 'configuration' ? 'key' : 'id';

                    database.createObjectStore(storeName, { keyPath });
                }
            });
        };
    });
}

// Initialize database with mock data
async function initializeData() {
    if (!db) await openDatabase();

    const tx = db.transaction(['bookmarks', 'notes', 'contacts', 'messages', 'rooms', 'entities', 'browserHistory', 'socialPosts', 'tags', 'categories'], 'readwrite');

    // Populate each store
    const stores = {
        bookmarks: MockData.bookmarks,
        notes: MockData.notes,
        contacts: MockData.contacts,
        messages: MockData.messages,
        rooms: MockData.rooms,
        entities: MockData.entities,
        browserHistory: MockData.browserHistory,
        socialPosts: MockData.socialPosts,
        tags: MockData.tags,
        categories: MockData.categories
    };

    for (const [storeName, data] of Object.entries(stores)) {
        const store = tx.objectStore(storeName);
        for (const item of data) {
            store.put(item);
        }
    }

    return new Promise((resolve, reject) => {
        tx.oncomplete = resolve;
        tx.onerror = () => reject(tx.error);
    });
}

// Simulate network delay
function delay(min = 100, max = 300) {
    const ms = Math.random() * (max - min) + min;
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Create JSON response
function jsonResponse(data, status = 200) {
    return new Response(JSON.stringify(data), {
        status,
        headers: {
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*'
        }
    });
}

// Parse URL parameters
function parseParams(url) {
    const urlObj = new URL(url);
    const params = {};
    urlObj.searchParams.forEach((value, key) => {
        params[key] = value;
    });
    return params;
}

// Get path segments
function getPathSegments(url) {
    const urlObj = new URL(url);
    return urlObj.pathname.split('/').filter(Boolean);
}

// Pagination helper
function paginate(data, page = 1, pageSize = 10) {
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return {
        data: data.slice(start, end),
        page: parseInt(page),
        pageSize: parseInt(pageSize),
        totalPages: Math.ceil(data.length / pageSize),
        totalItems: data.length
    };
}

// API Route Handlers
const handlers = {
    // Dashboard
    'GET /api/dashboard/stats': async () => {
        await delay();
        return jsonResponse(MockData.dashboardStats);
    },

    // Bookmarks
    'GET /api/bookmarks': async (url) => {
        await delay();
        const params = parseParams(url);
        const page = parseInt(params.page) || 1;
        const limit = parseInt(params.limit) || 10;
        const searchQuery = (params.searchQuery || params.search)?.toLowerCase();
        const category = params.category || params.categoryId;
        const status = params.status;
        const sort = params.sort || 'date_desc';

        let bookmarks = [...MockData.bookmarks];

        // Apply search filter
        if (searchQuery) {
            bookmarks = bookmarks.filter(b =>
                b.title?.toLowerCase().includes(searchQuery) ||
                b.url.toLowerCase().includes(searchQuery) ||
                b.summary?.toLowerCase().includes(searchQuery)
            );
        }

        // Apply category filter
        if (category) {
            bookmarks = bookmarks.filter(b => b.category_name === category);
        }

        // Apply status filter
        if (status) {
            bookmarks = bookmarks.filter(b => (b.status || 'pending') === status);
        }

        // Apply sorting
        switch (sort) {
            case 'date_asc':
                bookmarks.sort((a, b) => new Date(a.creation_date) - new Date(b.creation_date));
                break;
            case 'date_desc':
                bookmarks.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
                break;
            case 'title_asc':
                bookmarks.sort((a, b) => (a.title || '').localeCompare(b.title || ''));
                break;
            case 'title_desc':
                bookmarks.sort((a, b) => (b.title || '').localeCompare(a.title || ''));
                break;
        }

        return jsonResponse(paginate(bookmarks, page, limit));
    },

    'POST /api/bookmarks': async (url, request) => {
        await delay();
        const body = await request.json();
        const newBookmark = {
            bookmark_id: `b${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            url: body.url,
            title: body.title || `Bookmark from ${new URL(body.url).hostname}`,
            summary: body.summary || body.notes || '',
            category_name: body.category_name || body.category || null,
            creation_date: new Date().toISOString(),
            status: 'pending',
            tags: body.tags || []
        };
        MockData.bookmarks.unshift(newBookmark);
        return jsonResponse(newBookmark, 201);
    },

    'PUT /api/bookmarks/:id': async (url, request) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const body = await request.json();
        const bookmarkIndex = MockData.bookmarks.findIndex(b => b.bookmark_id === id);

        if (bookmarkIndex === -1) {
            return jsonResponse({ error: 'Bookmark not found' }, 404);
        }

        MockData.bookmarks[bookmarkIndex] = {
            ...MockData.bookmarks[bookmarkIndex],
            ...body,
            updated_at: new Date().toISOString()
        };

        return jsonResponse(MockData.bookmarks[bookmarkIndex]);
    },

    'DELETE /api/bookmarks/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const bookmarkIndex = MockData.bookmarks.findIndex(b => b.bookmark_id === id);

        if (bookmarkIndex === -1) {
            return jsonResponse({ error: 'Bookmark not found' }, 404);
        }

        MockData.bookmarks.splice(bookmarkIndex, 1);
        return jsonResponse({ message: 'Bookmark deleted successfully' });
    },

    'GET /api/bookmarks/random': async () => {
        await delay();
        const randomIndex = Math.floor(Math.random() * MockData.bookmarks.length);
        return jsonResponse(MockData.bookmarks[randomIndex]);
    },

    'GET /api/bookmarks/search': async (url) => {
        await delay();
        const params = parseParams(url);
        const query = params.query?.toLowerCase() || '';

        const results = MockData.bookmarks.filter(b =>
            b.title?.toLowerCase().includes(query) ||
            b.url.toLowerCase().includes(query) ||
            b.summary?.toLowerCase().includes(query)
        ).slice(0, 20);

        return jsonResponse(results);
    },

    'GET /api/bookmarks/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const bookmark = MockData.bookmarks.find(b => b.bookmark_id === id);

        if (!bookmark) {
            return jsonResponse({ error: 'Bookmark not found' }, 404);
        }

        return jsonResponse({
            ...bookmark,
            lynx: 'Processed content from lynx...',
            reader: 'Reader mode content...',
            questions: []
        });
    },

    'GET /api/bookmarks/missing/http': async () => {
        await delay();
        return jsonResponse(MockData.bookmarks.slice(0, 3));
    },

    'GET /api/bookmarks/missing/reader': async () => {
        await delay();
        return jsonResponse(MockData.bookmarks.slice(0, 5));
    },

    // Notes
    'GET /api/notes': async (url) => {
        await delay();
        const params = parseParams(url);
        const page = parseInt(params.page) || 1;
        const pageSize = parseInt(params.pageSize) || 12;
        const searchQuery = params.searchQuery?.toLowerCase();

        let notes = MockData.notes.map(n => ({
            id: n.id,
            title: n.title,
            tags: n.tags,
            created: n.created,
            modified: n.modified
        }));

        if (searchQuery) {
            notes = notes.filter(n =>
                n.title?.toLowerCase().includes(searchQuery) ||
                n.tags?.some(t => t.toLowerCase().includes(searchQuery))
            );
        }

        const result = paginate(notes, page, pageSize);
        return jsonResponse({
            notes: result.data,
            totalPages: result.totalPages
        });
    },

    'GET /api/notes/search': async (url) => {
        await delay();
        const params = parseParams(url);
        const query = params.q?.toLowerCase() || '';

        const results = MockData.notes.filter(n =>
            n.title?.toLowerCase().includes(query) ||
            n.contents?.toLowerCase().includes(query) ||
            n.tags?.some(t => t.toLowerCase().includes(query))
        ).map(n => ({
            id: n.id,
            title: n.title,
            tags: n.tags,
            created: n.created,
            modified: n.modified
        })).slice(0, 20);

        return jsonResponse(results);
    },

    'GET /api/notes/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const note = MockData.notes.find(n => n.id === id);

        if (!note) {
            return jsonResponse({ error: 'Note not found' }, 404);
        }

        return jsonResponse({
            ...note,
            processedContents: note.contents
        });
    },

    'POST /api/notes': async (url, request) => {
        await delay();
        const body = await request.json();
        const newNote = {
            id: `n${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            title: body.title,
            contents: body.contents || '',
            tags: body.tags || [],
            created: Date.now(),
            modified: Date.now()
        };
        MockData.notes.unshift(newNote);
        return jsonResponse(newNote, 201);
    },

    'PUT /api/notes/:id': async (url, request) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const body = await request.json();
        const noteIndex = MockData.notes.findIndex(n => n.id === id);

        if (noteIndex === -1) {
            return jsonResponse({ error: 'Note not found' }, 404);
        }

        MockData.notes[noteIndex] = {
            ...MockData.notes[noteIndex],
            ...body,
            modified: Date.now()
        };

        return jsonResponse(MockData.notes[noteIndex]);
    },

    'DELETE /api/notes/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const noteIndex = MockData.notes.findIndex(n => n.id === id);

        if (noteIndex === -1) {
            return jsonResponse({ error: 'Note not found' }, 404);
        }

        MockData.notes.splice(noteIndex, 1);
        return jsonResponse({ message: 'Note deleted successfully' });
    },

    // Contacts
    'GET /api/contacts': async (url) => {
        await delay();
        const params = parseParams(url);
        const page = parseInt(params.page) || 1;
        const pageSize = parseInt(params.pageSize) || 20;
        const search = (params.search || params.searchQuery)?.toLowerCase();

        let contacts = [...MockData.contacts];

        if (search) {
            contacts = contacts.filter(c =>
                c.name.toLowerCase().includes(search) ||
                c.email?.toLowerCase().includes(search)
            );
        }

        return jsonResponse(paginate(contacts, page, pageSize));
    },

    'GET /api/contacts/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const contact = MockData.contacts.find(c => c.contact_id === id);

        if (!contact) {
            return jsonResponse({ error: 'Contact not found' }, 404);
        }

        return jsonResponse({
            ...contact,
            known_names: [],
            alternativeNames: [],
            rooms: MockData.rooms.slice(0, 3).map(r => ({
                room_id: r.room_id,
                display_name: r.display_name,
                user_defined_name: r.user_defined_name
            })),
            sources: [{ id: '1', source_id: 'matrix', source_name: 'Matrix' }]
        });
    },

    'POST /api/contacts': async (url, request) => {
        await delay();
        const body = await request.json();
        const newContact = {
            contact_id: `ct${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            name: body.name,
            email: body.email,
            phone: body.phone,
            birthday: body.birthday,
            notes: body.notes,
            creation_date: new Date().toISOString(),
            last_update: new Date().toISOString(),
            last_week_messages: 0,
            groups_in_common: 0,
            importance: 5,
            closeness: 5,
            fondness: 5,
            tags: []
        };
        MockData.contacts.unshift(newContact);
        return jsonResponse(newContact, 201);
    },

    'PUT /api/contacts/:id': async (url, request) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const body = await request.json();
        const contactIndex = MockData.contacts.findIndex(c => c.contact_id === id);

        if (contactIndex === -1) {
            return jsonResponse({ error: 'Contact not found' }, 404);
        }

        MockData.contacts[contactIndex] = {
            ...MockData.contacts[contactIndex],
            ...body,
            last_update: new Date().toISOString()
        };

        return jsonResponse({ message: 'Contact updated successfully' });
    },

    'DELETE /api/contacts/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const contactIndex = MockData.contacts.findIndex(c => c.contact_id === id);

        if (contactIndex === -1) {
            return jsonResponse({ error: 'Contact not found' }, 404);
        }

        MockData.contacts.splice(contactIndex, 1);
        return jsonResponse({ message: 'Contact deleted successfully' });
    },

    'PUT /api/contacts/:id/evaluation': async (url, request) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const body = await request.json();
        const contactIndex = MockData.contacts.findIndex(c => c.contact_id === id);

        if (contactIndex === -1) {
            return jsonResponse({ error: 'Contact not found' }, 404);
        }

        if (body.importance !== undefined) MockData.contacts[contactIndex].importance = body.importance;
        if (body.closeness !== undefined) MockData.contacts[contactIndex].closeness = body.closeness;
        if (body.fondness !== undefined) MockData.contacts[contactIndex].fondness = body.fondness;

        return jsonResponse({ message: 'Contact evaluation updated successfully' });
    },

    'GET /api/contact-tags': async () => {
        await delay();
        const tags = new Map();
        MockData.contacts.forEach(c => {
            c.tags?.forEach(t => {
                if (!tags.has(t.tag_id)) {
                    tags.set(t.tag_id, t);
                }
            });
        });
        return jsonResponse(Array.from(tags.values()));
    },

    // Messages
    'GET /api/messages': async (url) => {
        await delay();
        const params = parseParams(url);
        const roomId = params.roomId;
        const page = parseInt(params.page) || 1;
        const pageSize = parseInt(params.pageSize) || 20;

        let messages = [...MockData.messages];

        if (roomId) {
            messages = messages.filter(m => m.room_id === roomId);
        }

        return jsonResponse(paginate(messages, page, pageSize));
    },

    'GET /api/messages/search': async (url) => {
        await delay();
        const params = parseParams(url);
        const query = params.query?.toLowerCase() || '';

        const results = MockData.messages.filter(m =>
            m.body?.toLowerCase().includes(query)
        ).slice(0, 50);

        return jsonResponse({
            messages: results,
            contacts: MockData.contacts.reduce((acc, c) => {
                acc[c.contact_id] = c;
                return acc;
            }, {})
        });
    },

    // Rooms
    'GET /api/rooms': async () => {
        await delay();
        return jsonResponse(MockData.rooms);
    },

    'GET /api/rooms/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const room = MockData.rooms.find(r => r.room_id === id);

        if (!room) {
            return jsonResponse({ error: 'Room not found' }, 404);
        }

        const roomMessages = MockData.messages.filter(m => m.room_id === id);
        const participants = [...new Set(roomMessages.map(m => m.sender_contact_id))];

        return jsonResponse({
            room,
            messages: roomMessages,
            participants: participants.map(pid => {
                const contact = MockData.contacts.find(c => c.contact_id === pid);
                return contact ? { contact_id: pid, name: contact.name, email: contact.email } : null;
            }).filter(Boolean),
            contacts: MockData.contacts.reduce((acc, c) => {
                acc[c.contact_id] = c;
                return acc;
            }, {})
        });
    },

    // Entities
    'GET /api/entities': async (url) => {
        await delay();
        const params = parseParams(url);
        const page = parseInt(params.page) || 1;
        const pageSize = parseInt(params.pageSize) || 20;
        const type = params.type;

        let entities = [...MockData.entities];

        if (type) {
            entities = entities.filter(e => e.type === type);
        }

        return jsonResponse(paginate(entities, page, pageSize));
    },

    'GET /api/entities/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const entity = MockData.entities.find(e => e.entity_id === id);

        if (!entity) {
            return jsonResponse({ error: 'Entity not found' }, 404);
        }

        return jsonResponse({
            ...entity,
            relationships: []
        });
    },

    'POST /api/entities': async (url, request) => {
        await delay();
        const body = await request.json();
        const newEntity = {
            entity_id: `en${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            name: body.name,
            type: body.type,
            description: body.description,
            properties: body.properties || {},
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
        MockData.entities.unshift(newEntity);
        return jsonResponse(newEntity, 201);
    },

    // Browser History
    'GET /api/browser-history': async (url) => {
        await delay();
        const params = parseParams(url);
        const page = parseInt(params.page) || 1;
        const pageSize = parseInt(params.pageSize) || 20;
        const searchQuery = params.searchQuery?.toLowerCase();
        const domain = params.domain;

        let history = [...MockData.browserHistory];

        if (searchQuery) {
            history = history.filter(h =>
                h.title?.toLowerCase().includes(searchQuery) ||
                h.url.toLowerCase().includes(searchQuery)
            );
        }

        if (domain) {
            history = history.filter(h => h.domain === domain);
        }

        return jsonResponse(paginate(history, page, pageSize));
    },

    'GET /api/browser-history/domains': async () => {
        await delay();
        const domainCounts = {};
        MockData.browserHistory.forEach(h => {
            domainCounts[h.domain] = (domainCounts[h.domain] || 0) + 1;
        });

        const domains = Object.entries(domainCounts)
            .map(([domain, count]) => ({ domain, visit_count: count }))
            .sort((a, b) => b.visit_count - a.visit_count);

        return jsonResponse(domains);
    },

    // Social Posts
    'GET /api/social-posts': async (url) => {
        await delay();
        const params = parseParams(url);
        const page = parseInt(params.page) || 1;
        const limit = parseInt(params.limit) || 10;
        const status = params.status;

        let posts = [...MockData.socialPosts];

        if (status) {
            posts = posts.filter(p => p.status === status);
        }

        return jsonResponse(paginate(posts, page, limit));
    },

    'GET /api/social-posts/:id': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const post = MockData.socialPosts.find(p => p.post_id === id);

        if (!post) {
            return jsonResponse({ error: 'Post not found' }, 404);
        }

        return jsonResponse(post);
    },

    'POST /api/social-posts': async (url, request) => {
        await delay();
        const body = await request.json();
        const newPost = {
            post_id: `sp${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            content: body.content,
            twitter_post_id: null,
            bluesky_post_id: null,
            created_at: new Date().toISOString(),
            status: 'draft'
        };
        MockData.socialPosts.unshift(newPost);
        return jsonResponse(newPost, 201);
    },

    'POST /api/social-posts/:id/publish': async (url) => {
        await delay();
        const segments = getPathSegments(url);
        const id = segments[2];
        const postIndex = MockData.socialPosts.findIndex(p => p.post_id === id);

        if (postIndex === -1) {
            return jsonResponse({ error: 'Post not found' }, 404);
        }

        MockData.socialPosts[postIndex].status = 'posted';
        MockData.socialPosts[postIndex].twitter_post_id = `17${Date.now()}`;
        MockData.socialPosts[postIndex].bluesky_post_id = `3k${Math.random().toString(36).substr(2, 10)}`;

        return jsonResponse(MockData.socialPosts[postIndex]);
    },

    'GET /api/social/credentials': async () => {
        await delay();
        return jsonResponse({
            twitter: {
                configured: true,
                working: true,
                profile: { username: 'gardenuser', display_name: 'Garden User' }
            },
            bluesky: {
                configured: true,
                working: true,
                profile: { handle: 'garden.bsky.social', did: 'did:plc:example123' }
            }
        });
    },

    // Tags
    'GET /api/tags': async () => {
        await delay();
        return jsonResponse(MockData.tags);
    },

    // Categories
    'GET /api/categories': async () => {
        await delay();
        return jsonResponse(MockData.categories);
    },

    // Search (Global)
    'GET /api/search': async (url) => {
        await delay();
        const params = parseParams(url);
        const query = params.q?.toLowerCase() || '';

        const bookmarks = MockData.bookmarks.filter(b =>
            b.title?.toLowerCase().includes(query) ||
            b.url.toLowerCase().includes(query)
        ).slice(0, 5).map(b => ({ type: 'bookmark', ...b }));

        const notes = MockData.notes.filter(n =>
            n.title?.toLowerCase().includes(query) ||
            n.contents?.toLowerCase().includes(query)
        ).slice(0, 5).map(n => ({ type: 'note', ...n }));

        const contacts = MockData.contacts.filter(c =>
            c.name.toLowerCase().includes(query) ||
            c.email?.toLowerCase().includes(query)
        ).slice(0, 5).map(c => ({ type: 'contact', ...c }));

        const entities = MockData.entities.filter(e =>
            e.name.toLowerCase().includes(query) ||
            e.description?.toLowerCase().includes(query)
        ).slice(0, 5).map(e => ({ type: 'entity', ...e }));

        return jsonResponse({
            bookmarks,
            notes,
            contacts,
            entities,
            total: bookmarks.length + notes.length + contacts.length + entities.length
        });
    },

    // Configuration
    'GET /api/configuration': async () => {
        await delay();
        return jsonResponse(MockData.configuration);
    },

    'PUT /api/configuration': async (url, request) => {
        await delay();
        const body = await request.json();
        // Deep merge configuration
        function deepMerge(target, source) {
            for (const key of Object.keys(source)) {
                if (source[key] instanceof Object && key in target) {
                    Object.assign(source[key], deepMerge(target[key], source[key]));
                }
            }
            return { ...target, ...source };
        }
        MockData.configuration = deepMerge(MockData.configuration, body);
        return jsonResponse(MockData.configuration);
    },

    'PATCH /api/configuration/:section': async (url, request) => {
        await delay();
        const segments = getPathSegments(url);
        const section = segments[2];
        const body = await request.json();

        if (MockData.configuration[section]) {
            MockData.configuration[section] = { ...MockData.configuration[section], ...body };
            return jsonResponse(MockData.configuration[section]);
        }
        return jsonResponse({ error: 'Configuration section not found' }, 404);
    },

    // Ollama connection test
    'POST /api/integrations/ollama/test': async (url, request) => {
        await delay(500, 1500);
        const body = await request.json();
        // Simulate connection test
        const success = body.url && body.url.includes('localhost');
        if (success) {
            MockData.configuration.integrations.ollama.status = 'connected';
            MockData.configuration.integrations.ollama.lastConnected = new Date().toISOString();
            return jsonResponse({
                success: true,
                message: 'Connected to Ollama successfully',
                models: ['llama3.2', 'llama3.1', 'mistral', 'codellama', 'nomic-embed-text']
            });
        }
        return jsonResponse({ success: false, error: 'Failed to connect to Ollama' }, 400);
    },

    // Logseq sync
    'POST /api/integrations/logseq/sync': async () => {
        await delay(1000, 2000);
        MockData.configuration.integrations.logseq.lastSync = new Date().toISOString();
        MockData.configuration.integrations.logseq.status = 'synced';
        return jsonResponse({
            success: true,
            message: 'Logseq sync completed',
            itemsSynced: Math.floor(Math.random() * 50) + 10
        });
    },

    // Social connect/disconnect
    'POST /api/integrations/twitter/connect': async () => {
        await delay(500, 1000);
        MockData.configuration.integrations.twitter.connected = true;
        MockData.configuration.integrations.twitter.connectedAt = new Date().toISOString();
        return jsonResponse({ success: true, message: 'Twitter connected' });
    },

    'POST /api/integrations/twitter/disconnect': async () => {
        await delay();
        MockData.configuration.integrations.twitter.connected = false;
        MockData.configuration.integrations.twitter.username = null;
        return jsonResponse({ success: true, message: 'Twitter disconnected' });
    },

    'POST /api/integrations/bluesky/connect': async () => {
        await delay(500, 1000);
        MockData.configuration.integrations.bluesky.connected = true;
        MockData.configuration.integrations.bluesky.connectedAt = new Date().toISOString();
        return jsonResponse({ success: true, message: 'Bluesky connected' });
    },

    'POST /api/integrations/bluesky/disconnect': async () => {
        await delay();
        MockData.configuration.integrations.bluesky.connected = false;
        MockData.configuration.integrations.bluesky.handle = null;
        return jsonResponse({ success: true, message: 'Bluesky disconnected' });
    },

    // Data operations
    'POST /api/data/export': async () => {
        await delay(500, 1500);
        return jsonResponse({
            success: true,
            data: {
                bookmarks: MockData.bookmarks,
                notes: MockData.notes,
                contacts: MockData.contacts,
                entities: MockData.entities,
                configuration: MockData.configuration
            },
            exportedAt: new Date().toISOString()
        });
    },

    'POST /api/data/import': async (url, request) => {
        await delay(1000, 2000);
        // Simulate import
        return jsonResponse({
            success: true,
            message: 'Data imported successfully',
            imported: { bookmarks: 10, notes: 5, contacts: 8 }
        });
    },

    'DELETE /api/data/cache': async () => {
        await delay();
        MockData.configuration.storage.breakdown.cache = 0;
        return jsonResponse({ success: true, message: 'Cache cleared', freedSpace: 35000000 });
    },

    'DELETE /api/data/all': async () => {
        await delay(1000, 2000);
        // Don't actually delete in mock, just simulate
        return jsonResponse({ success: true, message: 'All data deleted' });
    },

    'DELETE /api/data/history': async () => {
        await delay();
        MockData.browserHistory = [];
        return jsonResponse({ success: true, message: 'History cleared' });
    },

    // Reset settings
    'POST /api/configuration/reset': async (url, request) => {
        await delay();
        const body = await request.json();
        const section = body.section;

        // Reset specific section to defaults
        const defaults = {
            general: {
                displayName: "Garden User",
                email: "user@example.com",
                language: "en",
                timezone: "America/New_York",
                defaultView: { dashboard: "overview", bookmarks: "grid", notes: "grid", contacts: "list", messages: "threaded" }
            },
            appearance: {
                theme: "dark",
                accentColor: "#0070f3",
                sidebarPosition: "left",
                compactMode: false,
                fontSize: 16,
                codeFont: "SF Mono",
                showPreview: true
            },
            privacy: {
                dataRetention: { enabled: true, historyDays: 90, messageDays: 365, deletedItemsDays: 30 },
                autoDelete: { enabled: false, interval: "monthly" },
                anonymizeExports: true,
                shareAnalytics: false
            },
            advanced: {
                apiEndpoint: "http://localhost:8000",
                debugMode: false,
                consoleLogging: false,
                experimentalFeatures: false,
                rawConfig: {}
            }
        };

        if (section && defaults[section]) {
            MockData.configuration[section] = defaults[section];
            return jsonResponse({ success: true, message: `${section} settings reset`, data: defaults[section] });
        } else if (section === 'all') {
            Object.keys(defaults).forEach(key => {
                MockData.configuration[key] = defaults[key];
            });
            return jsonResponse({ success: true, message: 'All settings reset' });
        }

        return jsonResponse({ error: 'Invalid section' }, 400);
    },

    // Credentials
    'DELETE /api/credentials': async () => {
        await delay();
        MockData.configuration.integrations.twitter.connected = false;
        MockData.configuration.integrations.bluesky.connected = false;
        MockData.configuration.integrations.ollama.status = 'disconnected';
        return jsonResponse({ success: true, message: 'All credentials cleared' });
    }
};

// Match route pattern to handler
function matchRoute(method, pathname) {
    const routeKey = `${method} ${pathname}`;

    // Try exact match first
    if (handlers[routeKey]) {
        return handlers[routeKey];
    }

    // Try pattern matching for routes with :id
    for (const [pattern, handler] of Object.entries(handlers)) {
        const [patternMethod, patternPath] = pattern.split(' ');
        if (method !== patternMethod) continue;

        const patternParts = patternPath.split('/');
        const pathParts = pathname.split('/');

        if (patternParts.length !== pathParts.length) continue;

        let matches = true;
        for (let i = 0; i < patternParts.length; i++) {
            if (patternParts[i].startsWith(':')) continue;
            if (patternParts[i] !== pathParts[i]) {
                matches = false;
                break;
            }
        }

        if (matches) return handler;
    }

    return null;
}

// Service Worker Event Handlers
self.addEventListener('install', (event) => {
    console.log('Service Worker installing...');
    event.waitUntil(
        Promise.all([
            openDatabase(),
            self.skipWaiting()
        ])
    );
});

self.addEventListener('activate', (event) => {
    console.log('Service Worker activating...');
    event.waitUntil(
        Promise.all([
            initializeData(),
            self.clients.claim()
        ])
    );
});

self.addEventListener('fetch', (event) => {
    const url = new URL(event.request.url);

    // Only intercept /api/* requests
    if (!url.pathname.startsWith('/api/')) {
        return;
    }

    const method = event.request.method;
    const pathname = url.pathname;

    // Handle CORS preflight
    if (method === 'OPTIONS') {
        event.respondWith(new Response(null, {
            status: 204,
            headers: {
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
                'Access-Control-Allow-Headers': 'Content-Type, Authorization'
            }
        }));
        return;
    }

    const handler = matchRoute(method, pathname);

    if (handler) {
        event.respondWith(
            handler(event.request.url, event.request)
                .catch(error => {
                    console.error('Handler error:', error);
                    return jsonResponse({ error: 'Internal server error' }, 500);
                })
        );
    } else {
        // Return 404 for unmatched API routes
        event.respondWith(
            delay().then(() => jsonResponse({ error: 'Not found', path: pathname }, 404))
        );
    }
});

// Handle messages from clients
self.addEventListener('message', (event) => {
    if (event.data.type === 'SKIP_WAITING') {
        self.skipWaiting();
    }

    if (event.data.type === 'RESET_DATA') {
        // Reset mock data
        initializeData().then(() => {
            event.ports[0].postMessage({ success: true });
        });
    }
});
