'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
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

export function AuthProvider({ children }: AuthProviderProps) {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Check if user is authenticated and load profile
    useEffect(() => {
        const initializeAuth = async () => {
            try {
                if (authApi.isAuthenticated()) {
                    try {
                        const userProfile = await authApi.getProfile();
                        setUser(userProfile);
                    } catch (profileError) {
                        // If profile fails, try to refresh token
                        console.log('Profile fetch failed, attempting token refresh...');
                        try {
                            const refreshResponse = await authApi.refresh();
                            setUser(refreshResponse.user);
                            console.log('Token refreshed successfully');
                        } catch (refreshError) {
                            // If refresh also fails, clear tokens and log out
                            authApi.clearTokens();
                            console.error('Token refresh failed:', refreshError);
                        }
                    }
                }
            } catch (error) {
                // If something goes wrong, clear tokens
                authApi.clearTokens();
                console.error('Auth initialization failed:', error);
            } finally {
                setIsLoading(false);
            }
        };

        initializeAuth();
    }, []);

    const login = async (credentials: LoginRequest) => {
        try {
            const response = await authApi.login(credentials);
            // Ensure tokens are properly stored before setting user state
            setUser(response.user);
        } catch (error) {
            throw error;
        }
    };

    const logout = async () => {
        try {
            await authApi.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            setUser(null);
        }
    };

    const refreshToken = async () => {
        try {
            const response = await authApi.refresh();
            setUser(response.user);
        } catch (error) {
            // If refresh fails, log the user out
            setUser(null);
            authApi.clearTokens();
            throw error;
        }
    };

    const value: AuthContextType = {
        user,
        isAuthenticated: !!user && authApi.isAuthenticated(),
        isLoading,
        login,
        logout,
        refreshToken,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
