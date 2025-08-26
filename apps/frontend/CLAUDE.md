When working on app features, follow this folder structure:

## CORE STRUCTURE (REQUIRED)
- `api/` → API concerns (hooks, services, types, models)
- `components/` → UI components (ui/, views/, forms/)
- `(routes)/` → Next.js pages (page.tsx files only)

## OPTIONAL STRUCTURE (CREATE AS NEEDED)
- `contexts/` → React contexts (grouped by feature)
- `hooks/` → General React hooks (non-API hooks)
- `lib/` → Utilities and constants
- `types/` → Frontend-specific types

## PLACEMENT GUIDELINES

### API Directory (`api/`)
- `api/hooks/` → TanStack Query hooks, data fetching
- `api/services/` → Raw API functions, HTTP clients
- `api/types/` → API response types, request payloads
- `api/models/` → Data transformation, factory functions

### Components Directory (`components/`)
- `components/views/` → Page-level components, main views
- `components/ui/` → Reusable UI components, design system
- `components/forms/` → Form components, input controls

### Optional Directories
- `contexts/{feature}/` → Feature-specific contexts
  - `Context.ts` → Context definition
  - `Provider.tsx` → Context provider component
  - `useContext.ts` → Context hook
- `hooks/` → Custom React hooks (non-API)
- `lib/utils/` → Helper functions, formatters
- `lib/constants/` → App constants, configurations
- `types/` → Component props, form states, frontend types

## AI AGENT GUIDANCE

**File Placement Decision Tree:**
1. **API-related?** → `api/` (always create this folder)
2. **UI Component?** → `components/` (always create this folder)
3. **Next.js page?** → `(routes)/` (always create this folder)
4. **React context?** → `contexts/` (create if multiple contexts exist)
5. **Custom hook?** → `hooks/` (create if multiple hooks exist)
6. **Utility function?** → `lib/` (create if utilities are substantial)
7. **Frontend type?** → `types/` (create if types are complex)

**When to create optional folders:**
- Create `contexts/` when you have 2+ context providers
- Create `hooks/` when you have 3+ custom hooks
- Create `lib/` when you have utility functions beyond simple helpers
- Create `types/` when you have component props/state types separate from API types

**Folder Examples:**
```
✅ Good structure:
├── api/
│   ├── hooks/useContacts.ts
│   ├── services/contactsApi.ts
│   └── types/index.ts
├── components/
│   ├── views/ContactsView.tsx
│   ├── ui/ContactCard.tsx
│   └── forms/ContactForm.tsx
└── (routes)/
    └── page.tsx

⚠️ Minimal structure (also acceptable):
├── api/
│   └── contactsApi.ts
├── components/
│   └── ContactsView.tsx
└── (routes)/
    └── page.tsx
```

Follow this structure for consistent app development.
