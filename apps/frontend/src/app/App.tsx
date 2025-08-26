'use client';

import { SnackbarProvider } from 'notistack';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { enUS } from 'date-fns/locale/en-US';
import { lazy, Suspense } from 'react';
import ErrorBoundary from '@fuse/utils/ErrorBoundary';
import { SettingsProvider } from '@fuse/core/Settings/SettingsProvider';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import MainThemeProvider from '../contexts/MainThemeProvider';
import AppContext from '@/contexts/AppContext';
import RootThemeProvider from '@/contexts/RootThemeProvider';
import UIProviders from '@/contexts/UIProviders';
import { AuthProvider } from '@auth/AuthContext';
import FontOptimization from '@/components/FontOptimization';

// Lazy load React Query Devtools only in development
const ReactQueryDevtools = lazy(() =>
	process.env.NODE_ENV === 'development'
		? import('@tanstack/react-query-devtools').then((mod) => ({ default: mod.ReactQueryDevtools }))
		: Promise.resolve({ default: () => null })
);

// Create query client outside component to prevent recreation
const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 5 * 60 * 1000, // 5 minutes
			gcTime: 10 * 60 * 1000, // 10 minutes (garbage collection)
			retry: 1,
			refetchOnWindowFocus: false, // Reduce unnecessary re-fetching
			refetchOnMount: false
		}
	}
});

type AppProps = {
	children?: React.ReactNode;
};

/**
 * The main App component.
 */
function App(props: AppProps) {
	const { children } = props;
	const AppContextValue = {};

	return (
		<ErrorBoundary>
			<AppContext value={AppContextValue}>
				<FontOptimization />
				{/* Date Picker Localization Provider */}
				<LocalizationProvider
					dateAdapter={AdapterDateFns}
					adapterLocale={enUS}
				>
					<QueryClientProvider client={queryClient}>
						<AuthProvider>
							<SettingsProvider>
								{/* Theme Provider */}
								<RootThemeProvider>
									<MainThemeProvider>
										<UIProviders>
											{/* Notistack Notification Provider */}
											<SnackbarProvider
												maxSnack={5}
												anchorOrigin={{
													vertical: 'bottom',
													horizontal: 'right'
												}}
												classes={{
													containerRoot: 'bottom-0 right-0 mb-13 md:mb-17 mr-2 lg:mr-20 z-99'
												}}
											>
												{children}
											</SnackbarProvider>
										</UIProviders>
									</MainThemeProvider>
								</RootThemeProvider>
							</SettingsProvider>
						</AuthProvider>
						{process.env.NODE_ENV === 'development' && (
							<Suspense fallback={null}>
								<ReactQueryDevtools initialIsOpen={false} />
							</Suspense>
						)}
					</QueryClientProvider>
				</LocalizationProvider>
			</AppContext>
		</ErrorBoundary>
	);
}

export default App;
