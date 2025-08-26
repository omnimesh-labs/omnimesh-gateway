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

## FORM FIELD STYLING & ACCESSIBILITY

### Form Field Background Colors

The application uses two styling systems that need proper contrast for ADA compliance:

#### 1. Custom UI Components (Tailwind)
Located in `src/components/ui/`:
- **Input** (`input.tsx`): Light mode `bg-gray-50`, Dark mode `dark:bg-gray-800`
- **Textarea** (`textarea.tsx`): Light mode `bg-gray-50`, Dark mode `dark:bg-gray-800`
- **Select** (`select.tsx`): Light mode `bg-gray-50`, Dark mode `dark:bg-gray-800`

To adjust these:
```tsx
// Example: Change input background colors
className={cn(
  'bg-gray-50',           // Light mode background
  'dark:bg-gray-800',      // Dark mode background
  // ... other classes
)}
```

#### 2. Material-UI Components
Located in `src/@fuse/default-settings/DefaultSettings.ts`:

- **MuiOutlinedInput**: Controls TextField with variant="outlined"
```javascript
MuiOutlinedInput: {
  styleOverrides: {
    root: ({theme}) => ({
      backgroundColor: theme.palette.mode === 'dark' 
        ? 'rgba(255, 255, 255, 0.05)'  // Dark mode: subtle white overlay
        : '#f8f9fa',                    // Light mode: light gray
    }),
  }
}
```

- **MuiFilledInput**: Controls TextField with variant="filled"
```javascript
MuiFilledInput: {
  styleOverrides: {
    root: ({theme}) => ({
      backgroundColor: theme.palette.mode === 'dark'
        ? 'rgba(255, 255, 255, 0.05)'  // Dark mode
        : '#f3f4f6',                    // Light mode
    }),
  }
}
```

### Accessibility Requirements

Form fields must have sufficient contrast from the background:
- **WCAG AA Standard**: Minimum 3:1 contrast ratio for UI elements
- **Best Practice**: Use subtle background colors that distinguish fields from the page background
- **Testing**: Always verify form visibility in both light and dark modes

### Common Background Color Values

**Light Mode Options:**
- `#f8f9fa` - Very light gray (recommended for outlined inputs)
- `#f3f4f6` - Light gray (recommended for filled inputs)
- `gray-50` - Tailwind's lightest gray (#f9fafb)

**Dark Mode Options:**
- `rgba(255, 255, 255, 0.05)` - 5% white overlay (subtle)
- `#2a2d35` - Dark gray (for dropdowns/papers)
- `gray-800` - Tailwind's dark gray (#1f2937)

### Quick Reference for Future Adjustments

1. **To make form fields more visible in light mode**: Increase the gray tone (e.g., from `gray-50` to `gray-100`)
2. **To make form fields more visible in dark mode**: Increase the white overlay (e.g., from `0.05` to `0.08`)
3. **Always test changes** by toggling between light/dark modes in the UI

Follow this structure for consistent app development.
