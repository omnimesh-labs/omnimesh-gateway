import * as z from 'zod';
import { commonValidation, enums, validationUtils } from './common';

// MCP Prompt validation schema
export const promptSchema = z.object({
  name: commonValidation.name,
  description: commonValidation.description,
  prompt_template: z.string()
    .min(1, 'Prompt template is required')
    .max(10000, 'Prompt template must be less than 10,000 characters'),
  category: z.enum(enums.promptCategory, {
    errorMap: () => ({ message: 'Please select a valid prompt category' })
  }),
  parameters: commonValidation.jsonString,
  usage_count: z.number().int().min(0).default(0),
  is_active: commonValidation.isActive,
  tags: commonValidation.tags,
  metadata: commonValidation.jsonString,
});

// Enhanced prompt schema with template and parameter validation
export const promptSchemaWithTemplateValidation = promptSchema
  .refine((data) => {
    // Validate prompt template for parameter consistency
    const templateParams = extractTemplateParameters(data.prompt_template);

    if (data.parameters && data.parameters.trim() !== '') {
      const parseResult = validationUtils.parseJson(data.parameters);
      if (!parseResult.success) return false;

      const definedParams = parseResult.data;
      if (!Array.isArray(definedParams)) return false;

      const definedParamNames = definedParams.map((p: any) => p.name);

      // Check if all template parameters are defined
      for (const templateParam of templateParams) {
        if (!definedParamNames.includes(templateParam)) {
          return false;
        }
      }
    }

    return true;
  }, (data) => {
    const templateParams = extractTemplateParameters(data.prompt_template);
    const undefinedParams = templateParams;

    if (data.parameters && data.parameters.trim() !== '') {
      const parseResult = validationUtils.parseJson(data.parameters);
      if (parseResult.success && Array.isArray(parseResult.data)) {
        const definedParamNames = parseResult.data.map((p: any) => p.name);
        undefinedParams.splice(0, undefinedParams.length,
          ...templateParams.filter(p => !definedParamNames.includes(p)));
      }
    }

    return {
      message: `Template parameters not defined in parameters: ${undefinedParams.join(', ')}`,
      path: ['parameters'],
    };
  });

// Extract template parameters from prompt template
export const extractTemplateParameters = (template: string): string[] => {
  const paramRegex = /\{\{(\w+)\}\}/g;
  const params: string[] = [];
  let match;

  while ((match = paramRegex.exec(template)) !== null) {
    const paramName = match[1];
    if (!params.includes(paramName)) {
      params.push(paramName);
    }
  }

  return params;
};

// Validate prompt parameters
export const validatePromptParameters = (parametersString: string): {
  valid: boolean;
  error?: string;
  suggestions?: string[];
  parameters?: Array<{ name: string; type?: string; description?: string; required?: boolean }>;
} => {
  if (!parametersString || parametersString.trim() === '') {
    return { valid: true, parameters: [] };
  }

  const parseResult = validationUtils.parseJson(parametersString);
  if (!parseResult.success) {
    return {
      valid: false,
      error: parseResult.error,
      suggestions: ['Parameters must be valid JSON array format']
    };
  }

  const parameters = parseResult.data;

  if (!Array.isArray(parameters)) {
    return {
      valid: false,
      error: 'Parameters must be an array',
      suggestions: ['Wrap parameters in square brackets: [{"name": "param1", "type": "string"}]']
    };
  }

  // Validate each parameter
  const validatedParams = [];
  for (let i = 0; i < parameters.length; i++) {
    const param = parameters[i];

    if (typeof param !== 'object' || param === null) {
      return {
        valid: false,
        error: `Parameter ${i + 1} must be an object`,
        suggestions: ['Each parameter should be an object with "name" property']
      };
    }

    if (!param.name || typeof param.name !== 'string') {
      return {
        valid: false,
        error: `Parameter ${i + 1} must have a "name" property`,
        suggestions: ['Add "name" property with parameter name']
      };
    }

    // Validate parameter name format
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(param.name)) {
      return {
        valid: false,
        error: `Parameter "${param.name}" must be a valid identifier (letters, numbers, underscores only)`,
        suggestions: ['Parameter names should follow variable naming conventions']
      };
    }

    // Validate parameter type if provided
    const validTypes = ['string', 'number', 'boolean', 'array', 'object'];
    if (param.type && !validTypes.includes(param.type)) {
      return {
        valid: false,
        error: `Parameter "${param.name}" has invalid type "${param.type}"`,
        suggestions: [`Valid types are: ${validTypes.join(', ')}`]
      };
    }

    validatedParams.push({
      name: param.name,
      type: param.type || 'string',
      description: param.description || '',
      required: param.required !== false, // Default to true
    });
  }

  // Check for duplicate parameter names
  const paramNames = validatedParams.map(p => p.name);
  const duplicates = paramNames.filter((name, index) => paramNames.indexOf(name) !== index);
  if (duplicates.length > 0) {
    return {
      valid: false,
      error: `Duplicate parameter names: ${duplicates.join(', ')}`,
      suggestions: ['Each parameter must have a unique name']
    };
  }

  return {
    valid: true,
    parameters: validatedParams
  };
};

// Validate prompt template
export const validatePromptTemplate = (template: string): {
  valid: boolean;
  error?: string;
  warnings?: string[];
  parameters?: string[];
} => {
  if (!template || template.trim() === '') {
    return { valid: false, error: 'Prompt template cannot be empty' };
  }

  const warnings: string[] = [];
  const parameters = extractTemplateParameters(template);

  // Check for unclosed template variables
  const openBraces = (template.match(/\{\{/g) || []).length;
  const closeBraces = (template.match(/\}\}/g) || []).length;

  if (openBraces !== closeBraces) {
    return {
      valid: false,
      error: 'Unclosed template variables - check {{ }} braces',
      parameters
    };
  }

  // Check for invalid parameter names
  for (const param of parameters) {
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(param)) {
      return {
        valid: false,
        error: `Invalid parameter name "${param}" - use only letters, numbers, and underscores`,
        parameters
      };
    }
  }

  // Check for potential issues
  if (template.includes('{{') && !template.includes('}}')) {
    warnings.push('Template may have unclosed variable references');
  }

  if (parameters.length === 0) {
    warnings.push('Template does not contain any parameters - consider if this is intentional');
  }

  if (template.length < 20) {
    warnings.push('Template is very short - ensure it provides sufficient context');
  }

  if (template.length > 5000) {
    warnings.push('Template is very long - consider breaking into smaller, focused prompts');
  }

  return {
    valid: true,
    warnings: warnings.length > 0 ? warnings : undefined,
    parameters
  };
};

// Generate parameter template
export const generateParameterTemplate = (parameterNames: string[]): string => {
  const parameters = parameterNames.map(name => ({
    name,
    type: 'string',
    description: `Description for ${name}`,
    required: true
  }));

  return JSON.stringify(parameters, null, 2);
};

// Prompt category suggestions based on content
export const suggestPromptCategory = (template: string, name?: string): typeof enums.promptCategory[number] => {
  const text = `${name || ''} ${template}`.toLowerCase();

  const categoryKeywords = {
    coding: ['code', 'function', 'debug', 'programming', 'algorithm', 'syntax', 'bug'],
    analysis: ['analyze', 'examine', 'review', 'evaluate', 'assess', 'study', 'investigate'],
    creative: ['creative', 'story', 'write', 'generate', 'imagine', 'design', 'brainstorm'],
    educational: ['explain', 'teach', 'learn', 'lesson', 'tutorial', 'guide', 'help'],
    business: ['business', 'strategy', 'plan', 'proposal', 'meeting', 'report', 'analysis'],
  };

  for (const [category, keywords] of Object.entries(categoryKeywords)) {
    if (keywords.some(keyword => text.includes(keyword))) {
      return category as typeof enums.promptCategory[number];
    }
  }

  return 'general';
};

// Template examples by category
export const getTemplateExamples = (category: typeof enums.promptCategory[number]): string[] => {
  const examples = {
    general: [
      'Please help me with {{task}}. I need {{details}}.',
      'Analyze the following {{type}}: {{content}}'
    ],
    coding: [
      'Review this {{language}} code and suggest improvements:\n\n{{code}}',
      'Debug this {{language}} function that should {{purpose}}:\n\n{{code}}'
    ],
    analysis: [
      'Analyze the following {{data_type}} and provide insights on {{focus_area}}:\n\n{{data}}',
      'Compare {{item1}} and {{item2}} in terms of {{criteria}}'
    ],
    creative: [
      'Write a {{genre}} story about {{topic}} with {{mood}} tone.',
      'Create a {{type}} design concept for {{purpose}} targeting {{audience}}'
    ],
    educational: [
      'Explain {{concept}} in simple terms suitable for {{audience}}',
      'Create a lesson plan for teaching {{subject}} to {{grade_level}}'
    ],
    business: [
      'Create a business plan for {{business_type}} in the {{industry}} market',
      'Analyze the market opportunity for {{product}} in {{region}}'
    ],
    custom: [
      'Custom template for {{specific_use_case}} with {{parameters}}'
    ]
  };

  return examples[category] || examples.general;
};

// Prompt creation schema
export const createPromptSchema = promptSchemaWithTemplateValidation.omit({ usage_count: true });

// Prompt update schema
export const updatePromptSchema = promptSchemaWithTemplateValidation.partial();

export type PromptFormData = z.infer<typeof promptSchema>;
export type CreatePromptData = z.infer<typeof createPromptSchema>;
export type UpdatePromptData = z.infer<typeof updatePromptSchema>;
