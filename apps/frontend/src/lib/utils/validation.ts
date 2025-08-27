import { z } from 'zod';

// Validation utility functions
export const validationUtils = {
  // Parse JSON with detailed error information
  parseJson: <T>(jsonString: string): { success: true; data: T } | { success: false; error: string; line?: number } => {
    try {
      const parsed = JSON.parse(jsonString);
      return { success: true, data: parsed };
    } catch (error) {
      const err = error as Error;
      let line: number | undefined;

      // Try to extract line number from error message
      const lineMatch = err.message.match(/at position (\d+)/);
      if (lineMatch) {
        const position = parseInt(lineMatch[1]);
        const lines = jsonString.substring(0, position).split('\n');
        line = lines.length;
      }

      return {
        success: false,
        error: `Invalid JSON: ${err.message}`,
        line
      };
    }
  },

  // Validate and format JSON with indentation
  formatJson: (jsonString: string, indent: number = 2): { success: true; formatted: string } | { success: false; error: string } => {
    const parseResult = validationUtils.parseJson(jsonString);
    if (!parseResult.success) {
      return parseResult;
    }

    try {
      const formatted = JSON.stringify(parseResult.data, null, indent);
      return { success: true, formatted };
    } catch (error) {
      return { success: false, error: 'Failed to format JSON' };
    }
  },

  // Validate URL with additional checks
  validateUrl: (url: string, options?: {
    allowHttp?: boolean;
    allowLocalhost?: boolean;
    allowIp?: boolean;
  }): { valid: boolean; error?: string; suggestions?: string[] } => {
    const opts = {
      allowHttp: false,
      allowLocalhost: true,
      allowIp: false,
      ...options
    };

    try {
      const parsedUrl = new URL(url);
      const suggestions: string[] = [];

      // Protocol checks
      if (parsedUrl.protocol !== 'https:' && parsedUrl.protocol !== 'http:') {
        return { valid: false, error: 'URL must use HTTP or HTTPS protocol' };
      }

      if (parsedUrl.protocol === 'http:' && !opts.allowHttp) {
        if (parsedUrl.hostname === 'localhost' || parsedUrl.hostname.startsWith('127.') || parsedUrl.hostname.startsWith('192.168.')) {
          // Allow HTTP for local development
        } else {
          suggestions.push('Consider using HTTPS for security');
          return { valid: false, error: 'HTTP URLs are not allowed in production', suggestions };
        }
      }

      // Hostname checks
      if (!opts.allowLocalhost && (parsedUrl.hostname === 'localhost' || parsedUrl.hostname.startsWith('127.'))) {
        return { valid: false, error: 'Localhost URLs are not allowed' };
      }

      // IP address check
      const ipPattern = /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/;
      if (!opts.allowIp && ipPattern.test(parsedUrl.hostname)) {
        return { valid: false, error: 'IP addresses are not allowed, use domain names' };
      }

      // Port checks
      if (parsedUrl.port && parseInt(parsedUrl.port) > 65535) {
        return { valid: false, error: 'Invalid port number' };
      }

      return { valid: true, suggestions: suggestions.length > 0 ? suggestions : undefined };
    } catch (error) {
      return { valid: false, error: 'Invalid URL format' };
    }
  },

  // Validate file path
  validateFilePath: (path: string, options?: {
    allowAbsolute?: boolean;
    allowRelative?: boolean;
    allowParentDir?: boolean;
  }): { valid: boolean; error?: string } => {
    const opts = {
      allowAbsolute: true,
      allowRelative: true,
      allowParentDir: false,
      ...options
    };

    if (!path || path.trim() === '') {
      return { valid: false, error: 'File path cannot be empty' };
    }

    const trimmedPath = path.trim();

    // Check for parent directory references
    if (!opts.allowParentDir && trimmedPath.includes('..')) {
      return { valid: false, error: 'Path cannot contain parent directory references (..)' };
    }

    // Check absolute vs relative paths
    const isAbsolute = trimmedPath.startsWith('/') || /^[a-zA-Z]:/.test(trimmedPath);

    if (isAbsolute && !opts.allowAbsolute) {
      return { valid: false, error: 'Absolute paths are not allowed' };
    }

    if (!isAbsolute && !opts.allowRelative) {
      return { valid: false, error: 'Relative paths are not allowed' };
    }

    // Check for invalid characters (basic check for cross-platform compatibility)
    const invalidChars = /[<>:"|?*]/;
    if (invalidChars.test(trimmedPath)) {
      return { valid: false, error: 'Path contains invalid characters' };
    }

    return { valid: true };
  },

  // Validate email with domain suggestions
  validateEmail: (email: string): { valid: boolean; error?: string; suggestion?: string } => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    if (!emailRegex.test(email)) {
      return { valid: false, error: 'Invalid email format' };
    }

    const [, domain] = email.split('@');

    // Common domain typos
    const domainSuggestions: Record<string, string> = {
      'gmail.co': 'gmail.com',
      'gmail.cm': 'gmail.com',
      'gmial.com': 'gmail.com',
      'gmai.com': 'gmail.com',
      'yahooo.com': 'yahoo.com',
      'yaho.com': 'yahoo.com',
      'hotmial.com': 'hotmail.com',
      'hotmai.com': 'hotmail.com',
      'outlok.com': 'outlook.com',
      'outloo.com': 'outlook.com',
    };

    const suggestion = domainSuggestions[domain.toLowerCase()];
    if (suggestion) {
      return {
        valid: false,
        error: `Did you mean ${email.replace(domain, suggestion)}?`,
        suggestion: email.replace(domain, suggestion)
      };
    }

    return { valid: true };
  },

  // Validate password strength
  validatePassword: (password: string): {
    score: number;
    strength: 'weak' | 'fair' | 'good' | 'strong';
    feedback: string[];
    isValid: boolean;
  } => {
    let score = 0;
    const feedback: string[] = [];

    // Length checks
    if (password.length >= 8) score += 1;
    else feedback.push('Use at least 8 characters');

    if (password.length >= 12) score += 1;
    else if (password.length >= 8) feedback.push('Consider using 12+ characters');

    // Character variety
    if (/[a-z]/.test(password)) score += 1;
    else feedback.push('Add lowercase letters');

    if (/[A-Z]/.test(password)) score += 1;
    else feedback.push('Add uppercase letters');

    if (/[0-9]/.test(password)) score += 1;
    else feedback.push('Add numbers');

    if (/[^A-Za-z0-9]/.test(password)) score += 1;
    else feedback.push('Add special characters');

    // Bonus points
    if (password.length >= 16) score += 1;
    if (/[^A-Za-z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) score += 1;

    // Penalties
    if (/(.)\1{2,}/.test(password)) {
      score -= 1;
      feedback.push('Avoid repeating characters');
    }

    if (/123456|abcdef|qwerty|password|admin|login/i.test(password)) {
      score -= 2;
      feedback.push('Avoid common patterns');
    }

    score = Math.max(0, score);

    const strength = score <= 2 ? 'weak' : score <= 4 ? 'fair' : score <= 6 ? 'good' : 'strong';
    const isValid = score >= 4; // Require at least 'fair' strength

    return {
      score,
      strength,
      feedback,
      isValid
    };
  },

  // Extract validation errors from Zod error
  extractZodErrors: (error: z.ZodError): Record<string, string> => {
    const errors: Record<string, string> = {};

    error.errors.forEach((err) => {
      const path = err.path.join('.');
      errors[path] = err.message;
    });

    return errors;
  },

  // Flatten nested validation errors
  flattenErrors: (errors: Record<string, any>): Record<string, string> => {
    const flattened: Record<string, string> = {};

    const flatten = (obj: any, prefix = '') => {
      Object.keys(obj).forEach(key => {
        const fullKey = prefix ? `${prefix}.${key}` : key;

        if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
          flatten(obj[key], fullKey);
        } else {
          flattened[fullKey] = String(obj[key]);
        }
      });
    };

    flatten(errors);
    return flattened;
  },

  // Debounce validation function
  debounce: <T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): ((...args: Parameters<T>) => void) => {
    let timeout: NodeJS.Timeout;

    return (...args: Parameters<T>) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => func(...args), wait);
    };
  },

  // Sanitize string input
  sanitizeString: (input: string, options?: {
    maxLength?: number;
    allowHtml?: boolean;
    allowNewlines?: boolean;
  }): string => {
    const opts = {
      maxLength: 1000,
      allowHtml: false,
      allowNewlines: true,
      ...options
    };

    let sanitized = input.trim();

    // Remove HTML tags if not allowed
    if (!opts.allowHtml) {
      sanitized = sanitized.replace(/<[^>]*>/g, '');
    }

    // Remove newlines if not allowed
    if (!opts.allowNewlines) {
      sanitized = sanitized.replace(/\n\r/g, ' ').replace(/\s+/g, ' ');
    }

    // Truncate if too long
    if (sanitized.length > opts.maxLength) {
      sanitized = sanitized.substring(0, opts.maxLength).trim();
    }

    return sanitized;
  },

  // Format file size
  formatFileSize: (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

  // Validate file size
  validateFileSize: (bytes: number, maxSizeBytes: number): { valid: boolean; error?: string } => {
    if (bytes > maxSizeBytes) {
      return {
        valid: false,
        error: `File size (${validationUtils.formatFileSize(bytes)}) exceeds maximum allowed (${validationUtils.formatFileSize(maxSizeBytes)})`
      };
    }
    return { valid: true };
  },

  // Generate slug from string
  generateSlug: (input: string): string => {
    return input
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9\s-]/g, '') // Remove special characters
      .replace(/\s+/g, '-') // Replace spaces with hyphens
      .replace(/-+/g, '-') // Replace multiple hyphens with single
      .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
  },

  // Validate cron expression (basic)
  validateCron: (expression: string): { valid: boolean; error?: string; nextRun?: Date } => {
    const cronParts = expression.trim().split(/\s+/);

    if (cronParts.length !== 5 && cronParts.length !== 6) {
      return { valid: false, error: 'Cron expression must have 5 or 6 parts' };
    }

    // Basic validation for common patterns
    const cronRegex = /^(\*|[0-9]+(,[0-9]+)*|[0-9]+-[0-9]+|\*\/[0-9]+)$/;

    for (let i = 0; i < cronParts.length; i++) {
      if (!cronRegex.test(cronParts[i])) {
        return { valid: false, error: `Invalid cron expression part: ${cronParts[i]}` };
      }
    }

    return { valid: true };
  },
};
