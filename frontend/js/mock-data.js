/**
 * Garden PKM - Mock Data
 * Comprehensive mock data for all application entities
 */

const MockData = {
    // ============================================
    // Bookmarks
    // ============================================
    bookmarks: [
        {
            bookmark_id: "b1a2c3d4-e5f6-7890-abcd-ef1234567890",
            url: "https://react.dev/learn",
            creation_date: "2024-01-15T10:30:00Z",
            title: "Quick Start - React",
            summary: "The official React documentation providing a comprehensive guide to getting started with React, including component creation, state management, and hooks.",
            category_name: "Development"
        },
        {
            bookmark_id: "b2a3c4d5-e6f7-8901-bcde-f12345678901",
            url: "https://www.typescriptlang.org/docs/handbook/intro.html",
            creation_date: "2024-01-14T14:20:00Z",
            title: "The TypeScript Handbook",
            summary: "A comprehensive guide to TypeScript, covering types, interfaces, classes, generics, and advanced patterns for building type-safe JavaScript applications.",
            category_name: "Development"
        },
        {
            bookmark_id: "b3a4c5d6-e7f8-9012-cdef-123456789012",
            url: "https://tailwindcss.com/docs/installation",
            creation_date: "2024-01-13T09:15:00Z",
            title: "Installation - Tailwind CSS",
            summary: "Installation guide for Tailwind CSS, a utility-first CSS framework for rapidly building custom user interfaces without leaving your HTML.",
            category_name: "Design"
        },
        {
            bookmark_id: "b4a5c6d7-e8f9-0123-def1-234567890123",
            url: "https://go.dev/doc/effective_go",
            creation_date: "2024-01-12T16:45:00Z",
            title: "Effective Go - The Go Programming Language",
            summary: "Tips for writing clear, idiomatic Go code. Covers formatting, commentary, names, control structures, functions, data, and more.",
            category_name: "Development"
        },
        {
            bookmark_id: "b5a6c7d8-e9f0-1234-ef12-345678901234",
            url: "https://www.postgresql.org/docs/current/tutorial.html",
            creation_date: "2024-01-11T11:00:00Z",
            title: "PostgreSQL Tutorial",
            summary: "Introduction to PostgreSQL covering SQL basics, advanced features, and database administration for building robust data-driven applications.",
            category_name: "Database"
        },
        {
            bookmark_id: "b6a7c8d9-e0f1-2345-f123-456789012345",
            url: "https://kubernetes.io/docs/tutorials/kubernetes-basics/",
            creation_date: "2024-01-10T08:30:00Z",
            title: "Learn Kubernetes Basics",
            summary: "Interactive tutorial covering Kubernetes fundamentals including deploying, scaling, and updating containerized applications.",
            category_name: "DevOps"
        },
        {
            bookmark_id: "b7a8c9d0-e1f2-3456-0123-567890123456",
            url: "https://stripe.com/docs/api",
            creation_date: "2024-01-09T13:20:00Z",
            title: "Stripe API Reference",
            summary: "Complete API reference for Stripe payments, including authentication, requests, responses, and integration patterns.",
            category_name: "API"
        },
        {
            bookmark_id: "b8a9c0d1-e2f3-4567-1234-678901234567",
            url: "https://www.figma.com/best-practices/guide-to-design-systems/",
            creation_date: "2024-01-08T15:50:00Z",
            title: "Guide to Design Systems - Figma",
            summary: "Best practices for creating and maintaining design systems that scale across products and teams.",
            category_name: "Design"
        },
        {
            bookmark_id: "b9a0c1d2-e3f4-5678-2345-789012345678",
            url: "https://martinfowler.com/articles/microservices.html",
            creation_date: "2024-01-07T10:10:00Z",
            title: "Microservices - Martin Fowler",
            summary: "In-depth exploration of microservices architecture, including characteristics, benefits, and implementation considerations.",
            category_name: "Architecture"
        },
        {
            bookmark_id: "b0a1c2d3-e4f5-6789-3456-890123456789",
            url: "https://testing-library.com/docs/react-testing-library/intro/",
            creation_date: "2024-01-06T17:30:00Z",
            title: "React Testing Library - Introduction",
            summary: "Getting started with React Testing Library for writing maintainable tests that focus on user behavior.",
            category_name: "Testing"
        },
        {
            bookmark_id: "c1b2a3d4-f5e6-7890-4567-901234567890",
            url: "https://www.anthropic.com/news/claude-3-family",
            creation_date: "2024-01-05T12:00:00Z",
            title: "Claude 3 Family - Anthropic",
            summary: "Announcement of the Claude 3 model family with improved capabilities in analysis, forecasting, and nuanced content creation.",
            category_name: "AI"
        },
        {
            bookmark_id: "c2b3a4d5-f6e7-8901-5678-012345678901",
            url: "https://nextjs.org/docs/app/building-your-application/routing",
            creation_date: "2024-01-04T09:45:00Z",
            title: "Routing - Next.js Documentation",
            summary: "Comprehensive guide to Next.js App Router, including dynamic routes, layouts, loading states, and error handling.",
            category_name: "Development"
        },
        {
            bookmark_id: "c3b4a5d6-f7e8-9012-6789-123456789012",
            url: "https://blog.pragmaticengineer.com/software-architecture-is-overrated/",
            creation_date: "2024-01-03T14:15:00Z",
            title: "Software Architecture is Overrated - The Pragmatic Engineer",
            summary: "A balanced perspective on when architecture matters and when simpler solutions are more appropriate.",
            category_name: "Architecture"
        },
        {
            bookmark_id: "c4b5a6d7-f8e9-0123-7890-234567890123",
            url: "https://docs.github.com/en/actions",
            creation_date: "2024-01-02T11:30:00Z",
            title: "GitHub Actions Documentation",
            summary: "Complete guide to GitHub Actions for automating builds, tests, and deployments directly from your GitHub repository.",
            category_name: "DevOps"
        },
        {
            bookmark_id: "c5b6a7d8-f9e0-1234-8901-345678901234",
            url: "https://www.nngroup.com/articles/ten-usability-heuristics/",
            creation_date: "2024-01-01T08:00:00Z",
            title: "10 Usability Heuristics for User Interface Design",
            summary: "Jakob Nielsen's 10 general principles for interaction design, also known as heuristics for usability evaluation.",
            category_name: "UX"
        },
        {
            bookmark_id: "c6b7a8d9-f0e1-2345-9012-456789012345",
            url: "https://vercel.com/docs/edge-network/overview",
            creation_date: "2023-12-31T16:20:00Z",
            title: "Edge Network Overview - Vercel",
            summary: "Understanding Vercel's Edge Network for global deployment and optimal performance.",
            category_name: "Infrastructure"
        },
        {
            bookmark_id: "c7b8a9d0-f1e2-3456-0123-567890123456",
            url: "https://supabase.com/docs/guides/getting-started",
            creation_date: "2023-12-30T13:40:00Z",
            title: "Getting Started with Supabase",
            summary: "Quick start guide for Supabase, the open-source Firebase alternative with Postgres database and authentication.",
            category_name: "Database"
        },
        {
            bookmark_id: "c8b9a0d1-f2e3-4567-1234-678901234567",
            url: "https://redis.io/docs/getting-started/",
            creation_date: "2023-12-29T10:15:00Z",
            title: "Getting Started with Redis",
            summary: "Introduction to Redis, the in-memory data structure store used as database, cache, message broker, and queue.",
            category_name: "Database"
        },
        {
            bookmark_id: "c9b0a1d2-f3e4-5678-2345-789012345678",
            url: "https://www.smashingmagazine.com/2024/01/guide-dark-mode-design/",
            creation_date: "2023-12-28T15:55:00Z",
            title: "Complete Guide to Dark Mode Design - Smashing Magazine",
            summary: "Best practices for implementing dark mode, including color choices, contrast, and accessibility considerations.",
            category_name: "Design"
        },
        {
            bookmark_id: "c0b1a2d3-f4e5-6789-3456-890123456789",
            url: "https://orm.drizzle.team/docs/overview",
            creation_date: "2023-12-27T11:25:00Z",
            title: "Drizzle ORM - Overview",
            summary: "TypeScript ORM with type safety, auto-completion, and a query builder that feels like SQL.",
            category_name: "Development"
        },
        {
            bookmark_id: "d1c2b3a4-0567-8901-4567-901234567890",
            url: "https://htmx.org/docs/",
            creation_date: "2023-12-26T09:00:00Z",
            title: "htmx - Documentation",
            summary: "High power tools for HTML - access modern browser features directly from HTML with minimal JavaScript.",
            category_name: "Development"
        },
        {
            bookmark_id: "d2c3b4a5-1678-9012-5678-012345678901",
            url: "https://opentelemetry.io/docs/",
            creation_date: "2023-12-25T14:30:00Z",
            title: "OpenTelemetry Documentation",
            summary: "Observability framework for cloud-native software, providing APIs, libraries, and tools for telemetry data.",
            category_name: "DevOps"
        }
    ],

    // ============================================
    // Notes
    // ============================================
    notes: [
        {
            id: "n1a2b3c4-d5e6-7890-abcd-ef1234567890",
            title: "System Design Interview Preparation",
            contents: "# System Design Interview Preparation\n\n## Key Topics\n- Load balancing strategies\n- Database sharding and replication\n- Caching layers (Redis, Memcached)\n- Message queues (Kafka, RabbitMQ)\n- CAP theorem and trade-offs\n\n## Practice Problems\n1. Design a URL shortener\n2. Design a rate limiter\n3. Design a notification system\n4. Design a distributed cache\n\n## Resources\n- Designing Data-Intensive Applications\n- System Design Primer on GitHub",
            tags: ["interview", "system-design", "career"],
            created: 1705312200000,
            modified: 1705398600000
        },
        {
            id: "n2a3b4c5-d6e7-8901-bcde-f12345678901",
            title: "React Performance Optimization",
            contents: "# React Performance Optimization\n\n## Memoization\n- `useMemo` for expensive calculations\n- `useCallback` for function references\n- `React.memo` for component re-renders\n\n## Code Splitting\n- Dynamic imports with `React.lazy`\n- Route-based splitting\n- Component-based splitting\n\n## Virtual Lists\n- react-window for long lists\n- react-virtualized for complex grids\n\n## Profiling\n- React DevTools Profiler\n- Performance API\n- Lighthouse audits",
            tags: ["react", "performance", "frontend"],
            created: 1705225800000,
            modified: 1705312200000
        },
        {
            id: "n3a4b5c6-d7e8-9012-cdef-123456789012",
            title: "Weekly Goals - January Week 3",
            contents: "# Weekly Goals\n\n## Work\n- [ ] Complete API redesign proposal\n- [ ] Review team PRs\n- [x] Set up monitoring dashboard\n- [ ] Document authentication flow\n\n## Personal\n- [x] Finish reading Atomic Habits\n- [ ] 3 gym sessions\n- [ ] Meal prep for the week\n\n## Learning\n- [ ] Complete Rust chapter 10\n- [x] Watch database indexing video\n- [ ] Practice LeetCode (3 problems)",
            tags: ["weekly-goals", "planning"],
            created: 1705139400000,
            modified: 1705398600000
        },
        {
            id: "n4a5b6c7-d8e9-0123-def1-234567890123",
            title: "PostgreSQL Optimization Notes",
            contents: "# PostgreSQL Optimization\n\n## Indexes\n- B-tree (default, range queries)\n- Hash (equality only)\n- GiST (geometric, text search)\n- GIN (arrays, JSONB)\n\n## Query Analysis\n```sql\nEXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';\n```\n\n## Key Metrics\n- `shared_buffers`: 25% of RAM\n- `effective_cache_size`: 75% of RAM\n- `work_mem`: 64MB for complex queries\n\n## Vacuum Settings\n- `autovacuum = on`\n- Monitor pg_stat_user_tables for dead tuples",
            tags: ["database", "postgresql", "optimization"],
            created: 1705053000000,
            modified: 1705139400000
        },
        {
            id: "n5a6b7c8-d9e0-1234-ef12-345678901234",
            title: "Meeting Notes: Product Roadmap Q1",
            contents: "# Product Roadmap Q1 2024\n\n**Date**: January 8, 2024\n**Attendees**: Product, Engineering, Design\n\n## Key Initiatives\n1. **User Dashboard Redesign**\n   - Timeline: 4 weeks\n   - Owner: Design team\n   - Dependencies: New component library\n\n2. **API v2 Migration**\n   - Timeline: 6 weeks\n   - Backward compatibility required\n   - Breaking changes documented\n\n3. **Mobile App Beta**\n   - React Native implementation\n   - Feature parity with web (core features)\n   - TestFlight/Play Store beta by Feb 15\n\n## Action Items\n- [ ] @john Create technical spec for API v2\n- [ ] @sarah Finalize mobile designs by EOW\n- [ ] @mike Set up feature flags for gradual rollout",
            tags: ["meeting-notes", "roadmap", "product"],
            created: 1704966600000,
            modified: 1704966600000
        },
        {
            id: "n6a7b8c9-d0e1-2345-f123-456789012345",
            title: "Git Workflow Best Practices",
            contents: "# Git Workflow Best Practices\n\n## Branch Naming\n- `feature/JIRA-123-user-auth`\n- `bugfix/JIRA-456-fix-login`\n- `hotfix/critical-security-patch`\n\n## Commit Messages\n```\nfeat(auth): add OAuth2 support for Google login\n\n- Implement Google OAuth2 flow\n- Add token refresh mechanism\n- Update user model with provider field\n\nCloses #123\n```\n\n## Code Review Guidelines\n1. Check tests are included\n2. Verify documentation updated\n3. Look for performance implications\n4. Ensure backward compatibility\n\n## Merge Strategy\n- Squash for feature branches\n- Merge commit for release branches\n- Rebase for local cleanup only",
            tags: ["git", "workflow", "best-practices"],
            created: 1704880200000,
            modified: 1705139400000
        },
        {
            id: "n7a8b9c0-d1e2-3456-0123-567890123456",
            title: "Docker Compose Development Setup",
            contents: "# Docker Compose Development Setup\n\n```yaml\nversion: '3.8'\nservices:\n  app:\n    build: .\n    ports:\n      - '3000:3000'\n    volumes:\n      - .:/app\n    depends_on:\n      - db\n      - redis\n\n  db:\n    image: postgres:15\n    environment:\n      POSTGRES_PASSWORD: dev_password\n    volumes:\n      - pgdata:/var/lib/postgresql/data\n\n  redis:\n    image: redis:7-alpine\n    ports:\n      - '6379:6379'\n\nvolumes:\n  pgdata:\n```\n\n## Commands\n- Start: `docker compose up -d`\n- Logs: `docker compose logs -f app`\n- Reset: `docker compose down -v`",
            tags: ["docker", "development", "devops"],
            created: 1704793800000,
            modified: 1704880200000
        },
        {
            id: "n8a9b0c1-d2e3-4567-1234-678901234567",
            title: "API Design Principles",
            contents: "# API Design Principles\n\n## RESTful Conventions\n- Use nouns for resources (`/users`, `/orders`)\n- HTTP methods for actions (GET, POST, PUT, DELETE)\n- Consistent response structure\n- Proper status codes\n\n## Pagination\n```json\n{\n  \"data\": [...],\n  \"pagination\": {\n    \"page\": 1,\n    \"pageSize\": 20,\n    \"totalPages\": 5,\n    \"totalItems\": 100\n  }\n}\n```\n\n## Error Handling\n```json\n{\n  \"error\": {\n    \"code\": \"VALIDATION_ERROR\",\n    \"message\": \"Email is required\",\n    \"details\": [...]\n  }\n}\n```\n\n## Versioning\n- URL versioning: `/api/v1/users`\n- Header versioning: `Accept: application/vnd.api+json;version=1`",
            tags: ["api", "design", "backend"],
            created: 1704707400000,
            modified: 1704880200000
        },
        {
            id: "n9a0b1c2-d3e4-5678-2345-789012345678",
            title: "TypeScript Tips and Tricks",
            contents: "# TypeScript Tips and Tricks\n\n## Utility Types\n- `Partial<T>` - All properties optional\n- `Required<T>` - All properties required\n- `Pick<T, K>` - Select properties\n- `Omit<T, K>` - Exclude properties\n- `Record<K, V>` - Key-value object type\n\n## Discriminated Unions\n```typescript\ntype Result<T> = \n  | { success: true; data: T }\n  | { success: false; error: string };\n```\n\n## Template Literal Types\n```typescript\ntype EventName = `on${Capitalize<string>}`;\n// 'onClick', 'onSubmit', etc.\n```\n\n## Const Assertions\n```typescript\nconst config = {\n  endpoint: '/api',\n  timeout: 5000\n} as const;\n```",
            tags: ["typescript", "tips", "frontend"],
            created: 1704621000000,
            modified: 1704707400000
        },
        {
            id: "n0a1b2c3-d4e5-6789-3456-890123456789",
            title: "Reading List 2024",
            contents: "# Reading List 2024\n\n## Currently Reading\n- **Designing Data-Intensive Applications** by Martin Kleppmann\n  - Chapter 8: The Trouble with Distributed Systems\n\n## Completed\n1. Atomic Habits - James Clear\n2. The Staff Engineer's Path - Tanya Reilly\n\n## Up Next\n- Software Engineering at Google\n- A Philosophy of Software Design\n- Building Microservices (2nd Edition)\n- The Pragmatic Programmer (20th Anniversary)\n\n## Notes\nAim for 1 technical book per month\nAlternate with non-fiction for balance",
            tags: ["reading", "books", "learning"],
            created: 1704534600000,
            modified: 1705312200000
        },
        {
            id: "na1b2c3d-e4f5-6789-4567-901234567890",
            title: "Testing Strategy",
            contents: "# Testing Strategy\n\n## Testing Pyramid\n- Unit tests: 70%\n- Integration tests: 20%\n- E2E tests: 10%\n\n## Unit Testing Best Practices\n- Test behavior, not implementation\n- One assertion per test (generally)\n- Use descriptive test names\n- Mock external dependencies\n\n## Integration Testing\n- Test API endpoints\n- Use test database\n- Seed data for each test\n- Clean up after tests\n\n## E2E Testing with Playwright\n```typescript\ntest('user can complete checkout', async ({ page }) => {\n  await page.goto('/products');\n  await page.click('[data-testid=\"add-to-cart\"]');\n  await page.click('[data-testid=\"checkout\"]');\n  await expect(page).toHaveURL('/confirmation');\n});\n```",
            tags: ["testing", "quality", "development"],
            created: 1704448200000,
            modified: 1704621000000
        },
        {
            id: "nb2c3d4e-f5a6-7890-5678-012345678901",
            title: "AWS Services Overview",
            contents: "# AWS Services Overview\n\n## Compute\n- **EC2**: Virtual servers\n- **Lambda**: Serverless functions\n- **ECS/EKS**: Container orchestration\n- **Fargate**: Serverless containers\n\n## Storage\n- **S3**: Object storage\n- **EBS**: Block storage for EC2\n- **EFS**: File system for EC2\n\n## Database\n- **RDS**: Managed relational DBs\n- **DynamoDB**: NoSQL\n- **ElastiCache**: Redis/Memcached\n- **Aurora**: High-performance MySQL/PostgreSQL\n\n## Networking\n- **VPC**: Virtual network\n- **CloudFront**: CDN\n- **Route 53**: DNS\n- **API Gateway**: API management",
            tags: ["aws", "cloud", "infrastructure"],
            created: 1704361800000,
            modified: 1704448200000
        },
        {
            id: "nc3d4e5f-a6b7-8901-6789-123456789012",
            title: "CSS Grid Cheatsheet",
            contents: "# CSS Grid Cheatsheet\n\n## Container Properties\n```css\n.container {\n  display: grid;\n  grid-template-columns: repeat(3, 1fr);\n  grid-template-rows: auto;\n  gap: 1rem;\n  justify-items: center;\n  align-items: start;\n}\n```\n\n## Item Properties\n```css\n.item {\n  grid-column: 1 / 3;\n  grid-row: 1 / 2;\n  justify-self: end;\n  align-self: center;\n}\n```\n\n## Named Areas\n```css\n.container {\n  grid-template-areas:\n    'header header header'\n    'sidebar main main'\n    'footer footer footer';\n}\n\n.header { grid-area: header; }\n```\n\n## Auto-fit/Auto-fill\n```css\ngrid-template-columns: repeat(auto-fit, minmax(250px, 1fr));\n```",
            tags: ["css", "grid", "frontend"],
            created: 1704275400000,
            modified: 1704361800000
        },
        {
            id: "nd4e5f6a-b7c8-9012-7890-234567890123",
            title: "Monitoring and Alerting Setup",
            contents: "# Monitoring and Alerting\n\n## Key Metrics\n- Request latency (p50, p95, p99)\n- Error rate (5xx responses)\n- Request throughput\n- CPU/Memory utilization\n- Database connection pool\n\n## Alerting Rules\n| Metric | Warning | Critical |\n|--------|---------|----------|\n| Error Rate | > 1% | > 5% |\n| P99 Latency | > 500ms | > 2s |\n| CPU Usage | > 70% | > 90% |\n\n## Dashboards\n1. Service Overview\n2. Database Performance\n3. Infrastructure Health\n4. Business Metrics\n\n## Tools\n- Prometheus for metrics\n- Grafana for visualization\n- PagerDuty for alerting",
            tags: ["monitoring", "devops", "observability"],
            created: 1704189000000,
            modified: 1704275400000
        },
        {
            id: "ne5f6a7b-c8d9-0123-8901-345678901234",
            title: "Security Best Practices",
            contents: "# Security Best Practices\n\n## Authentication\n- Use industry standards (OAuth2, OIDC)\n- Implement MFA where possible\n- Secure token storage\n- Session management\n\n## Input Validation\n- Validate on both client and server\n- Sanitize user inputs\n- Use parameterized queries\n- Implement rate limiting\n\n## OWASP Top 10\n1. Broken Access Control\n2. Cryptographic Failures\n3. Injection\n4. Insecure Design\n5. Security Misconfiguration\n\n## Headers\n```\nContent-Security-Policy: default-src 'self'\nX-Frame-Options: DENY\nX-Content-Type-Options: nosniff\nStrict-Transport-Security: max-age=31536000\n```",
            tags: ["security", "best-practices", "backend"],
            created: 1704102600000,
            modified: 1704189000000
        }
    ],

    // ============================================
    // Contacts
    // ============================================
    contacts: [
        {
            contact_id: "ct1a2b3c-d4e5-6789-abcd-ef1234567890",
            name: "Alice Chen",
            email: "alice.chen@example.com",
            phone: "+1 (555) 123-4567",
            birthday: "1990-03-15",
            notes: "Met at React Conf 2023. Works at Stripe on the payments team.",
            creation_date: "2023-06-15T10:00:00Z",
            last_update: "2024-01-10T14:30:00Z",
            last_week_messages: 12,
            groups_in_common: 3,
            importance: 8,
            closeness: 7,
            fondness: 9,
            tags: [{ tag_id: "tg1", name: "tech" }, { tag_id: "tg2", name: "conference" }]
        },
        {
            contact_id: "ct2b3c4d-e5f6-7890-bcde-f12345678901",
            name: "Bob Martinez",
            email: "bob.m@company.com",
            phone: "+1 (555) 234-5678",
            birthday: "1985-07-22",
            notes: "College roommate. Now working in venture capital.",
            creation_date: "2020-01-10T08:00:00Z",
            last_update: "2024-01-08T09:15:00Z",
            last_week_messages: 5,
            groups_in_common: 2,
            importance: 9,
            closeness: 9,
            fondness: 10,
            tags: [{ tag_id: "tg3", name: "college" }, { tag_id: "tg4", name: "close-friend" }]
        },
        {
            contact_id: "ct3c4d5e-f6a7-8901-cdef-123456789012",
            name: "Carol Williams",
            email: "carol.williams@startup.io",
            phone: null,
            birthday: null,
            notes: "CEO of TechStartup. Potential investor contact.",
            creation_date: "2023-09-20T16:00:00Z",
            last_update: "2024-01-05T11:00:00Z",
            last_week_messages: 0,
            groups_in_common: 1,
            importance: 7,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg5", name: "business" }, { tag_id: "tg6", name: "investor" }]
        },
        {
            contact_id: "ct4d5e6f-a7b8-9012-def1-234567890123",
            name: "David Kim",
            email: "david.kim@google.com",
            phone: "+1 (555) 345-6789",
            birthday: "1992-11-08",
            notes: "Former colleague from previous job. Senior engineer at Google.",
            creation_date: "2021-03-15T14:00:00Z",
            last_update: "2024-01-12T16:45:00Z",
            last_week_messages: 8,
            groups_in_common: 4,
            importance: 8,
            closeness: 6,
            fondness: 8,
            tags: [{ tag_id: "tg1", name: "tech" }, { tag_id: "tg7", name: "ex-colleague" }]
        },
        {
            contact_id: "ct5e6f7a-b8c9-0123-ef12-345678901234",
            name: "Emma Thompson",
            email: "emma.t@design.co",
            phone: "+1 (555) 456-7890",
            birthday: "1988-04-30",
            notes: "Lead designer. Great collaborator on UI/UX projects.",
            creation_date: "2022-07-10T10:30:00Z",
            last_update: "2024-01-14T08:00:00Z",
            last_week_messages: 15,
            groups_in_common: 2,
            importance: 7,
            closeness: 5,
            fondness: 7,
            tags: [{ tag_id: "tg8", name: "design" }, { tag_id: "tg9", name: "collaborator" }]
        },
        {
            contact_id: "ct6f7a8b-c9d0-1234-f123-456789012345",
            name: "Frank Johnson",
            email: "frank.j@corp.com",
            phone: "+1 (555) 567-8901",
            birthday: "1980-09-12",
            notes: "Manager at previous company. Good mentor.",
            creation_date: "2019-05-20T09:00:00Z",
            last_update: "2023-12-20T12:00:00Z",
            last_week_messages: 2,
            groups_in_common: 1,
            importance: 6,
            closeness: 4,
            fondness: 6,
            tags: [{ tag_id: "tg7", name: "ex-colleague" }, { tag_id: "tg10", name: "mentor" }]
        },
        {
            contact_id: "ct7a8b9c-d0e1-2345-0123-567890123456",
            name: "Grace Lee",
            email: "grace.lee@university.edu",
            phone: null,
            birthday: "1995-01-25",
            notes: "PhD researcher in machine learning. Met at AI conference.",
            creation_date: "2023-11-05T15:00:00Z",
            last_update: "2024-01-11T10:30:00Z",
            last_week_messages: 6,
            groups_in_common: 1,
            importance: 5,
            closeness: 3,
            fondness: 6,
            tags: [{ tag_id: "tg11", name: "ai" }, { tag_id: "tg2", name: "conference" }]
        },
        {
            contact_id: "ct8b9c0d-e1f2-3456-1234-678901234567",
            name: "Henry Wilson",
            email: "henry.wilson@agency.com",
            phone: "+1 (555) 678-9012",
            birthday: "1987-06-18",
            notes: "Marketing consultant. Helped with product launch.",
            creation_date: "2022-04-12T11:00:00Z",
            last_update: "2024-01-03T14:15:00Z",
            last_week_messages: 0,
            groups_in_common: 1,
            importance: 4,
            closeness: 2,
            fondness: 4,
            tags: [{ tag_id: "tg12", name: "marketing" }, { tag_id: "tg5", name: "business" }]
        },
        {
            contact_id: "ct9c0d1e-f2a3-4567-2345-789012345678",
            name: "Isabella Garcia",
            email: "isabella@freelance.io",
            phone: "+1 (555) 789-0123",
            birthday: "1993-08-05",
            notes: "Freelance writer. Writes technical documentation.",
            creation_date: "2023-02-28T13:00:00Z",
            last_update: "2024-01-09T09:45:00Z",
            last_week_messages: 4,
            groups_in_common: 2,
            importance: 5,
            closeness: 4,
            fondness: 6,
            tags: [{ tag_id: "tg13", name: "freelance" }, { tag_id: "tg14", name: "writer" }]
        },
        {
            contact_id: "ct0d1e2f-a3b4-5678-3456-890123456789",
            name: "James Brown",
            email: "james.brown@startup.co",
            phone: "+1 (555) 890-1234",
            birthday: null,
            notes: "Co-founder of a dev tools startup. Great technical discussions.",
            creation_date: "2023-08-15T16:30:00Z",
            last_update: "2024-01-13T11:20:00Z",
            last_week_messages: 10,
            groups_in_common: 3,
            importance: 7,
            closeness: 5,
            fondness: 7,
            tags: [{ tag_id: "tg1", name: "tech" }, { tag_id: "tg15", name: "startup" }]
        },
        {
            contact_id: "cta1e2f3-b4c5-6789-4567-901234567890",
            name: "Karen Smith",
            email: "karen.s@law.com",
            phone: "+1 (555) 901-2345",
            birthday: "1982-12-03",
            notes: "Corporate lawyer. Handles legal matters.",
            creation_date: "2021-10-05T10:00:00Z",
            last_update: "2023-11-15T14:00:00Z",
            last_week_messages: 0,
            groups_in_common: 0,
            importance: 6,
            closeness: 2,
            fondness: 3,
            tags: [{ tag_id: "tg16", name: "legal" }, { tag_id: "tg5", name: "business" }]
        },
        {
            contact_id: "ctb2f3a4-c5d6-7890-5678-012345678901",
            name: "Leo Zhang",
            email: "leo.zhang@bigtech.com",
            phone: "+1 (555) 012-3456",
            birthday: "1991-02-28",
            notes: "Staff engineer. Expert in distributed systems.",
            creation_date: "2022-09-18T09:30:00Z",
            last_update: "2024-01-14T16:00:00Z",
            last_week_messages: 7,
            groups_in_common: 2,
            importance: 8,
            closeness: 6,
            fondness: 8,
            tags: [{ tag_id: "tg1", name: "tech" }, { tag_id: "tg17", name: "distributed-systems" }]
        },
        {
            contact_id: "ctc3a4b5-d6e7-8901-6789-123456789012",
            name: "Maria Rodriguez",
            email: "maria.r@nonprofit.org",
            phone: null,
            birthday: "1986-05-20",
            notes: "Runs a coding bootcamp for underrepresented groups.",
            creation_date: "2023-04-22T12:00:00Z",
            last_update: "2024-01-07T10:15:00Z",
            last_week_messages: 3,
            groups_in_common: 1,
            importance: 6,
            closeness: 4,
            fondness: 7,
            tags: [{ tag_id: "tg18", name: "education" }, { tag_id: "tg19", name: "nonprofit" }]
        },
        {
            contact_id: "ctd4b5c6-e7f8-9012-7890-234567890123",
            name: "Nathan Park",
            email: "nathan.park@media.com",
            phone: "+1 (555) 234-5670",
            birthday: "1994-10-15",
            notes: "Tech journalist. Covers AI and developer tools.",
            creation_date: "2023-07-08T14:45:00Z",
            last_update: "2024-01-10T08:30:00Z",
            last_week_messages: 1,
            groups_in_common: 1,
            importance: 5,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg20", name: "media" }, { tag_id: "tg11", name: "ai" }]
        },
        {
            contact_id: "cte5c6d7-f8a9-0123-8901-345678901234",
            name: "Olivia Taylor",
            email: "olivia.t@vc.fund",
            phone: "+1 (555) 345-6780",
            birthday: null,
            notes: "Partner at VC fund. Focuses on B2B SaaS investments.",
            creation_date: "2023-10-30T11:00:00Z",
            last_update: "2024-01-06T15:45:00Z",
            last_week_messages: 2,
            groups_in_common: 1,
            importance: 8,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg6", name: "investor" }, { tag_id: "tg21", name: "vc" }]
        },
        {
            contact_id: "ctf6d7e8-a9b0-1234-9012-456789012345",
            name: "Patrick O'Brien",
            email: "patrick.ob@consulting.com",
            phone: "+1 (555) 456-7891",
            birthday: "1979-08-30",
            notes: "Cloud architecture consultant. AWS certified.",
            creation_date: "2022-01-15T09:00:00Z",
            last_update: "2024-01-02T12:30:00Z",
            last_week_messages: 0,
            groups_in_common: 1,
            importance: 5,
            closeness: 3,
            fondness: 4,
            tags: [{ tag_id: "tg22", name: "aws" }, { tag_id: "tg23", name: "consulting" }]
        },
        {
            contact_id: "cta7e8f9-b0c1-2345-0123-567890123456",
            name: "Quinn Davis",
            email: "quinn.d@opensource.io",
            phone: null,
            birthday: "1996-03-08",
            notes: "Open source maintainer. Works on popular JS framework.",
            creation_date: "2023-05-12T16:00:00Z",
            last_update: "2024-01-11T14:20:00Z",
            last_week_messages: 9,
            groups_in_common: 2,
            importance: 6,
            closeness: 4,
            fondness: 7,
            tags: [{ tag_id: "tg24", name: "open-source" }, { tag_id: "tg1", name: "tech" }]
        },
        {
            contact_id: "ctb8f9a0-c1d2-3456-1234-678901234567",
            name: "Rachel Kim",
            email: "rachel.kim@product.co",
            phone: "+1 (555) 567-8902",
            birthday: "1990-11-22",
            notes: "Product manager at fintech company. Great product sense.",
            creation_date: "2022-11-08T10:30:00Z",
            last_update: "2024-01-13T09:00:00Z",
            last_week_messages: 5,
            groups_in_common: 2,
            importance: 7,
            closeness: 5,
            fondness: 7,
            tags: [{ tag_id: "tg25", name: "product" }, { tag_id: "tg26", name: "fintech" }]
        },
        {
            contact_id: "ctc9a0b1-d2e3-4567-2345-789012345678",
            name: "Samuel Wright",
            email: "sam.wright@devops.io",
            phone: "+1 (555) 678-9013",
            birthday: "1988-07-04",
            notes: "DevOps lead. Expert in Kubernetes and CI/CD.",
            creation_date: "2021-08-20T13:00:00Z",
            last_update: "2024-01-08T16:30:00Z",
            last_week_messages: 4,
            groups_in_common: 3,
            importance: 7,
            closeness: 5,
            fondness: 6,
            tags: [{ tag_id: "tg27", name: "devops" }, { tag_id: "tg28", name: "kubernetes" }]
        },
        {
            contact_id: "ctd0b1c2-e3f4-5678-3456-890123456789",
            name: "Tina Nguyen",
            email: "tina.n@mobile.dev",
            phone: null,
            birthday: "1993-04-12",
            notes: "Mobile developer. React Native and Flutter expert.",
            creation_date: "2023-03-05T11:00:00Z",
            last_update: "2024-01-12T10:00:00Z",
            last_week_messages: 6,
            groups_in_common: 2,
            importance: 6,
            closeness: 4,
            fondness: 6,
            tags: [{ tag_id: "tg29", name: "mobile" }, { tag_id: "tg1", name: "tech" }]
        },
        {
            contact_id: "cte1c2d3-f4a5-6789-4567-901234567890",
            name: "Victor Hernandez",
            email: "victor.h@security.co",
            phone: "+1 (555) 789-0124",
            birthday: "1984-09-18",
            notes: "Security engineer. Helps with security audits.",
            creation_date: "2022-06-30T14:00:00Z",
            last_update: "2024-01-05T11:45:00Z",
            last_week_messages: 1,
            groups_in_common: 1,
            importance: 6,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg30", name: "security" }, { tag_id: "tg1", name: "tech" }]
        },
        {
            contact_id: "ctf2d3e4-a5b6-7890-5678-012345678901",
            name: "Wendy Chen",
            email: "wendy.chen@data.ai",
            phone: "+1 (555) 890-1235",
            birthday: "1991-12-08",
            notes: "Data scientist. Works on ML pipelines at scale.",
            creation_date: "2023-01-18T09:30:00Z",
            last_update: "2024-01-14T12:00:00Z",
            last_week_messages: 8,
            groups_in_common: 2,
            importance: 7,
            closeness: 5,
            fondness: 7,
            tags: [{ tag_id: "tg31", name: "data-science" }, { tag_id: "tg11", name: "ai" }]
        },
        {
            contact_id: "cta3e4f5-b6c7-8901-6789-123456789012",
            name: "Xavier Jones",
            email: "xavier.j@gaming.co",
            phone: null,
            birthday: "1997-06-25",
            notes: "Game developer. Unity and Unreal specialist.",
            creation_date: "2023-08-25T15:30:00Z",
            last_update: "2024-01-07T14:00:00Z",
            last_week_messages: 3,
            groups_in_common: 1,
            importance: 4,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg32", name: "gaming" }, { tag_id: "tg1", name: "tech" }]
        },
        {
            contact_id: "ctb4f5a6-c7d8-9012-7890-234567890123",
            name: "Yuki Tanaka",
            email: "yuki.t@localization.io",
            phone: "+1 (555) 901-2346",
            birthday: "1989-02-14",
            notes: "Localization expert. Helps with Japanese market expansion.",
            creation_date: "2023-06-10T10:00:00Z",
            last_update: "2024-01-04T09:30:00Z",
            last_week_messages: 2,
            groups_in_common: 1,
            importance: 5,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg33", name: "localization" }, { tag_id: "tg5", name: "business" }]
        },
        {
            contact_id: "ctc5a6b7-d8e9-0123-8901-345678901234",
            name: "Zoe Anderson",
            email: "zoe.a@accelerator.vc",
            phone: "+1 (555) 012-3457",
            birthday: "1992-10-30",
            notes: "Runs startup accelerator program. Good for intro to other founders.",
            creation_date: "2023-09-02T12:00:00Z",
            last_update: "2024-01-10T16:15:00Z",
            last_week_messages: 4,
            groups_in_common: 2,
            importance: 7,
            closeness: 4,
            fondness: 6,
            tags: [{ tag_id: "tg15", name: "startup" }, { tag_id: "tg21", name: "vc" }]
        },
        {
            contact_id: "ctd6b7c8-e9f0-1234-9012-456789012345",
            name: "Aaron Mitchell",
            email: "aaron.m@enterprise.com",
            phone: "+1 (555) 123-4568",
            birthday: "1981-05-05",
            notes: "Enterprise sales. Helped close several large deals.",
            creation_date: "2021-12-10T11:00:00Z",
            last_update: "2024-01-03T10:00:00Z",
            last_week_messages: 1,
            groups_in_common: 1,
            importance: 6,
            closeness: 3,
            fondness: 4,
            tags: [{ tag_id: "tg34", name: "sales" }, { tag_id: "tg5", name: "business" }]
        },
        {
            contact_id: "cte7c8d9-f0a1-2345-0123-567890123456",
            name: "Bella Foster",
            email: "bella.f@hr.corp",
            phone: null,
            birthday: "1987-08-17",
            notes: "HR director. Good resource for hiring best practices.",
            creation_date: "2022-03-22T14:30:00Z",
            last_update: "2023-12-15T09:00:00Z",
            last_week_messages: 0,
            groups_in_common: 1,
            importance: 4,
            closeness: 2,
            fondness: 4,
            tags: [{ tag_id: "tg35", name: "hr" }, { tag_id: "tg5", name: "business" }]
        },
        {
            contact_id: "ctf8d9e0-a1b2-3456-1234-678901234567",
            name: "Chris Evans",
            email: "chris.e@community.dev",
            phone: "+1 (555) 234-5679",
            birthday: "1994-01-20",
            notes: "Developer relations at major tech company.",
            creation_date: "2023-04-05T16:00:00Z",
            last_update: "2024-01-12T11:30:00Z",
            last_week_messages: 5,
            groups_in_common: 3,
            importance: 6,
            closeness: 4,
            fondness: 6,
            tags: [{ tag_id: "tg36", name: "devrel" }, { tag_id: "tg1", name: "tech" }]
        },
        {
            contact_id: "cta9e0f1-b2c3-4567-2345-789012345678",
            name: "Diana Price",
            email: "diana.p@podcast.fm",
            phone: "+1 (555) 345-6781",
            birthday: null,
            notes: "Hosts popular tech podcast. Good for PR opportunities.",
            creation_date: "2023-07-20T10:00:00Z",
            last_update: "2024-01-08T14:45:00Z",
            last_week_messages: 2,
            groups_in_common: 1,
            importance: 5,
            closeness: 3,
            fondness: 5,
            tags: [{ tag_id: "tg20", name: "media" }, { tag_id: "tg37", name: "podcast" }]
        },
        {
            contact_id: "ctb0f1a2-c3d4-5678-3456-890123456789",
            name: "Ethan Brooks",
            email: "ethan.b@analytics.io",
            phone: null,
            birthday: "1990-06-12",
            notes: "Analytics expert. Built several data pipelines together.",
            creation_date: "2022-08-15T13:00:00Z",
            last_update: "2024-01-11T08:00:00Z",
            last_week_messages: 6,
            groups_in_common: 2,
            importance: 6,
            closeness: 5,
            fondness: 6,
            tags: [{ tag_id: "tg38", name: "analytics" }, { tag_id: "tg31", name: "data-science" }]
        }
    ],

    // ============================================
    // Messages
    // ============================================
    messages: [],

    // ============================================
    // Rooms
    // ============================================
    rooms: [
        {
            room_id: "rm1a2b3c-d4e5-6789-abcd-ef1234567890",
            display_name: "Engineering Team",
            user_defined_name: null,
            source_id: "matrix_eng_team",
            last_activity: "2024-01-14T16:30:00Z",
            participant_count: 12
        },
        {
            room_id: "rm2b3c4d-e5f6-7890-bcde-f12345678901",
            display_name: "Product Discussion",
            user_defined_name: "Product Sync",
            source_id: "matrix_product",
            last_activity: "2024-01-14T15:45:00Z",
            participant_count: 8
        },
        {
            room_id: "rm3c4d5e-f6a7-8901-cdef-123456789012",
            display_name: "Design Feedback",
            user_defined_name: null,
            source_id: "matrix_design",
            last_activity: "2024-01-14T14:20:00Z",
            participant_count: 6
        },
        {
            room_id: "rm4d5e6f-a7b8-9012-def1-234567890123",
            display_name: null,
            user_defined_name: "Alice Chen DM",
            source_id: "matrix_dm_alice",
            last_activity: "2024-01-14T12:00:00Z",
            participant_count: 2
        },
        {
            room_id: "rm5e6f7a-b8c9-0123-ef12-345678901234",
            display_name: "Frontend Guild",
            user_defined_name: null,
            source_id: "matrix_frontend",
            last_activity: "2024-01-13T18:30:00Z",
            participant_count: 15
        },
        {
            room_id: "rm6f7a8b-c9d0-1234-f123-456789012345",
            display_name: "Infrastructure",
            user_defined_name: "Infra Alerts",
            source_id: "matrix_infra",
            last_activity: "2024-01-14T11:15:00Z",
            participant_count: 5
        },
        {
            room_id: "rm7a8b9c-d0e1-2345-0123-567890123456",
            display_name: null,
            user_defined_name: "Bob Martinez DM",
            source_id: "matrix_dm_bob",
            last_activity: "2024-01-12T20:00:00Z",
            participant_count: 2
        },
        {
            room_id: "rm8b9c0d-e1f2-3456-1234-678901234567",
            display_name: "Open Source Contributors",
            user_defined_name: null,
            source_id: "matrix_oss",
            last_activity: "2024-01-14T10:45:00Z",
            participant_count: 25
        },
        {
            room_id: "rm9c0d1e-f2a3-4567-2345-789012345678",
            display_name: "Hiring Committee",
            user_defined_name: null,
            source_id: "matrix_hiring",
            last_activity: "2024-01-11T16:00:00Z",
            participant_count: 4
        },
        {
            room_id: "rm0d1e2f-a3b4-5678-3456-890123456789",
            display_name: "Random",
            user_defined_name: "Watercooler",
            source_id: "matrix_random",
            last_activity: "2024-01-14T17:00:00Z",
            participant_count: 30
        }
    ],

    // ============================================
    // Entities
    // ============================================
    entities: [
        {
            entity_id: "en1a2b3c-d4e5-6789-abcd-ef1234567890",
            name: "React",
            type: "technology",
            description: "A JavaScript library for building user interfaces",
            properties: { website: "https://react.dev", category: "frontend" },
            created_at: "2023-01-15T10:00:00Z",
            updated_at: "2024-01-10T14:30:00Z"
        },
        {
            entity_id: "en2b3c4d-e5f6-7890-bcde-f12345678901",
            name: "TypeScript",
            type: "technology",
            description: "A strongly typed programming language that builds on JavaScript",
            properties: { website: "https://typescriptlang.org", category: "language" },
            created_at: "2023-01-15T10:00:00Z",
            updated_at: "2024-01-10T14:30:00Z"
        },
        {
            entity_id: "en3c4d5e-f6a7-8901-cdef-123456789012",
            name: "PostgreSQL",
            type: "technology",
            description: "Open source relational database management system",
            properties: { website: "https://postgresql.org", category: "database" },
            created_at: "2023-02-20T12:00:00Z",
            updated_at: "2024-01-08T09:15:00Z"
        },
        {
            entity_id: "en4d5e6f-a7b8-9012-def1-234567890123",
            name: "Acme Corp",
            type: "organization",
            description: "Technology company focused on developer tools",
            properties: { industry: "technology", size: "startup" },
            created_at: "2023-03-10T14:00:00Z",
            updated_at: "2024-01-05T11:00:00Z"
        },
        {
            entity_id: "en5e6f7a-b8c9-0123-ef12-345678901234",
            name: "Q1 Product Launch",
            type: "project",
            description: "Major product release planned for Q1 2024",
            properties: { status: "in-progress", deadline: "2024-03-31" },
            created_at: "2023-10-01T09:00:00Z",
            updated_at: "2024-01-14T16:00:00Z"
        },
        {
            entity_id: "en6f7a8b-c9d0-1234-f123-456789012345",
            name: "Kubernetes",
            type: "technology",
            description: "Container orchestration platform",
            properties: { website: "https://kubernetes.io", category: "infrastructure" },
            created_at: "2023-04-15T11:30:00Z",
            updated_at: "2024-01-12T10:00:00Z"
        },
        {
            entity_id: "en7a8b9c-d0e1-2345-0123-567890123456",
            name: "GraphQL",
            type: "technology",
            description: "Query language for APIs",
            properties: { website: "https://graphql.org", category: "api" },
            created_at: "2023-05-20T14:00:00Z",
            updated_at: "2024-01-09T08:45:00Z"
        },
        {
            entity_id: "en8b9c0d-e1f2-3456-1234-678901234567",
            name: "TechConf 2024",
            type: "event",
            description: "Annual technology conference",
            properties: { date: "2024-06-15", location: "San Francisco" },
            created_at: "2023-11-01T10:00:00Z",
            updated_at: "2024-01-11T15:30:00Z"
        },
        {
            entity_id: "en9c0d1e-f2a3-4567-2345-789012345678",
            name: "OpenAI",
            type: "organization",
            description: "AI research and deployment company",
            properties: { industry: "AI", size: "enterprise" },
            created_at: "2023-06-10T09:00:00Z",
            updated_at: "2024-01-13T11:00:00Z"
        },
        {
            entity_id: "en0d1e2f-a3b4-5678-3456-890123456789",
            name: "API Redesign",
            type: "project",
            description: "Complete redesign of public API",
            properties: { status: "planning", priority: "high" },
            created_at: "2023-12-01T14:00:00Z",
            updated_at: "2024-01-14T12:00:00Z"
        },
        {
            entity_id: "ena1e2f3-b4c5-6789-4567-901234567890",
            name: "Vercel",
            type: "organization",
            description: "Cloud platform for frontend frameworks",
            properties: { industry: "cloud", website: "https://vercel.com" },
            created_at: "2023-07-15T10:30:00Z",
            updated_at: "2024-01-10T09:00:00Z"
        },
        {
            entity_id: "enb2f3a4-c5d6-7890-5678-012345678901",
            name: "Redis",
            type: "technology",
            description: "In-memory data structure store",
            properties: { website: "https://redis.io", category: "database" },
            created_at: "2023-08-20T11:00:00Z",
            updated_at: "2024-01-08T14:15:00Z"
        }
    ],

    // ============================================
    // Browser History
    // ============================================
    browserHistory: [
        {
            id: 1,
            url: "https://github.com/features/actions",
            title: "GitHub Actions - GitHub Features",
            visit_date: "2024-01-14T16:45:00Z",
            domain: "github.com"
        },
        {
            id: 2,
            url: "https://docs.anthropic.com/claude/reference/getting-started",
            title: "Getting Started - Anthropic API",
            visit_date: "2024-01-14T16:30:00Z",
            domain: "docs.anthropic.com"
        },
        {
            id: 3,
            url: "https://stackoverflow.com/questions/tagged/typescript",
            title: "Questions tagged [typescript] - Stack Overflow",
            visit_date: "2024-01-14T15:20:00Z",
            domain: "stackoverflow.com"
        },
        {
            id: 4,
            url: "https://news.ycombinator.com/",
            title: "Hacker News",
            visit_date: "2024-01-14T14:00:00Z",
            domain: "news.ycombinator.com"
        },
        {
            id: 5,
            url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            title: "System Design Interview Tips - YouTube",
            visit_date: "2024-01-14T12:30:00Z",
            domain: "youtube.com"
        },
        {
            id: 6,
            url: "https://twitter.com/home",
            title: "Home / X",
            visit_date: "2024-01-14T11:15:00Z",
            domain: "twitter.com"
        },
        {
            id: 7,
            url: "https://linear.app/team/issues",
            title: "Issues - Linear",
            visit_date: "2024-01-14T10:00:00Z",
            domain: "linear.app"
        },
        {
            id: 8,
            url: "https://notion.so/workspace",
            title: "Workspace - Notion",
            visit_date: "2024-01-14T09:30:00Z",
            domain: "notion.so"
        },
        {
            id: 9,
            url: "https://calendar.google.com/",
            title: "Google Calendar",
            visit_date: "2024-01-14T09:00:00Z",
            domain: "calendar.google.com"
        },
        {
            id: 10,
            url: "https://mail.google.com/",
            title: "Gmail - Inbox",
            visit_date: "2024-01-14T08:45:00Z",
            domain: "mail.google.com"
        },
        {
            id: 11,
            url: "https://react.dev/reference/react/useState",
            title: "useState - React",
            visit_date: "2024-01-13T17:30:00Z",
            domain: "react.dev"
        },
        {
            id: 12,
            url: "https://tailwindcss.com/docs/flex",
            title: "Flex - Tailwind CSS",
            visit_date: "2024-01-13T16:45:00Z",
            domain: "tailwindcss.com"
        },
        {
            id: 13,
            url: "https://stripe.com/docs/payments/quickstart",
            title: "Quickstart - Stripe Payments",
            visit_date: "2024-01-13T15:20:00Z",
            domain: "stripe.com"
        },
        {
            id: 14,
            url: "https://vercel.com/dashboard",
            title: "Dashboard - Vercel",
            visit_date: "2024-01-13T14:00:00Z",
            domain: "vercel.com"
        },
        {
            id: 15,
            url: "https://www.figma.com/file/abc123",
            title: "App Redesign - Figma",
            visit_date: "2024-01-13T11:30:00Z",
            domain: "figma.com"
        },
        {
            id: 16,
            url: "https://supabase.com/docs/guides/auth",
            title: "Auth - Supabase",
            visit_date: "2024-01-13T10:15:00Z",
            domain: "supabase.com"
        },
        {
            id: 17,
            url: "https://platform.openai.com/docs/guides/gpt",
            title: "GPT Guide - OpenAI Platform",
            visit_date: "2024-01-12T16:00:00Z",
            domain: "platform.openai.com"
        },
        {
            id: 18,
            url: "https://www.prisma.io/docs/concepts/components/prisma-schema",
            title: "Prisma Schema - Prisma Docs",
            visit_date: "2024-01-12T14:30:00Z",
            domain: "prisma.io"
        },
        {
            id: 19,
            url: "https://nextjs.org/docs/app/api-reference/functions/next-response",
            title: "NextResponse - Next.js",
            visit_date: "2024-01-12T13:00:00Z",
            domain: "nextjs.org"
        },
        {
            id: 20,
            url: "https://www.postgresql.org/docs/current/indexes.html",
            title: "Indexes - PostgreSQL Documentation",
            visit_date: "2024-01-12T11:45:00Z",
            domain: "postgresql.org"
        }
    ],

    // ============================================
    // Social Posts
    // ============================================
    socialPosts: [
        {
            post_id: "sp1a2b3c-d4e5-6789-abcd-ef1234567890",
            content: "Just deployed our new API redesign. Performance improved by 40%! Excited to see how users respond.",
            twitter_post_id: "1750123456789012345",
            bluesky_post_id: "3kfg7h2n9p1m",
            created_at: "2024-01-14T10:00:00Z",
            status: "posted"
        },
        {
            post_id: "sp2b3c4d-e5f6-7890-bcde-f12345678901",
            content: "Interesting thread on distributed systems design patterns. The CAP theorem discussion is particularly relevant.",
            twitter_post_id: "1749987654321098765",
            bluesky_post_id: null,
            created_at: "2024-01-13T15:30:00Z",
            status: "posted"
        },
        {
            post_id: "sp3c4d5e-f6a7-8901-cdef-123456789012",
            content: "Working on a comprehensive guide to React Server Components. Should be ready to publish next week.",
            twitter_post_id: null,
            bluesky_post_id: "3kfg8j3o0q2n",
            created_at: "2024-01-12T09:15:00Z",
            status: "posted"
        },
        {
            post_id: "sp4d5e6f-a7b8-9012-def1-234567890123",
            content: "TypeScript 5.4 beta looks promising. The new NoInfer utility type will help with generic inference issues.",
            twitter_post_id: null,
            bluesky_post_id: null,
            created_at: "2024-01-11T14:00:00Z",
            status: "draft"
        },
        {
            post_id: "sp5e6f7a-b8c9-0123-ef12-345678901234",
            content: "Hot take: Most microservices architectures would be better as modular monoliths.",
            twitter_post_id: "1748555666777888999",
            bluesky_post_id: "3kfg9k4p1r3o",
            created_at: "2024-01-10T11:45:00Z",
            status: "posted"
        },
        {
            post_id: "sp6f7a8b-c9d0-1234-f123-456789012345",
            content: "Testing a new social post about PostgreSQL query optimization tips.",
            twitter_post_id: null,
            bluesky_post_id: null,
            created_at: "2024-01-09T16:20:00Z",
            status: "failed",
            error_message: "Rate limit exceeded. Please try again later."
        },
        {
            post_id: "sp7a8b9c-d0e1-2345-0123-567890123456",
            content: "Preparing content about edge computing and its impact on web application architecture.",
            twitter_post_id: null,
            bluesky_post_id: null,
            created_at: "2024-01-08T10:00:00Z",
            status: "pending"
        }
    ],

    // ============================================
    // Tags (Note Tags)
    // ============================================
    tags: [
        { id: "tag1", name: "interview", created: 1704067200000, modified: 1704067200000, usage_count: 3 },
        { id: "tag2", name: "system-design", created: 1704067200000, modified: 1704067200000, usage_count: 4 },
        { id: "tag3", name: "career", created: 1704067200000, modified: 1704067200000, usage_count: 2 },
        { id: "tag4", name: "react", created: 1704067200000, modified: 1704153600000, usage_count: 5 },
        { id: "tag5", name: "performance", created: 1704067200000, modified: 1704067200000, usage_count: 3 },
        { id: "tag6", name: "frontend", created: 1704067200000, modified: 1704240000000, usage_count: 8 },
        { id: "tag7", name: "weekly-goals", created: 1704067200000, modified: 1704326400000, usage_count: 4 },
        { id: "tag8", name: "planning", created: 1704067200000, modified: 1704326400000, usage_count: 3 },
        { id: "tag9", name: "database", created: 1704067200000, modified: 1704067200000, usage_count: 5 },
        { id: "tag10", name: "postgresql", created: 1704067200000, modified: 1704067200000, usage_count: 2 },
        { id: "tag11", name: "optimization", created: 1704067200000, modified: 1704067200000, usage_count: 3 },
        { id: "tag12", name: "meeting-notes", created: 1704067200000, modified: 1704067200000, usage_count: 5 },
        { id: "tag13", name: "roadmap", created: 1704067200000, modified: 1704067200000, usage_count: 2 },
        { id: "tag14", name: "product", created: 1704067200000, modified: 1704067200000, usage_count: 4 },
        { id: "tag15", name: "git", created: 1704067200000, modified: 1704067200000, usage_count: 2 },
        { id: "tag16", name: "workflow", created: 1704067200000, modified: 1704067200000, usage_count: 2 },
        { id: "tag17", name: "best-practices", created: 1704067200000, modified: 1704153600000, usage_count: 6 },
        { id: "tag18", name: "docker", created: 1704067200000, modified: 1704067200000, usage_count: 3 },
        { id: "tag19", name: "development", created: 1704067200000, modified: 1704153600000, usage_count: 7 },
        { id: "tag20", name: "devops", created: 1704067200000, modified: 1704240000000, usage_count: 5 },
        { id: "tag21", name: "api", created: 1704067200000, modified: 1704067200000, usage_count: 4 },
        { id: "tag22", name: "design", created: 1704067200000, modified: 1704067200000, usage_count: 3 },
        { id: "tag23", name: "backend", created: 1704067200000, modified: 1704153600000, usage_count: 5 },
        { id: "tag24", name: "typescript", created: 1704067200000, modified: 1704067200000, usage_count: 4 },
        { id: "tag25", name: "tips", created: 1704067200000, modified: 1704067200000, usage_count: 2 },
        { id: "tag26", name: "reading", created: 1704067200000, modified: 1704326400000, usage_count: 2 },
        { id: "tag27", name: "books", created: 1704067200000, modified: 1704326400000, usage_count: 2 },
        { id: "tag28", name: "learning", created: 1704067200000, modified: 1704326400000, usage_count: 4 },
        { id: "tag29", name: "testing", created: 1704067200000, modified: 1704067200000, usage_count: 3 },
        { id: "tag30", name: "quality", created: 1704067200000, modified: 1704067200000, usage_count: 2 }
    ],

    // ============================================
    // Categories (Bookmark Categories)
    // ============================================
    categories: [
        { category_id: "cat1", name: "Development", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat2", name: "Design", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat3", name: "Database", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat4", name: "DevOps", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat5", name: "API", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat6", name: "Architecture", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat7", name: "Testing", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat8", name: "AI", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat9", name: "UX", created_at: "2023-01-01T00:00:00Z" },
        { category_id: "cat10", name: "Infrastructure", created_at: "2023-01-01T00:00:00Z" }
    ],

    // ============================================
    // Dashboard Stats
    // ============================================
    dashboardStats: {
        contacts: {
            total: 30,
            recentCount: 8,
            recentlyActive: 15,
            monthOverMonthChange: 12.5
        },
        sessions: {
            total: 156,
            recentCount: 23,
            monthOverMonthChange: -5.2
        },
        bookmarks: {
            total: 22,
            recentCount: 12,
            monthOverMonthChange: 18.3
        },
        history: {
            total: 1250,
            recentCount: 156,
            monthOverMonthChange: 8.7
        },
        recentItems: [
            { id: "1", category: "bookmark", name: "React Documentation", date: "2024-01-14T10:30:00Z" },
            { id: "2", category: "note", name: "System Design Notes", date: "2024-01-14T09:15:00Z" },
            { id: "3", category: "contact", name: "Alice Chen", date: "2024-01-14T08:45:00Z" },
            { id: "4", category: "bookmark", name: "TypeScript Handbook", date: "2024-01-13T16:20:00Z" },
            { id: "5", category: "note", name: "Weekly Goals", date: "2024-01-13T14:00:00Z" }
        ]
    },

    // ============================================
    // Configuration
    // ============================================
    configuration: {
        // General Settings
        general: {
            displayName: "Garden User",
            email: "user@example.com",
            language: "en",
            timezone: "America/New_York",
            defaultView: {
                dashboard: "overview",
                bookmarks: "grid",
                notes: "grid",
                contacts: "list",
                messages: "threaded"
            }
        },
        // Appearance Settings
        appearance: {
            theme: "dark",
            accentColor: "#0070f3",
            sidebarPosition: "left",
            compactMode: false,
            fontSize: 16,
            codeFont: "SF Mono",
            showPreview: true
        },
        // Integrations
        integrations: {
            ollama: {
                enabled: true,
                url: "http://localhost:11434",
                embeddingModel: "nomic-embed-text",
                llmModel: "llama3.2",
                lastConnected: "2024-01-14T10:30:00Z",
                status: "connected"
            },
            twitter: {
                connected: true,
                username: "@gardenuser",
                autoPost: false,
                connectedAt: "2023-12-01T09:00:00Z"
            },
            bluesky: {
                connected: true,
                handle: "garden.bsky.social",
                autoPost: false,
                connectedAt: "2023-12-15T14:00:00Z"
            },
            browserExtension: {
                installed: true,
                version: "1.2.3",
                autoCapture: true,
                captureHistory: true,
                lastSync: "2024-01-14T16:00:00Z"
            },
            logseq: {
                enabled: true,
                syncPath: "/home/user/logseq",
                autoSync: true,
                syncInterval: 30,
                lastSync: "2024-01-14T15:30:00Z",
                conflictResolution: "newer",
                status: "synced"
            }
        },
        // Data & Storage
        storage: {
            totalUsed: 256000000,
            breakdown: {
                bookmarks: 45000000,
                notes: 32000000,
                contacts: 8000000,
                messages: 120000000,
                cache: 35000000,
                other: 16000000
            },
            stats: {
                bookmarks: 22,
                notes: 15,
                contacts: 30,
                messages: 156,
                entities: 12,
                historyItems: 1250
            }
        },
        // Privacy Settings
        privacy: {
            dataRetention: {
                enabled: true,
                historyDays: 90,
                messageDays: 365,
                deletedItemsDays: 30
            },
            autoDelete: {
                enabled: false,
                interval: "monthly"
            },
            anonymizeExports: true,
            shareAnalytics: false
        },
        // Advanced Settings
        advanced: {
            apiEndpoint: "http://localhost:8000",
            debugMode: false,
            consoleLogging: false,
            experimentalFeatures: false,
            rawConfig: {}
        },
        // System Info
        system: {
            version: "1.0.0",
            buildNumber: "2024.01.14.001",
            nodeVersion: "20.10.0",
            platform: "linux",
            dbVersion: "1",
            lastUpdated: "2024-01-14T12:00:00Z"
        },
        // Notifications (legacy support)
        notifications: {
            email: true,
            push: true,
            digest: "daily"
        },
        // Sync status
        sync: {
            enabled: true,
            lastSync: "2024-01-14T16:45:00Z",
            status: "synced"
        }
    }
};

// Generate messages for rooms
(function generateMessages() {
    const senders = MockData.contacts.slice(0, 10);
    const messageTemplates = [
        "Just pushed the latest changes to the repo. Can someone review?",
        "Has anyone looked into the performance issues we discussed?",
        "The new design looks great! A few minor suggestions in the Figma comments.",
        "Meeting moved to 3pm. Same link.",
        "Quick question - what's our approach for handling the edge cases?",
        "I'll take a look at that bug this afternoon.",
        "Great progress on the sprint! Keep it up team.",
        "Can we sync on the API changes tomorrow?",
        "Just submitted my PR for the auth refactor.",
        "Anyone available for a quick code review?",
        "The tests are passing now. Ready for QA.",
        "I updated the documentation. Let me know if anything's unclear.",
        "Found the root cause of the issue. Working on a fix.",
        "Let's discuss this in the standup tomorrow.",
        "Deployed to staging. Please test when you get a chance.",
        "Good catch! I'll fix that right away.",
        "The client loved the demo!",
        "Need some help debugging this async issue.",
        "Merged! Thanks for the thorough review.",
        "Just a heads up - I'll be OOO tomorrow."
    ];

    let messageId = 1;
    MockData.rooms.forEach(room => {
        const messageCount = Math.floor(Math.random() * 8) + 3;
        for (let i = 0; i < messageCount; i++) {
            const sender = senders[Math.floor(Math.random() * senders.length)];
            const daysAgo = Math.floor(Math.random() * 7);
            const hoursAgo = Math.floor(Math.random() * 24);
            const date = new Date();
            date.setDate(date.getDate() - daysAgo);
            date.setHours(date.getHours() - hoursAgo);

            MockData.messages.push({
                message_id: `msg${String(messageId++).padStart(6, '0')}-${room.room_id.substring(0, 8)}`,
                sender_contact_id: sender.contact_id,
                room_id: room.room_id,
                event_id: `evt_${Date.now()}_${messageId}`,
                event_datetime: date.toISOString(),
                body: messageTemplates[Math.floor(Math.random() * messageTemplates.length)],
                formatted_body: null,
                message_type: "m.text",
                sender_name: sender.name,
                sender_email: sender.email
            });
        }
    });

    // Sort messages by date descending
    MockData.messages.sort((a, b) => new Date(b.event_datetime) - new Date(a.event_datetime));
})();

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MockData;
}
