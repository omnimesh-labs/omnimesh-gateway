import { User } from '@auth/user';
import UserModel from '@auth/user/models/UserModel';
import { PartialDeep } from 'type-fest';
import api from '@/utils/api';

/**
 * Get user by id
 */
export async function authGetDbUser(userId: string): Promise<Response> {
	return api.get(`auth/users/${userId}`);
}

/**
 * Get user by email
 */
export async function authGetDbUserByEmail(email: string): Promise<Response> {
	const params = new URLSearchParams({ email });
	return api.get(`auth/users?${params.toString()}`);
}

/**
 * Update user
 */
export function authUpdateDbUser(user: PartialDeep<User>) {
	return api.put(`auth/users/${user.id}`, {
		body: JSON.stringify(UserModel(user))
	});
}

/**
 * Create user
 */
export async function authCreateDbUser(user: PartialDeep<User>) {
	return api.post('auth/users', {
		body: JSON.stringify(UserModel(user))
	});
}
