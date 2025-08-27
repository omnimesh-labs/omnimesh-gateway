// Export all validation schemas and utilities
export * from './common';
export * from './organization';
export * from './server';
export * from './resource';
export * from './tool';
export * from './prompt';
export * from './user';
export * from './rateLimit';
export * from './session';

// Re-export commonly used validation functions
export {
  commonValidation,
  validationUtils,
  enums,
  type PlanType,
  type Protocol,
  type ServerStatus,
  type SessionStatus,
  type ProcessStatus,
  type UserRole,
  type ResourceType,
  type ToolCategory,
  type ToolImplementationType,
  type PromptCategory,
  type AdapterType,
  type LogLevel,
  type WindowType,
  type RateLimitScope,
  type AuditAction,
  type ImportStatus,
  type ConflictStrategy
} from './common';

// Validation schema collections for easy access
export const validationSchemas = {
  // Organization schemas
  organization: {
    create: () => import('./organization').then(m => m.createOrganizationSchema),
    update: () => import('./organization').then(m => m.updateOrganizationSchema),
    full: () => import('./organization').then(m => m.organizationSchemaWithPlanValidation),
  },

  // Server schemas
  server: {
    create: () => import('./server').then(m => m.createServerSchema),
    update: () => import('./server').then(m => m.updateServerSchema),
    full: () => import('./server').then(m => m.serverSchemaWithProtocolValidation),
  },

  // Resource schemas
  resource: {
    create: () => import('./resource').then(m => m.createResourceSchema),
    update: () => import('./resource').then(m => m.updateResourceSchema),
    full: () => import('./resource').then(m => m.resourceSchemaWithTypeValidation),
  },

  // Tool schemas
  tool: {
    create: () => import('./tool').then(m => m.createToolSchema),
    update: () => import('./tool').then(m => m.updateToolSchema),
    full: () => import('./tool').then(m => m.toolSchemaWithImplementationValidation),
  },

  // Prompt schemas
  prompt: {
    create: () => import('./prompt').then(m => m.createPromptSchema),
    update: () => import('./prompt').then(m => m.updatePromptSchema),
    full: () => import('./prompt').then(m => m.promptSchemaWithTemplateValidation),
  },

  // User schemas
  user: {
    create: () => import('./user').then(m => m.createUserSchema),
    update: () => import('./user').then(m => m.updateUserSchema),
    register: () => import('./user').then(m => m.registerUserSchema),
    login: () => import('./user').then(m => m.loginSchema),
    changePassword: () => import('./user').then(m => m.changePasswordSchema),
    profile: () => import('./user').then(m => m.updateProfileSchema),
  },

  // Rate limit schemas
  rateLimit: {
    create: () => import('./rateLimit').then(m => m.createRateLimitSchema),
    update: () => import('./rateLimit').then(m => m.updateRateLimitSchema),
    full: () => import('./rateLimit').then(m => m.rateLimitSchemaWithValidation),
  },

  // Session schemas
  session: {
    create: () => import('./session').then(m => m.createSessionSchema),
    update: () => import('./session').then(m => m.updateSessionSchema),
    full: () => import('./session').then(m => m.sessionSchemaWithLifecycleValidation),
  },
};

// Validation helpers collection
export const validationHelpers = {
  organization: {
    validatePlanLimits: () => import('./organization').then(m => m.validatePlanLimits),
    planLimits: () => import('./organization').then(m => m.planLimits),
  },

  server: {
    validateEnvironmentVariable: () => import('./server').then(m => m.validateEnvironmentVariable),
    validateCommandArgument: () => import('./server').then(m => m.validateCommandArgument),
  },

  resource: {
    validateAccessPermissions: () => import('./resource').then(m => m.validateAccessPermissions),
    validateUriForType: () => import('./resource').then(m => m.validateUriForType),
    suggestMimeType: () => import('./resource').then(m => m.suggestMimeType),
  },

  tool: {
    validateToolSchema: () => import('./tool').then(m => m.validateToolSchema),
    validateToolExamples: () => import('./tool').then(m => m.validateToolExamples),
    suggestToolCategory: () => import('./tool').then(m => m.suggestToolCategory),
    generateToolSchemaTemplate: () => import('./tool').then(m => m.generateToolSchemaTemplate),
  },

  prompt: {
    validatePromptParameters: () => import('./prompt').then(m => m.validatePromptParameters),
    validatePromptTemplate: () => import('./prompt').then(m => m.validatePromptTemplate),
    extractTemplateParameters: () => import('./prompt').then(m => m.extractTemplateParameters),
    suggestPromptCategory: () => import('./prompt').then(m => m.suggestPromptCategory),
  },

  user: {
    validatePasswordStrength: () => import('./user').then(m => m.validatePasswordStrength),
    validateEmailDomain: () => import('./user').then(m => m.validateEmailDomain),
    validateRoleAssignment: () => import('./user').then(m => m.validateRoleAssignment),
    rolePermissions: () => import('./user').then(m => m.rolePermissions),
  },

  rateLimit: {
    validateScopeId: () => import('./rateLimit').then(m => m.validateScopeId),
    calculateDerivedLimits: () => import('./rateLimit').then(m => m.calculateDerivedLimits),
    validateRateLimitEffectiveness: () => import('./rateLimit').then(m => m.validateRateLimitEffectiveness),
    scopeConfigurations: () => import('./rateLimit').then(m => m.scopeConfigurations),
  },

  session: {
    validateStatusTransition: () => import('./session').then(m => m.validateStatusTransition),
    validateSessionTimeout: () => import('./session').then(m => m.validateSessionTimeout),
    validateProcessInfo: () => import('./session').then(m => m.validateProcessInfo),
    checkSessionHealth: () => import('./session').then(m => m.checkSessionHealth),
  },
};
