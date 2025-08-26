'use client';

import { createContext, useContext, useState, useEffect, ReactNode, useMemo, useCallback } from 'react';
import { authApi, User, LoginRequest } from '@/lib/api';

interface AuthContextType {
	user: User | null;
	isAuthenticated: boolean;
	isLoading: boolean;
	login: (credentials: LoginRequest) => Promise<void>;
	logout: () => Promise<void>;
	refreshToken: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
	const context = useContext(AuthContext);

	if (context === undefined) {
		throw new Error('useAuth must be used within an AuthProvider');
	}

	return context;
}

interface AuthProviderProps {
	children: ReactNode;
}

// Cache auth state in session storage to avoid redundant checks
const AUTH_CACHE_KEY = 'auth_cache';
const AUTH_CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

interface AuthCache {
	user: User | null;
	timestamp: number;
}

function getCachedAuth(): AuthCache | null {
	if (typeof window === 'undefined') return null;

	try {
		const cached = sessionStorage.getItem(AUTH_CACHE_KEY);

		if (!cached) return null;

		const data = JSON.parse(cached) as AuthCache;
		const now = Date.now();

		// Check if cache is expired
		if (now - data.timestamp > AUTH_CACHE_DURATION) {
			sessionStorage.removeItem(AUTH_CACHE_KEY);
			return null;
		}

		return data;
	} catch {
		return null;
	}
}

function setCachedAuth(user: User | null): void {
	if (typeof window === 'undefined') return;

	try {
		const cache: AuthCache = {
			user,
			timestamp: Date.now()
		};
		sessionStorage.setItem(AUTH_CACHE_KEY, JSON.stringify(cache));
	} catch {
		// Ignore cache errors
	}
}

function clearCachedAuth(): void {
	if (typeof window === 'undefined') return;

	sessionStorage.removeItem(AUTH_CACHE_KEY);
}

export function AuthProvider({ children }: AuthProviderProps) {
	const [user, setUser] = useState<User | null>(null);
	const [isLoading, setIsLoading] = useState(true);

	// Check if user is authenticated and load profile
	useEffect(() => {
		const initializeAuth = async () => {
			// First check cached auth
			const cached = getCachedAuth();

			if (cached?.user) {
				setUser(cached.user);
				setIsLoading(false);
				return;
			}

			// Only check auth if we have tokens
			if (!authApi.isAuthenticated()) {
				setIsLoading(false);
				return;
			}

			try {
				const userProfile = await authApi.getProfile();
				setUser(userProfile);
				setCachedAuth(userProfile);
			} catch (profileError) {
				// Only attempt refresh if it's a 401 error
				if (profileError?.message?.includes('401')) {
					try {
						const refreshResponse = await authApi.refresh();
						setUser(refreshResponse.user);
						setCachedAuth(refreshResponse.user);
					} catch (_refreshError) {
						// Clear everything on refresh failure
						authApi.clearTokens();
						clearCachedAuth();
						setUser(null);

						// Redirect to sign-in when token is expired
						if (typeof window !== 'undefined') {
							window.location.href = '/sign-in';
						}
					}
				} else {
					// Non-auth error, keep existing state
					console.error('Profile fetch error:', profileError);
				}
			} finally {
				setIsLoading(false);
			}
		};

		initializeAuth();
	}, []);

	const login = useCallback(async (credentials: LoginRequest) => {
		const response = await authApi.login(credentials);
		setUser(response.user);
		setCachedAuth(response.user);
	}, []);

	const logout = useCallback(async () => {
		try {
			await authApi.logout();
		} catch (error) {
			console.error('Logout error:', error);
		} finally {
			setUser(null);
			clearCachedAuth();

			// Redirect to sign-in page after logout
			if (typeof window !== 'undefined') {
				window.location.href = '/sign-in';
			}
		}
	}, []);

	const refreshToken = useCallback(async () => {
		try {
			const response = await authApi.refresh();
			setUser(response.user);
			setCachedAuth(response.user);
		} catch (error) {
			// If refresh fails, log the user out
			setUser(null);
			authApi.clearTokens();
			clearCachedAuth();

			// Redirect to sign-in when token refresh fails
			if (typeof window !== 'undefined') {
				window.location.href = '/sign-in';
			}

			throw error;
		}
	}, []);

	// Memoize the context value to prevent unnecessary re-renders
	const value = useMemo<AuthContextType>(
		() => ({
			user,
			isAuthenticated: !!user && authApi.isAuthenticated(),
			isLoading,
			login,
			logout,
			refreshToken
		}),
		[user, isLoading, login, logout, refreshToken]
	);

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
