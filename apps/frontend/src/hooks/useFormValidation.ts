import { useState, useCallback, useEffect } from 'react';
import { z } from 'zod';
import { validationUtils } from '@/lib/utils/validation';

// Form validation state
interface ValidationState {
  errors: Record<string, string>;
  isValidating: boolean;
  isValid: boolean;
  isDirty: boolean;
  touchedFields: Set<string>;
}

// Validation options
interface ValidationOptions {
  validateOnChange?: boolean;
  validateOnBlur?: boolean;
  debounceMs?: number;
  validateOnMount?: boolean;
}

// Form validation hook
export function useFormValidation<T extends Record<string, any>>(
  schema: z.ZodSchema<T>,
  initialData?: Partial<T>,
  options: ValidationOptions = {}
) {
  const {
    validateOnChange = true,
    validateOnBlur = true,
    debounceMs = 300,
    validateOnMount = false
  } = options;

  const [data, setData] = useState<Partial<T>>(initialData || {});
  const [validationState, setValidationState] = useState<ValidationState>({
    errors: {},
    isValidating: false,
    isValid: false,
    isDirty: false,
    touchedFields: new Set()
  });

  // Debounced validation function
  const debouncedValidate = useCallback(
    validationUtils.debounce(async (dataToValidate: Partial<T>) => {
      setValidationState(prev => ({ ...prev, isValidating: true }));

      try {
        await schema.parseAsync(dataToValidate);
        setValidationState(prev => ({
          ...prev,
          errors: {},
          isValid: true,
          isValidating: false
        }));
      } catch (error) {
        if (error instanceof z.ZodError) {
          const errors = validationUtils.extractZodErrors(error);
          setValidationState(prev => ({
            ...prev,
            errors,
            isValid: false,
            isValidating: false
          }));
        }
      }
    }, debounceMs),
    [schema, debounceMs]
  );

  // Validate specific field
  const validateField = useCallback(async (fieldName: keyof T, value: any) => {
    const fieldPath = String(fieldName);

    try {
      // Create a partial schema for the field
      const fieldSchema = schema.pick({ [fieldName]: true } as any);
      await fieldSchema.parseAsync({ [fieldName]: value });

      setValidationState(prev => ({
        ...prev,
        errors: { ...prev.errors, [fieldPath]: '' }
      }));

      return true;
    } catch (error) {
      if (error instanceof z.ZodError) {
        const fieldError = error.errors.find(err =>
          err.path.includes(fieldName as string)
        );

        if (fieldError) {
          setValidationState(prev => ({
            ...prev,
            errors: { ...prev.errors, [fieldPath]: fieldError.message }
          }));
        }
      }
      return false;
    }
  }, [schema]);

  // Update field value
  const updateField = useCallback((fieldName: keyof T, value: any) => {
    const newData = { ...data, [fieldName]: value };
    setData(newData);

    setValidationState(prev => ({
      ...prev,
      isDirty: true,
      touchedFields: new Set(prev.touchedFields).add(String(fieldName))
    }));

    if (validateOnChange) {
      debouncedValidate(newData);
    }
  }, [data, validateOnChange, debouncedValidate]);

  // Handle field blur
  const handleFieldBlur = useCallback((fieldName: keyof T) => {
    setValidationState(prev => ({
      ...prev,
      touchedFields: new Set(prev.touchedFields).add(String(fieldName))
    }));

    if (validateOnBlur) {
      const value = data[fieldName];
      validateField(fieldName, value);
    }
  }, [data, validateOnBlur, validateField]);

  // Validate entire form
  const validateForm = useCallback(async (): Promise<boolean> => {
    setValidationState(prev => ({ ...prev, isValidating: true }));

    try {
      await schema.parseAsync(data);
      setValidationState(prev => ({
        ...prev,
        errors: {},
        isValid: true,
        isValidating: false
      }));
      return true;
    } catch (error) {
      if (error instanceof z.ZodError) {
        const errors = validationUtils.extractZodErrors(error);
        setValidationState(prev => ({
          ...prev,
          errors,
          isValid: false,
          isValidating: false
        }));
      }
      return false;
    }
  }, [data, schema]);

  // Reset form
  const resetForm = useCallback((newData?: Partial<T>) => {
    const resetData = newData || initialData || {};
    setData(resetData);
    setValidationState({
      errors: {},
      isValidating: false,
      isValid: false,
      isDirty: false,
      touchedFields: new Set()
    });
  }, [initialData]);

  // Get field error
  const getFieldError = useCallback((fieldName: keyof T): string => {
    const fieldPath = String(fieldName);
    return validationState.errors[fieldPath] || '';
  }, [validationState.errors]);

  // Check if field has error
  const hasFieldError = useCallback((fieldName: keyof T): boolean => {
    const fieldPath = String(fieldName);
    return Boolean(validationState.errors[fieldPath]);
  }, [validationState.errors]);

  // Check if field is touched
  const isFieldTouched = useCallback((fieldName: keyof T): boolean => {
    return validationState.touchedFields.has(String(fieldName));
  }, [validationState.touchedFields]);

  // Get field props for form inputs
  const getFieldProps = useCallback((fieldName: keyof T) => ({
    value: data[fieldName] || '',
    onChange: (value: any) => updateField(fieldName, value),
    onBlur: () => handleFieldBlur(fieldName),
    error: getFieldError(fieldName),
    hasError: hasFieldError(fieldName),
    isTouched: isFieldTouched(fieldName),
  }), [data, updateField, handleFieldBlur, getFieldError, hasFieldError, isFieldTouched]);

  // Validate on mount if required
  useEffect(() => {
    if (validateOnMount && Object.keys(data).length > 0) {
      debouncedValidate(data);
    }
  }, [validateOnMount, debouncedValidate]); // Exclude data from dependencies to avoid infinite loops

  return {
    // Form data
    data,
    setData,

    // Validation state
    ...validationState,

    // Field operations
    updateField,
    handleFieldBlur,
    getFieldError,
    hasFieldError,
    isFieldTouched,
    getFieldProps,

    // Form operations
    validateForm,
    validateField,
    resetForm,

    // Computed properties
    hasErrors: Object.keys(validationState.errors).some(key => validationState.errors[key]),
    errorCount: Object.keys(validationState.errors).filter(key => validationState.errors[key]).length,
    touchedFieldCount: validationState.touchedFields.size,
  };
}

// Async validation hook for server-side validation
export function useAsyncValidation<T>(
  validateFn: (value: T) => Promise<{ valid: boolean; error?: string }>,
  debounceMs: number = 500
) {
  const [isValidating, setIsValidating] = useState(false);
  const [error, setError] = useState<string>('');
  const [isValid, setIsValid] = useState<boolean | null>(null);

  const debouncedValidate = useCallback(
    validationUtils.debounce(async (value: T) => {
      if (!value) {
        setError('');
        setIsValid(null);
        setIsValidating(false);
        return;
      }

      setIsValidating(true);

      try {
        const result = await validateFn(value);
        setError(result.error || '');
        setIsValid(result.valid);
      } catch (err) {
        setError('Validation failed');
        setIsValid(false);
      } finally {
        setIsValidating(false);
      }
    }, debounceMs),
    [validateFn, debounceMs]
  );

  const validate = useCallback((value: T) => {
    debouncedValidate(value);
  }, [debouncedValidate]);

  const reset = useCallback(() => {
    setIsValidating(false);
    setError('');
    setIsValid(null);
  }, []);

  return {
    validate,
    reset,
    isValidating,
    error,
    isValid,
    hasError: Boolean(error),
  };
}

// Field-specific validation hook
export function useFieldValidation<T>(
  schema: z.ZodSchema<T>,
  fieldName: keyof T,
  debounceMs: number = 300
) {
  const [value, setValue] = useState<T[keyof T]>();
  const [error, setError] = useState<string>('');
  const [isValidating, setIsValidating] = useState(false);
  const [isTouched, setIsTouched] = useState(false);

  const debouncedValidate = useCallback(
    validationUtils.debounce(async (fieldValue: T[keyof T]) => {
      if (fieldValue === undefined || fieldValue === '') {
        setError('');
        setIsValidating(false);
        return;
      }

      setIsValidating(true);

      try {
        // Create a partial validation with just this field
        const partialSchema = z.object({ [fieldName]: schema.shape[fieldName] });
        await partialSchema.parseAsync({ [fieldName]: fieldValue });
        setError('');
      } catch (err) {
        if (err instanceof z.ZodError) {
          const fieldError = err.errors.find(e => e.path.includes(String(fieldName)));
          setError(fieldError?.message || 'Invalid value');
        }
      } finally {
        setIsValidating(false);
      }
    }, debounceMs),
    [schema, fieldName, debounceMs]
  );

  const updateValue = useCallback((newValue: T[keyof T]) => {
    setValue(newValue);
    setIsTouched(true);
    debouncedValidate(newValue);
  }, [debouncedValidate]);

  const reset = useCallback(() => {
    setValue(undefined);
    setError('');
    setIsValidating(false);
    setIsTouched(false);
  }, []);

  return {
    value,
    error,
    isValidating,
    isTouched,
    hasError: Boolean(error),
    updateValue,
    reset,
  };
}
