/**
 * Native JavaScript replacements for lodash functions
 */

/**
 * Deep merge objects
 */
export function merge<T extends Record<string, unknown>>(target: T, ...sources: Partial<T>[]): T {
	if (!sources.length) return target;

	const source = sources.shift();

	if (isObject(target) && isObject(source)) {
		for (const key in source) {
			if (isObject(source[key])) {
				if (!target[key]) Object.assign(target, { [key]: {} });

				merge(target[key] as Record<string, unknown>, source[key] as Record<string, unknown>);
			} else {
				Object.assign(target, { [key]: source[key] });
			}
		}
	}

	return merge(target, ...sources);
}

/**
 * Deep equality check
 */
export function isEqual(a: unknown, b: unknown): boolean {
	if (a === b) return true;

	if (a instanceof Date && b instanceof Date) {
		return a.getTime() === b.getTime();
	}

	if (!a || !b || (typeof a !== 'object' && typeof b !== 'object')) {
		return a === b;
	}

	if (a === null || a === undefined || b === null || b === undefined) {
		return false;
	}

	if (a.prototype !== b.prototype) return false;

	const keys = Object.keys(a);

	if (keys.length !== Object.keys(b).length) {
		return false;
	}

	return keys.every((k) => isEqual(a[k], b[k]));
}

/**
 * Check if value is empty
 */
export function isEmpty(value: unknown): boolean {
	if (value == null) return true;

	if (typeof value === 'string' || Array.isArray(value)) {
		return value.length === 0;
	}

	if (value instanceof Map || value instanceof Set) {
		return value.size === 0;
	}

	if (typeof value === 'object') {
		return Object.keys(value).length === 0;
	}

	return false;
}

/**
 * Check if value is an object
 */
export function isObject(value: unknown): value is Record<string, unknown> {
	return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/**
 * Set value at path in object
 */
export function set<T extends Record<string, unknown>>(obj: T, path: string | string[], value: unknown): T {
	const pathArray = Array.isArray(path) ? path : path.split('.');
	const newObj = { ...obj };

	let current = newObj;
	for (let i = 0; i < pathArray.length - 1; i++) {
		const key = pathArray[i];

		if (!current[key] || typeof current[key] !== 'object') {
			current[key] = {};
		}

		current = current[key];
	}

	current[pathArray[pathArray.length - 1]] = value;
	return newObj;
}

/**
 * Set default values
 */
export function defaults<T extends Record<string, unknown>>(obj: T, ...defaultObjs: Partial<T>[]): T {
	const result = { ...obj };

	for (const defaultObj of defaultObjs) {
		for (const key in defaultObj) {
			if (result[key] === undefined) {
				result[key] = defaultObj[key];
			}
		}
	}

	return result;
}

/**
 * Find item in array
 */
export function find<T>(array: T[], predicate: (item: T) => boolean): T | undefined {
	return array.find(predicate);
}

/**
 * Check if any item matches predicate
 */
export function some<T>(array: T[], predicate: (item: T) => boolean): boolean {
	return array.some(predicate);
}

/**
 * Transform object
 */
export function transform<T, R>(object: T, iteratee: (acc: R, value: unknown, key: string) => R, accumulator?: R): R {
	const isArr = Array.isArray(object);
	const initial = accumulator !== undefined ? accumulator : ((isArr ? [] : {}) as R);

	if (isArr) {
		return (object as unknown[]).reduce((acc, val, idx) => {
			iteratee(acc, val, String(idx));
			return acc;
		}, initial);
	}

	return Object.keys(object as Record<string, unknown>).reduce((acc, key) => {
		iteratee(acc, (object as Record<string, unknown>)[key], key);
		return acc;
	}, initial);
}

/**
 * XOR - symmetric difference of arrays
 */
export function xor<T>(...arrays: T[][]): T[] {
	const result: T[] = [];
	const map = new Map<T, number>();

	for (const array of arrays) {
		for (const item of array) {
			const count = map.get(item) || 0;
			map.set(item, count + 1);
		}
	}

	for (const [item, count] of map.entries()) {
		if (count === 1) {
			result.push(item);
		}
	}

	return result;
}

/**
 * Debounce function with cancel method
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
	func: T,
	wait: number
): ((...args: Parameters<T>) => void) & { cancel: () => void } {
	let timeout: NodeJS.Timeout | null = null;

	const debounced = function (this: unknown, ...args: Parameters<T>) {
		// eslint-disable-next-line @typescript-eslint/no-this-alias
		const context = this;

		if (timeout) clearTimeout(timeout);

		timeout = setTimeout(() => {
			func.apply(context, args);
		}, wait);
	};

	debounced.cancel = () => {
		if (timeout) {
			clearTimeout(timeout);
			timeout = null;
		}
	};

	return debounced as ((...args: Parameters<T>) => void) & { cancel: () => void };
}

/**
 * Deep clone using structured clone (modern browsers)
 */
export function cloneDeep<T>(obj: T): T {
	// Use structured clone if available (modern browsers/Node 17+)
	if (typeof structuredClone === 'function') {
		try {
			return structuredClone(obj);
		} catch {
			// Fall back to JSON method if structured clone fails
		}
	}

	// Fallback for simple objects
	if (obj === null || typeof obj !== 'object') return obj;

	if (obj instanceof Date) return new Date(obj.getTime()) as T;

	if (obj instanceof Array) return obj.map((item) => cloneDeep(item)) as T;

	if (obj instanceof Set) return new Set(Array.from(obj).map((item) => cloneDeep(item))) as T;

	if (obj instanceof Map)
		return new Map(Array.from(obj.entries()).map(([k, v]) => [cloneDeep(k), cloneDeep(v)])) as T;

	// Handle regular objects
	const clonedObj = {} as T;
	for (const key in obj) {
		if (Object.prototype.hasOwnProperty.call(obj, key)) {
			clonedObj[key] = cloneDeep(obj[key]);
		}
	}

	return clonedObj;
}

/**
 * Remove diacritical marks from string
 */
export function deburr(str: string): string {
	return str.normalize('NFD').replace(/[\u0300-\u036f]/g, '');
}
