# Frontend Testing Guide - Janex Gateway

## Overview

This guide provides comprehensive instructions for AI coding agents on working with the Janex Gateway frontend test suite. The project uses Jest with jsdom environment and follows a structured testing approach.

## Testing Architecture

### Framework & Tools
- **Test Runner**: Jest with jsdom environment 
- **React Testing**: React Testing Library (for component tests)
- **API Mocking**: Mock Service Worker (MSW) for integration tests
- **File Naming**: `*.simple.test.ts` for unit tests, `*.test.ts` for integration tests
- **Configuration**: Uses `jest.simple.config.cjs` for simple unit tests

### Test Categories

#### 1. Simple Unit Tests (`*.simple.test.ts`)
- **Purpose**: Fast, isolated tests with minimal dependencies
- **Mocking**: Uses Jest `mockFn()` to mock `fetch` globally
- **Environment**: jsdom with polyfills for browser APIs
- **Focus**: Pure functions, API clients, validation, utilities

#### 2. Integration Tests (`*.test.ts`) 
- **Purpose**: Tests with MSW server, React components
- **Mocking**: Uses Mock Service Worker for realistic API interactions
- **Environment**: Full browser simulation with MSW handlers
- **Focus**: Component behavior, user interactions, API integration

## Project Structure

### Test Directory Structure
```
src/test/
├── __mocks__/          # Mock files for modules
├── handlers/           # MSW request handlers
├── polyfills.ts        # Browser API polyfills
├── server.ts           # MSW server setup
├── setup-simple.ts     # Jest setup for simple tests
└── setup.ts           # Jest setup for integration tests
```

### Test File Locations
```
src/lib/__tests__/                    # API client tests
src/lib/validation/__tests__/         # Schema validation tests  
src/hooks/__tests__/                  # Custom hook tests
src/@auth/__tests__/                  # Auth context tests
src/app/(controlpanel)/*/api/hooks/__tests__/  # TanStack Query hooks
```

## Test Types & Coverage Areas

### API Client Testing
- **Authentication API** (`src/lib/auth-api.ts`)
  - File: `src/lib/__tests__/auth-api-extended.simple.test.ts`
  - Tests: Token management, auth flow, API key CRUD, error handling

- **Client API** (`src/lib/client-api.ts`)
  - File: `src/lib/__tests__/client-api.simple.test.ts`
  - Tests: All API classes, CRUD operations, HTTP status handling

- **Configuration API** (`src/lib/config-api.ts`)
  - File: `src/lib/__tests__/config-api.simple.test.ts` 
  - Tests: Export/import functionality, FormData handling, Blob responses

### Validation & Data Integrity
- **Validation Schemas** (`src/lib/validation/`)
  - Files: `src/lib/validation/__tests__/*.simple.test.ts`
  - Tests: Zod schema validation, constraint validation, enum validation

- **Utility Functions** (`src/lib/utils.ts`)
  - File: `src/lib/__tests__/utils.simple.test.ts`
  - Tests: `cn` function, className merging, Tailwind conflicts

### React Hooks & State Management
- **Custom Hooks** (`src/hooks/`)
  - Files: `src/hooks/__tests__/*.simple.test.ts`
  - Tests: Hook behavior, state updates, cleanup functions, edge cases

- **Authentication Context** (`src/@auth/AuthContext.tsx`)
  - File: `src/@auth/__tests__/AuthContext.simple.test.tsx`
  - Tests: Provider initialization, auth state, token refresh

### Data Fetching & API Integration
- **TanStack Query Hooks**
  - Files: `src/app/(controlpanel)/*/api/hooks/__tests__/*.simple.test.ts`
  - Tests: Query behavior, mutations, cache management, error states

## Test Implementation Guidelines

### File Structure Pattern
```typescript
import { apiClient } from '../api-module';

// Mock fetch globally for simple tests
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('API Module - Simple Tests', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    // Clear storage if needed
    if (typeof localStorage !== 'undefined') {
      localStorage.clear();
    }
  });

  describe('method group', () => {
    it('should do expected behavior', async () => {
      // Arrange: Setup mocks
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ data: 'expected' })
      });

      // Act: Call the function
      const result = await apiClient.someMethod();

      // Assert: Verify behavior
      expect(mockFetch).toHaveBeenCalledWith(/* expected args */);
      expect(result).toEqual(/* expected result */);
    });
  });
});
```

### Mock Patterns

#### 1. Fetch API Mocking (Simple Tests)
```typescript
// Success response
mockFetch.mockResolvedValueOnce({
  ok: true,
  status: 200,
  headers: { get: jest.fn().mockReturnValue('application/json') },
  json: jest.fn().mockResolvedValueOnce({ data: 'result' })
});

// Error response  
mockFetch.mockResolvedValueOnce({
  ok: false,
  status: 400,
  statusText: 'Bad Request'
});

// Network error
mockFetch.mockRejectedValueOnce(new Error('Network error'));
```

#### 2. LocalStorage Handling
```typescript
beforeEach(() => {
  // Clear storage safely
  if (typeof localStorage !== 'undefined') {
    localStorage.clear();
  }
  if (typeof sessionStorage !== 'undefined') {
    sessionStorage.clear();
  }
});
```

#### 3. Blob/FormData Handling
```typescript
// For file downloads (Blob)
mockFetch.mockResolvedValueOnce({
  ok: true,
  status: 200,
  blob: jest.fn().mockResolvedValueOnce(new Blob(['data']))
});

// For file uploads (FormData)  
const formData = new FormData();
formData.append('file', new Blob(['content']), 'file.json');
```

### Test Naming Conventions
- **Descriptive names**: `should return true when access token exists`
- **Group related tests**: Use `describe` blocks for logical grouping
- **Follow AAA pattern**: Arrange, Act, Assert
- **Error cases**: `should throw error when [condition]`
- **Edge cases**: `should handle [edge case] gracefully`

### Common Testing Patterns

#### API Client Testing
1. **Success cases**: Verify correct endpoint, method, headers, response parsing
2. **Error handling**: Test all HTTP status codes (401, 403, 404, 500)
3. **Query parameters**: Test URL building with various parameter combinations
4. **Request bodies**: Verify JSON serialization for POST/PUT requests
5. **Response parsing**: Test data extraction from response wrappers

#### Validation Testing  
1. **Valid inputs**: Test successful validation for correct data
2. **Invalid inputs**: Test error messages for constraint violations
3. **Edge cases**: Empty strings, null values, boundary conditions
4. **Schema constraints**: Required fields, string lengths, number ranges
5. **Enum validation**: Test allowed/disallowed enum values

#### Hook Testing
1. **Initial state**: Test hook initialization
2. **State updates**: Test state changes from hook actions
3. **Side effects**: Test useEffect behavior, cleanup functions
4. **Dependencies**: Test hook behavior with changing dependencies
5. **Error handling**: Test hook error states and recovery

## Critical Implementation Notes

### Query Parameter Issues
**Problem**: JavaScript falsy values (`false`, `0`, `""`) are excluded from API query strings due to truthy checks in client code.

**Example Issue**:
```typescript
// In client-api.ts - PROBLEMATIC
if (params.offset) query.set('offset', String(params.offset));
// When offset=0, this evaluates to false and param is omitted

// CORRECT approach for tests
if (params.offset !== undefined) query.set('offset', String(params.offset));
```

**Testing Strategy**: 
- Expect actual behavior (omitted falsy params) rather than ideal behavior
- Document the discrepancy for future API fixes
- Test both truthy and falsy parameter scenarios

### Error Handling Patterns
The codebase has specific error parsing logic:
```typescript
// Error parsing priority in apiRequest helper
1. error.message (from parsed JSON)  
2. message (from parsed JSON)
3. error (string from parsed JSON)
4. Generic status-based message fallback
```

Test error scenarios should match this parsing hierarchy.

### Environment Considerations
- **Browser APIs**: localStorage, sessionStorage are polyfilled in test environment
- **Fetch**: Globally mocked with Jest functions
- **Timers**: Use `jest.useFakeTimers()` for debounce/timeout testing
- **React components**: Use React Testing Library for component tests

## Running Tests

### Test Commands
```bash
# Run all simple unit tests
bun test

# Run with watch mode  
bun test:watch

# Run with coverage
bun test:coverage

# Run specific test file
bun test auth-api.simple.test.ts

# Run specific test pattern
bun test --testNamePattern="should handle error"
```

### Coverage Goals
- **Critical APIs**: 90%+ coverage
- **Validation**: 95%+ coverage  
- **Hooks**: 85%+ coverage
- **Utils**: 90%+ coverage

## Development Workflow

### Adding New Tests
1. **Create test file** in appropriate `__tests__/` directory
2. **Follow naming convention**: `*.simple.test.ts` for unit tests
3. **Implement test patterns** from this guide
4. **Verify all tests pass** with `bun test`

### Debugging Failed Tests
1. **Check mock setup**: Ensure fetch/localStorage mocks are correct
2. **Verify API behavior**: Use browser dev tools to see actual API calls
3. **Test isolation**: Ensure tests don't affect each other (clear mocks/storage)
4. **Error messages**: Check actual vs expected error message patterns

### Best Practices
- **Keep tests focused**: One behavior per test
- **Use descriptive names**: Test names should explain the scenario
- **Mock minimally**: Only mock external dependencies
- **Test edge cases**: Empty arrays, null values, error conditions
- **Maintain test independence**: Each test should run in isolation

## Test Coverage Areas

### Component Testing
- Form components with validation states
- Modal/dialog interactions
- Table components with sorting/filtering
- Navigation components

### Integration Testing
- End-to-end user workflows
- Multi-component interactions
- Authentication flows
- CRUD operation sequences

### Performance Testing
- Component render performance
- Hook performance under load
- Memory leak detection
- Bundle size impact analysis

---

**Note**: This guide focuses on architectural patterns and implementation guidelines for systematic test development.
