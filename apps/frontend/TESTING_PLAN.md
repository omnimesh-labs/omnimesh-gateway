# Frontend Unit Testing Plan - Omnimesh Gateway

## Overview

This document outlines the comprehensive unit testing strategy for the Omnimesh Gateway frontend application. The plan prioritizes critical functionality and provides a roadmap for achieving robust test coverage across the application.

## Testing Framework

- **Test Runner**: Jest with jsdom environment
- **React Testing**: React Testing Library
- **API Mocking**: Mock Service Worker (MSW) for integration tests
- **File Naming**: `*.simple.test.ts` for unit tests
- **Configuration**: Uses `jest.simple.config.cjs` for simple unit tests

## Priority Levels

### Priority 1: Critical Core Functionality (Implement First)

#### 1.1 Authentication API (`src/lib/auth-api.ts`) ✅ COMPLETED
**File**: `src/lib/__tests__/auth-api-extended.simple.test.ts`

**Test Coverage**:
- ✅ Token management (get, set, remove)
- ✅ Authentication state checking
- ✅ Login/logout flow
- ✅ Token refresh mechanism
- ✅ Error handling for different HTTP status codes
- ✅ API request helper function
- ✅ Edge cases (missing tokens, expired tokens)
- ✅ Profile management
- ✅ API key management (CRUD operations)
- ✅ HTTP status code handling (401, 403, 404, 500)
- ✅ 204 No Content response handling
- ✅ Content-Type parsing
- ✅ Error parsing and handling

**Status**: Implemented with comprehensive coverage (43 tests, 100% pass rate)

#### 1.2 Client API (`src/lib/client-api.ts`) ✅ COMPLETED
**File**: `src/lib/__tests__/client-api.simple.test.ts`

**Test Coverage**:
- ✅ API request helper function
- ✅ Error parsing and handling
- ✅ HTTP status code handling (401, 403, 404, 500)
- ✅ 204 No Content response handling
- ✅ Content-Type parsing
- ✅ AdminAPI, ServerAPI, NamespaceAPI, A2AApi, PolicyAPI, ContentFilterAPI, EndpointAPI, DiscoveryAPI, ToolsAPI, PromptsAPI, AuthConfigAPI classes
- ✅ CRUD operations for all entities (mocked)

**Status**: Implemented with comprehensive coverage (60+ tests, 95%+ pass rate)

#### 1.3 Configuration API (`src/lib/config-api.ts`) ✅ COMPLETED
**File**: `src/lib/__tests__/config-api.simple.test.ts`

**Test Coverage**:
- ✅ Configuration export functionality
- ✅ Configuration import functionality 
- ✅ Error handling (400, 401, 403, 500, network errors)
- ✅ Response parsing (Blob handling, FormData handling)
- ✅ API Base URL configuration
- ✅ Edge case handling

**Status**: Implemented with comprehensive coverage (29+ tests, 100% pass rate)

### Priority 2: Validation & Data Integrity ✅ COMPLETED

#### 2.1 Validation Schemas (`src/lib/validation/`) ✅ COMPLETED
**Files**:
- ✅ `src/lib/validation/__tests__/common.simple.test.ts`
- ✅ `src/lib/validation/__tests__/user.simple.test.ts`
- ✅ `src/lib/validation/__tests__/server.simple.test.ts`
- ✅ `src/lib/validation/__tests__/tool.simple.test.ts`

**Test Coverage**:
- ✅ Zod schema validation for all entities
- ✅ Schema constraint validation
- ✅ Error message accuracy
- ✅ Edge case validation
- ✅ Enum validations
- ✅ Complex validation rules (e.g., URL formats, email formats)
- ✅ Validation helper functions

**Status**: Implemented with comprehensive coverage (203+ tests, 95%+ pass rate)

#### 2.2 Utility Functions (`src/lib/utils.ts`) ✅ COMPLETED
**File**: ✅ `src/lib/__tests__/utils.simple.test.ts`

**Test Coverage**:
- ✅ `cn` function for className merging
- ✅ Various class combinations and edge cases
- ✅ Tailwind CSS class conflict resolution

**Status**: Implemented with comprehensive coverage (29 tests, 100% pass rate)

### Priority 3: React Hooks & State Management

#### 3.1 Custom Hooks (`src/hooks/`)
**Files**:
- `src/hooks/__tests__/useDebounce.simple.test.ts`
- `src/hooks/__tests__/useWindowSize.simple.test.ts`
- `src/hooks/__tests__/useMobile.simple.test.ts`
- `src/hooks/__tests__/useFormValidation.simple.test.ts`

**Test Coverage**:
- Hook behavior under different conditions
- State updates and side effects
- Cleanup functions
- Edge cases and error scenarios
- Performance characteristics

#### 3.2 Authentication Context (`src/@auth/AuthContext.tsx`)
**File**: `src/@auth/__tests__/AuthContext.simple.test.tsx`

**Test Coverage**:
- Provider initialization
- Authentication state management
- Login/logout functionality
- Token refresh handling
- Cache management (session storage)
- Error handling
- useAuth hook behavior
- Context error when used outside provider

### Priority 4: Data Fetching & API Integration

#### 4.1 TanStack Query Hooks
**Files**:
- `src/app/(controlpanel)/content/tools/api/hooks/__tests__/useTools.simple.test.ts`
- `src/app/(controlpanel)/content/prompts/api/hooks/__tests__/usePrompts.simple.test.ts`
- `src/app/(controlpanel)/content/resources/api/hooks/__tests__/useResources.simple.test.ts`

**Test Coverage**:
- Query key generation
- Query function behavior
- Mutation success/error handling
- Cache invalidation
- Optimistic updates
- Loading states
- Error states

## Test Implementation Guidelines

### File Structure
```
src/
├── lib/
│   ├── __tests__/
│   │   ├── auth-api.simple.test.ts (extend existing)
│   │   ├── client-api.simple.test.ts
│   │   └── config-api.simple.test.ts
│   └── validation/
│       └── __tests__/
│           ├── common.simple.test.ts
│           ├── user.simple.test.ts
│           └── server.simple.test.ts
├── hooks/
│   └── __tests__/
│       ├── useDebounce.simple.test.ts
│       └── useWindowSize.simple.test.ts
└── @auth/
    └── __tests__/
        └── AuthContext.simple.test.tsx
```

### Test Naming Conventions
- Use descriptive test names: `should return true when access token exists`
- Group related tests in `describe` blocks
- Use `beforeEach`/`afterEach` for setup/cleanup
- Follow AAA pattern: Arrange, Act, Assert

### Mock Strategy
- **Minimal Mocking**: Focus on testing pure functions and isolated units
- **API Calls**: Mock fetch or use MSW for integration tests
- **localStorage/sessionStorage**: Already mocked in test setup
- **Window objects**: Mock as needed for specific tests

### Code Coverage Goals
- **Critical APIs**: 90%+ coverage
- **Validation**: 95%+ coverage
- **Hooks**: 85%+ coverage
- **Utils**: 90%+ coverage

## Running Tests

```bash
# Run all simple unit tests
bun test

# Run with watch mode
bun test:watch

# Run with coverage
bun test:coverage

# Run specific test file
bun test auth-api.simple.test.ts
```

## Future Enhancements

### Phase 2: Component Testing
- Form components testing
- Modal/dialog components
- Table components
- Navigation components

### Phase 3: Integration Testing
- End-to-end user flows
- API integration tests with MSW
- Authentication flows
- CRUD operation workflows

### Phase 4: Performance Testing
- Component render performance
- Hook performance under load
- Memory leak detection
- Bundle size impact

## Test Data Management

### Mock Data Strategy
- Create reusable mock data factories
- Maintain consistent test data across tests
- Use realistic data that matches API contracts

### Test Database
- Not applicable for unit tests
- Integration tests may use in-memory databases

## CI/CD Integration

### Pre-commit Hooks
- Run tests before commits
- Ensure test coverage thresholds
- Lint test files

### Build Pipeline
- Run full test suite on PR
- Generate coverage reports
- Fail builds on coverage drops

## Success Metrics

1. **Coverage**: Achieve target coverage percentages
2. **Reliability**: Tests pass consistently
3. **Speed**: Test suite completes in under 30 seconds
4. **Maintainability**: Tests are easy to update when code changes
5. **Documentation**: Tests serve as living documentation

## Resources

- [Jest Documentation](https://jestjs.io/)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [MSW Documentation](https://mswjs.io/)
- [Zod Testing Patterns](https://zod.dev/)

---

**Note**: This plan focuses on unit tests that can be implemented independently by different developers. Each priority level can be assigned to separate team members or completed in phases.
