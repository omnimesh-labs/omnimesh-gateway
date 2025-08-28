import { lazy } from 'react';

// Lazy load heavy MUI components that aren't needed on initial render
export const LazyDataTable = lazy(() => import('@/components/data-table/DataTable'));
export const LazyConfigurationView = lazy(() => import('@/app/(controlpanel)/configuration/ConfigurationView'));
export const LazyA2AView = lazy(() => import('@/app/(controlpanel)/a2a/A2AView'));
export const LazyServersView = lazy(() => import('@/app/(controlpanel)/servers/ServersView'));
export const LazyNamespacesView = lazy(() => import('@/app/(controlpanel)/namespaces/NamespacesView'));
export const LazyEndpointsView = lazy(() => import('@/app/(controlpanel)/endpoints/EndpointsView'));
export const LazyContentView = lazy(() => import('@/app/(controlpanel)/content/ContentView'));
export const LazyLogsView = lazy(() => import('@/app/(controlpanel)/logs/LogsView'));

// Lazy load MUI Lab components
export const LazyDateTimePicker = lazy(() =>
	import('@mui/x-date-pickers/DateTimePicker').then((mod) => ({ default: mod.DateTimePicker }))
);

export const LazyAutocomplete = lazy(() =>
	import('@mui/material/Autocomplete').then((mod) => ({ default: mod.default }))
);
