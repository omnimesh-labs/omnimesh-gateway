import * as z from 'zod';
import { commonValidation, enums } from './common';

// MCP Session validation schema
export const sessionSchema = z.object({
  server_id: commonValidation.requiredUuid,
  status: z.enum(enums.sessionStatus, {
    errorMap: () => ({ message: 'Please select a valid session status' })
  }).default('initializing'),
  protocol: z.enum(enums.protocol, {
    errorMap: () => ({ message: 'Please select a valid protocol' })
  }),
  client_id: z.string()
    .max(255, 'Client ID must be less than 255 characters')
    .optional(),
  connection_id: z.string()
    .max(255, 'Connection ID must be less than 255 characters')
    .optional(),
  process_pid: z.number()
    .int('Process ID must be a whole number')
    .min(1, 'Process ID must be positive')
    .optional(),
  process_status: z.enum(enums.processStatus, {
    errorMap: () => ({ message: 'Please select a valid process status' })
  }).optional(),
  process_exit_code: z.number()
    .int('Exit code must be a whole number')
    .min(0, 'Exit code cannot be negative')
    .max(255, 'Exit code cannot exceed 255')
    .optional(),
  process_error: z.string()
    .max(1000, 'Process error must be less than 1000 characters')
    .optional(),
  user_id: z.string()
    .min(1, 'User ID is required')
    .max(255, 'User ID must be less than 255 characters')
    .default('default-user'),
  metadata: commonValidation.jsonString,
});

// Session lifecycle validation
export const sessionSchemaWithLifecycleValidation = sessionSchema
  .refine((data) => {
    // Validate status and process_status consistency
    if (data.status === 'active' && data.process_status === 'stopped') {
      return false;
    }
    if (data.status === 'closed' && data.process_status === 'running') {
      return false;
    }
    return true;
  }, (data) => ({
    message: `Session status "${data.status}" is inconsistent with process status "${data.process_status}"`,
    path: ['status'],
  }))
  .refine((data) => {
    // Validate that closed sessions have exit code or error
    if (data.status === 'closed' && !data.process_exit_code && !data.process_error) {
      return false;
    }
    return true;
  }, {
    message: 'Closed sessions must have either an exit code or error message',
    path: ['process_exit_code'],
  })
  .refine((data) => {
    // Validate that error status has error message
    if (data.status === 'error' && !data.process_error) {
      return false;
    }
    return true;
  }, {
    message: 'Error status requires a process error message',
    path: ['process_error'],
  });

// Session status transition validation
export const validateStatusTransition = (
  currentStatus: typeof enums.sessionStatus[number],
  newStatus: typeof enums.sessionStatus[number]
): { valid: boolean; error?: string } => {
  const validTransitions: Record<typeof enums.sessionStatus[number], typeof enums.sessionStatus[number][]> = {
    initializing: ['active', 'error', 'closed'],
    active: ['closed', 'error'],
    closed: [], // No transitions from closed
    error: ['closed'] // Can close errored sessions
  };

  const allowed = validTransitions[currentStatus];
  if (!allowed.includes(newStatus)) {
    return {
      valid: false,
      error: `Cannot transition from "${currentStatus}" to "${newStatus}". Valid transitions: ${allowed.join(', ')}`
    };
  }

  return { valid: true };
};

// Session timeout validation
export const validateSessionTimeout = (
  startedAt: Date,
  lastActivity: Date,
  maxIdleMinutes: number = 30
): { active: boolean; minutesIdle: number; warning?: string } => {
  const now = new Date();
  const minutesIdle = Math.floor((now.getTime() - lastActivity.getTime()) / (1000 * 60));
  const totalMinutes = Math.floor((now.getTime() - startedAt.getTime()) / (1000 * 60));

  const result = {
    active: minutesIdle < maxIdleMinutes,
    minutesIdle,
  };

  // Add warnings
  if (minutesIdle > maxIdleMinutes * 0.8) {
    return {
      ...result,
      warning: `Session will timeout in ${maxIdleMinutes - minutesIdle} minutes`
    };
  }

  if (totalMinutes > 24 * 60) { // 24 hours
    return {
      ...result,
      warning: 'Session has been running for over 24 hours'
    };
  }

  return result;
};

// Process validation helpers
export const validateProcessInfo = (
  pid?: number,
  status?: typeof enums.processStatus[number],
  exitCode?: number
): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  // If process ID is provided, validate it
  if (pid !== undefined) {
    if (pid <= 0) {
      errors.push('Process ID must be positive');
    }
    if (pid > 2147483647) { // Max 32-bit signed integer
      errors.push('Process ID exceeds maximum value');
    }
  }

  // Validate status and exit code consistency
  if (status === 'stopped' && exitCode === undefined) {
    errors.push('Stopped processes should have an exit code');
  }

  if (status === 'running' && exitCode !== undefined) {
    errors.push('Running processes should not have an exit code');
  }

  if (exitCode !== undefined && (exitCode < 0 || exitCode > 255)) {
    errors.push('Exit code must be between 0 and 255');
  }

  return {
    valid: errors.length === 0,
    errors
  };
};

// Connection validation
export const validateConnectionInfo = (
  protocol: typeof enums.protocol[number],
  connectionId?: string,
  clientId?: string
): { valid: boolean; errors: string[]; warnings: string[] } => {
  const errors: string[] = [];
  const warnings: string[] = [];

  // Protocol-specific validation
  switch (protocol) {
    case 'websocket':
    case 'sse':
      if (!connectionId) {
        warnings.push(`${protocol.toUpperCase()} sessions typically have connection IDs`);
      }
      break;
    case 'stdio':
      if (connectionId) {
        warnings.push('STDIO sessions typically do not have connection IDs');
      }
      break;
  }

  // Client ID validation
  if (clientId) {
    if (clientId.length < 3) {
      errors.push('Client ID must be at least 3 characters');
    }
    if (!/^[a-zA-Z0-9_-]+$/.test(clientId)) {
      errors.push('Client ID can only contain letters, numbers, hyphens, and underscores');
    }
  }

  // Connection ID validation
  if (connectionId) {
    if (connectionId.length < 8) {
      errors.push('Connection ID must be at least 8 characters');
    }
    if (!/^[a-zA-Z0-9_-]+$/.test(connectionId)) {
      errors.push('Connection ID can only contain letters, numbers, hyphens, and underscores');
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings
  };
};

// Generate session metadata template
export const generateSessionMetadata = (protocol: typeof enums.protocol[number]): string => {
  const templates = {
    http: {
      user_agent: 'MCP-Gateway/1.0',
      keep_alive: true,
      timeout: 30000
    },
    https: {
      user_agent: 'MCP-Gateway/1.0',
      keep_alive: true,
      timeout: 30000,
      tls_version: 'TLSv1.3'
    },
    websocket: {
      subprotocol: 'mcp',
      ping_interval: 30,
      max_message_size: 65536
    },
    sse: {
      retry_interval: 3000,
      keep_alive: true,
      last_event_id: null
    },
    stdio: {
      buffer_size: 8192,
      encoding: 'utf8',
      shell: false
    }
  };

  const template = templates[protocol] || {};
  return JSON.stringify(template, null, 2);
};

// Session health check
export const checkSessionHealth = (session: {
  status: typeof enums.sessionStatus[number];
  last_activity: Date;
  started_at: Date;
  process_status?: typeof enums.processStatus[number];
  process_error?: string;
}): {
  healthy: boolean;
  issues: string[];
  recommendations: string[];
} => {
  const issues: string[] = [];
  const recommendations: string[] = [];

  // Check status
  if (session.status === 'error') {
    issues.push('Session is in error state');
    if (session.process_error) {
      issues.push(`Process error: ${session.process_error}`);
    }
    recommendations.push('Review error logs and restart session if needed');
  }

  // Check activity
  const lastActivityMinutes = Math.floor((Date.now() - session.last_activity.getTime()) / (1000 * 60));
  if (lastActivityMinutes > 60) {
    issues.push(`No activity for ${lastActivityMinutes} minutes`);
    recommendations.push('Consider closing idle session to free resources');
  }

  // Check runtime
  const runtimeHours = Math.floor((Date.now() - session.started_at.getTime()) / (1000 * 60 * 60));
  if (runtimeHours > 24) {
    issues.push(`Session running for ${runtimeHours} hours`);
    recommendations.push('Long-running sessions may need periodic restarts');
  }

  // Check process status
  if (session.process_status === 'stopped' && session.status === 'active') {
    issues.push('Session marked active but process is stopped');
    recommendations.push('Update session status or restart process');
  }

  return {
    healthy: issues.length === 0,
    issues,
    recommendations
  };
};

// Session creation schema
export const createSessionSchema = sessionSchemaWithLifecycleValidation.omit({});

// Session update schema
export const updateSessionSchema = sessionSchemaWithLifecycleValidation.partial();

export type SessionFormData = z.infer<typeof sessionSchema>;
export type CreateSessionData = z.infer<typeof createSessionSchema>;
export type UpdateSessionData = z.infer<typeof updateSessionSchema>;
