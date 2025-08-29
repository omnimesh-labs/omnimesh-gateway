import * as z from 'zod';
import {
  userSchema,
  passwordSchema,
  createUserSchema,
  registerUserSchema,
  updateUserSchema,
  changePasswordSchema,
  updateProfileSchema,
  loginSchema,
  passwordResetRequestSchema,
  passwordResetSchema,
  rolePermissions,
  validateRoleAssignment,
  validatePasswordStrength,
  validateEmailDomain,
  generateUsernameFromEmail
} from '../user';

describe('User Validation Schemas', () => {
  describe('userSchema', () => {
    it('should validate correct user data', () => {
      const validUser = {
        email: 'test@example.com',
        name: 'John Doe',
        role: 'user' as const,
        is_active: true,
        email_verified: false
      };

      expect(() => userSchema.parse(validUser)).not.toThrow();
    });

    it('should apply defaults for optional fields', () => {
      const minimalUser = {
        email: 'test@example.com',
        name: 'John Doe'
      };

      const parsed = userSchema.parse(minimalUser);
      expect(parsed.role).toBe('user');
      expect(parsed.is_active).toBe(true);
      expect(parsed.email_verified).toBe(false);
    });

    it('should convert email to lowercase', () => {
      const user = {
        email: 'TEST@EXAMPLE.COM',
        name: 'John Doe'
      };

      const parsed = userSchema.parse(user);
      expect(parsed.email).toBe('test@example.com');
    });

    it('should trim name whitespace', () => {
      const user = {
        email: 'test@example.com',
        name: '  John Doe  '
      };

      const parsed = userSchema.parse(user);
      expect(parsed.name).toBe('John Doe');
    });

    it('should reject invalid email formats', () => {
      const invalidUser = {
        email: 'invalid-email',
        name: 'John Doe'
      };

      expect(() => userSchema.parse(invalidUser)).toThrow('Invalid email address');
    });

    it('should reject empty names', () => {
      const invalidUser = {
        email: 'test@example.com',
        name: ''
      };

      expect(() => userSchema.parse(invalidUser)).toThrow('Name is required');
    });

    it('should reject names that are too long', () => {
      const invalidUser = {
        email: 'test@example.com',
        name: 'x'.repeat(256)
      };

      expect(() => userSchema.parse(invalidUser)).toThrow('Name must be less than 255 characters');
    });

    it('should validate user roles', () => {
      const validRoles = ['admin', 'user', 'viewer', 'api_user'];

      validRoles.forEach(role => {
        const user = {
          email: 'test@example.com',
          name: 'John Doe',
          role: role as any
        };

        expect(() => userSchema.parse(user)).not.toThrow();
      });
    });

    it('should reject invalid user roles', () => {
      const user = {
        email: 'test@example.com',
        name: 'John Doe',
        role: 'invalid-role'
      };

      expect(() => userSchema.parse(user)).toThrow();
    });
  });

  describe('passwordSchema', () => {
    it('should validate strong passwords', () => {
      const strongPassword = {
        password: 'StrongPassword123!',
        confirmPassword: 'StrongPassword123!'
      };

      expect(() => passwordSchema.parse(strongPassword)).not.toThrow();
    });

    it('should require minimum length', () => {
      const shortPassword = {
        password: 'Short1!',
        confirmPassword: 'Short1!'
      };

      expect(() => passwordSchema.parse(shortPassword)).toThrow('Password must be at least 8 characters');
    });

    it('should require uppercase letter', () => {
      const noUppercase = {
        password: 'lowercase123!',
        confirmPassword: 'lowercase123!'
      };

      expect(() => passwordSchema.parse(noUppercase)).toThrow('Password must contain at least one uppercase letter');
    });

    it('should require lowercase letter', () => {
      const noLowercase = {
        password: 'UPPERCASE123!',
        confirmPassword: 'UPPERCASE123!'
      };

      expect(() => passwordSchema.parse(noLowercase)).toThrow('Password must contain at least one lowercase letter');
    });

    it('should require number', () => {
      const noNumber = {
        password: 'NoNumberHere!',
        confirmPassword: 'NoNumberHere!'
      };

      expect(() => passwordSchema.parse(noNumber)).toThrow('Password must contain at least one number');
    });

    it('should require special character', () => {
      const noSpecial = {
        password: 'NoSpecial123',
        confirmPassword: 'NoSpecial123'
      };

      expect(() => passwordSchema.parse(noSpecial)).toThrow('Password must contain at least one special character');
    });

    it('should validate password confirmation match', () => {
      const mismatchedPasswords = {
        password: 'StrongPassword123!',
        confirmPassword: 'DifferentPassword123!'
      };

      expect(() => passwordSchema.parse(mismatchedPasswords)).toThrow('Passwords do not match');
    });

    it('should reject too long passwords', () => {
      const tooLongPassword = 'A1!' + 'x'.repeat(126);
      const longPassword = {
        password: tooLongPassword,
        confirmPassword: tooLongPassword
      };

      expect(() => passwordSchema.parse(longPassword)).toThrow('Password must be less than 128 characters');
    });
  });

  describe('changePasswordSchema', () => {
    it('should validate password change', () => {
      const validChange = {
        currentPassword: 'OldPassword123!',
        newPassword: 'NewPassword456!',
        confirmNewPassword: 'NewPassword456!'
      };

      expect(() => changePasswordSchema.parse(validChange)).not.toThrow();
    });

    it('should require current password', () => {
      const noCurrentPassword = {
        currentPassword: '',
        newPassword: 'NewPassword456!',
        confirmNewPassword: 'NewPassword456!'
      };

      expect(() => changePasswordSchema.parse(noCurrentPassword)).toThrow('Current password is required');
    });

    it('should validate new password confirmation', () => {
      const mismatchedNew = {
        currentPassword: 'OldPassword123!',
        newPassword: 'NewPassword456!',
        confirmNewPassword: 'DifferentPassword456!'
      };

      expect(() => changePasswordSchema.parse(mismatchedNew)).toThrow('New passwords do not match');
    });

    it('should require new password to be different', () => {
      const samePassword = {
        currentPassword: 'SamePassword123!',
        newPassword: 'SamePassword123!',
        confirmNewPassword: 'SamePassword123!'
      };

      expect(() => changePasswordSchema.parse(samePassword)).toThrow('New password must be different from current password');
    });
  });

  describe('updateProfileSchema', () => {
    it('should validate profile updates', () => {
      const validProfile = {
        name: 'Updated Name',
        email: 'updated@example.com'
      };

      expect(() => updateProfileSchema.parse(validProfile)).not.toThrow();
    });

    it('should trim name and lowercase email', () => {
      const profile = {
        name: '  Updated Name  ',
        email: 'UPDATED@EXAMPLE.COM'
      };

      const parsed = updateProfileSchema.parse(profile);
      expect(parsed.name).toBe('Updated Name');
      expect(parsed.email).toBe('updated@example.com');
    });
  });

  describe('loginSchema', () => {
    it('should validate login credentials', () => {
      const validLogin = {
        email: 'user@example.com',
        password: 'password123',
        remember: true
      };

      expect(() => loginSchema.parse(validLogin)).not.toThrow();
    });

    it('should default remember to false', () => {
      const login = {
        email: 'user@example.com',
        password: 'password123'
      };

      const parsed = loginSchema.parse(login);
      expect(parsed.remember).toBe(false);
    });

    it('should require password', () => {
      const noPassword = {
        email: 'user@example.com',
        password: ''
      };

      expect(() => loginSchema.parse(noPassword)).toThrow('Password is required');
    });
  });

  describe('passwordResetSchema', () => {
    it('should validate password reset', () => {
      const validReset = {
        token: 'valid-reset-token',
        password: 'NewPassword123!',
        confirmPassword: 'NewPassword123!'
      };

      expect(() => passwordResetSchema.parse(validReset)).not.toThrow();
    });

    it('should require reset token', () => {
      const noToken = {
        token: '',
        password: 'NewPassword123!',
        confirmPassword: 'NewPassword123!'
      };

      expect(() => passwordResetSchema.parse(noToken)).toThrow('Reset token is required');
    });
  });

  describe('schema compositions', () => {
    it('should create user with createUserSchema', () => {
      const createUser = {
        email: 'new@example.com',
        name: 'New User',
        password: 'NewPassword123!'
      };

      expect(() => createUserSchema.parse(createUser)).not.toThrow();
    });

    it('should register user with registerUserSchema', () => {
      const registerUser = {
        email: 'register@example.com',
        name: 'Register User',
        password: 'RegisterPassword123!',
        confirmPassword: 'RegisterPassword123!'
      };

      expect(() => registerUserSchema.parse(registerUser)).not.toThrow();
    });

    it('should allow partial updates with updateUserSchema', () => {
      const partialUpdate = {
        name: 'Updated Name Only'
      };

      expect(() => updateUserSchema.parse(partialUpdate)).not.toThrow();
    });
  });
});

describe('Role Permissions', () => {
  describe('rolePermissions structure', () => {
    it('should define permissions for all roles', () => {
      const requiredRoles = ['admin', 'user', 'viewer', 'api_user'];

      requiredRoles.forEach(role => {
        expect(rolePermissions[role as keyof typeof rolePermissions]).toBeDefined();
        expect(typeof rolePermissions[role as keyof typeof rolePermissions].description).toBe('string');
      });
    });

    it('should give admin full permissions', () => {
      const adminPerms = rolePermissions.admin;
      expect(adminPerms.canManageUsers).toBe(true);
      expect(adminPerms.canManageOrganization).toBe(true);
      expect(adminPerms.canManageServers).toBe(true);
      expect(adminPerms.canViewLogs).toBe(true);
      expect(adminPerms.canManageSettings).toBe(true);
    });

    it('should give viewer minimal permissions', () => {
      const viewerPerms = rolePermissions.viewer;
      expect(viewerPerms.canManageUsers).toBe(false);
      expect(viewerPerms.canManageOrganization).toBe(false);
      expect(viewerPerms.canManageServers).toBe(false);
      expect(viewerPerms.canViewLogs).toBe(false);
      expect(viewerPerms.canManageSettings).toBe(false);
    });

    it('should give user and api_user server permissions', () => {
      const userPerms = rolePermissions.user;
      const apiUserPerms = rolePermissions.api_user;

      expect(userPerms.canManageServers).toBe(true);
      expect(apiUserPerms.canManageServers).toBe(true);
    });
  });
});

describe('Role Assignment Validation', () => {
  describe('validateRoleAssignment', () => {
    it('should allow admin to assign any role', () => {
      const roles = ['admin', 'user', 'viewer', 'api_user'] as const;

      roles.forEach(targetRole => {
        const result = validateRoleAssignment('admin', targetRole);
        expect(result.valid).toBe(true);
      });
    });

    it('should not allow non-admin to assign admin role', () => {
      const nonAdminRoles = ['user', 'viewer', 'api_user'] as const;

      nonAdminRoles.forEach(assignerRole => {
        const result = validateRoleAssignment(assignerRole, 'admin');
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Only administrators can assign admin role');
      });
    });

    it('should not allow user/viewer/api_user to assign any roles', () => {
      const restrictedRoles = ['user', 'viewer', 'api_user'] as const;

      restrictedRoles.forEach(assignerRole => {
        const result = validateRoleAssignment(assignerRole, 'user');
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Insufficient permissions to assign roles');
      });
    });
  });
});

describe('Password Strength Validation', () => {
  describe('validatePasswordStrength', () => {
    it('should rate weak passwords', () => {
      const weakPasswords = ['123456', 'password', 'abc123'];

      weakPasswords.forEach(password => {
        const result = validatePasswordStrength(password);
        expect(result.strength).toBe('weak');
        expect(result.score).toBeLessThanOrEqual(2);
        expect(result.feedback.length).toBeGreaterThan(0);
      });
    });

    it('should rate strong passwords', () => {
      const strongPasswords = ['MyVeryStr0ng!P@ssw0rd2024', 'C0mpl3x&V3ryL0ngP@ssw0rd!'];

      strongPasswords.forEach(password => {
        const result = validatePasswordStrength(password);
        expect(result.strength).toBe('strong');
        expect(result.score).toBeGreaterThanOrEqual(7);
      });
    });

    it('should detect common patterns', () => {
      const commonPatterns = ['password123!', 'qwerty123!', 'abcdef123!'];

      commonPatterns.forEach(password => {
        const result = validatePasswordStrength(password);
        expect(result.feedback).toContain('Avoid common patterns and dictionary words');
      });
    });

    it('should provide specific feedback', () => {
      const result = validatePasswordStrength('short');
      expect(result.feedback).toContain('Use at least 8 characters');
      expect(result.feedback).toContain('Add uppercase letters');
      expect(result.feedback).toContain('Add numbers');
      expect(result.feedback).toContain('Add special characters (!@#$%^&*)');
    });

    it('should rate fair passwords', () => {
      const fairPassword = 'GoodPass1';
      const result = validatePasswordStrength(fairPassword);
      expect(result.strength).toBe('fair');
      expect(result.score).toBeGreaterThan(2);
      expect(result.score).toBeLessThanOrEqual(4);
    });

    it('should rate good passwords', () => {
      const goodPassword = 'MyStrong1Pass!';
      const result = validatePasswordStrength(goodPassword);
      expect(['good', 'strong']).toContain(result.strength);
      expect(result.score).toBeGreaterThan(4);
    });
  });
});

describe('Email Domain Validation', () => {
  describe('validateEmailDomain', () => {
    it('should accept valid domains', () => {
      const validEmails = ['user@gmail.com', 'test@company.co.uk', 'admin@example.org'];

      validEmails.forEach(email => {
        const result = validateEmailDomain(email);
        expect(result.valid).toBe(true);
        expect(result.error).toBeUndefined();
      });
    });

    it('should detect common domain typos', () => {
      const typoMap = {
        'user@gmail.co': 'gmail.com',
        'test@gmial.com': 'gmail.com',
        'admin@yahooo.com': 'yahoo.com',
        'contact@hotmial.com': 'hotmail.com',
        'support@outlok.com': 'outlook.com'
      };

      Object.entries(typoMap).forEach(([email, suggestion]) => {
        const result = validateEmailDomain(email);
        expect(result.valid).toBe(false);
        expect(result.error).toBe(`Did you mean ${suggestion}?`);
      });
    });

    it('should reject invalid domain formats', () => {
      const invalidEmails = ['user@invalid.', 'test@.com', 'admin@domain'];

      invalidEmails.forEach(email => {
        const result = validateEmailDomain(email);
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Invalid email domain format');
      });
    });

    it('should handle malformed emails', () => {
      const result = validateEmailDomain('invalid-email');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Invalid email format');
    });
  });
});

describe('Username Generation', () => {
  describe('generateUsernameFromEmail', () => {
    it('should extract username from email', () => {
      expect(generateUsernameFromEmail('johndoe@example.com')).toBe('johndoe');
      expect(generateUsernameFromEmail('test_user@domain.org')).toBe('test_user');
    });

    it('should clean special characters', () => {
      expect(generateUsernameFromEmail('user+tag@example.com')).toBe('usertag');
      expect(generateUsernameFromEmail('test-user@domain.com')).toBe('testuser');
    });

    it('should convert to lowercase', () => {
      expect(generateUsernameFromEmail('JohnDoe@Example.Com')).toBe('johndoe');
      expect(generateUsernameFromEmail('TEST_USER@DOMAIN.ORG')).toBe('test_user');
    });

    it('should preserve numbers and underscores', () => {
      expect(generateUsernameFromEmail('user_123@example.com')).toBe('user_123');
      expect(generateUsernameFromEmail('test123@domain.org')).toBe('test123');
    });
  });
});
