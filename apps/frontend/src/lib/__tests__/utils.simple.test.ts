import { cn } from '../utils';

describe('Utils', () => {
  describe('cn function (className merging)', () => {
    it('should merge single class names', () => {
      expect(cn('class1')).toBe('class1');
      expect(cn('text-lg')).toBe('text-lg');
      expect(cn('bg-blue-500')).toBe('bg-blue-500');
    });

    it('should merge multiple class names', () => {
      expect(cn('class1', 'class2')).toBe('class1 class2');
      expect(cn('text-lg', 'font-bold', 'text-blue-500')).toBe('text-lg font-bold text-blue-500');
    });

    it('should handle arrays of class names', () => {
      expect(cn(['class1', 'class2'])).toBe('class1 class2');
      expect(cn(['text-lg', 'font-bold'], 'text-blue-500')).toBe('text-lg font-bold text-blue-500');
    });

    it('should handle conditional class names with objects', () => {
      expect(cn({
        'active': true,
        'inactive': false
      })).toBe('active');

      expect(cn({
        'text-lg': true,
        'text-sm': false,
        'font-bold': true
      })).toBe('text-lg font-bold');
    });

    it('should handle mixed inputs', () => {
      expect(cn(
        'base-class',
        ['conditional-class'],
        { 'active': true, 'disabled': false },
        'final-class'
      )).toBe('base-class conditional-class active final-class');
    });

    it('should handle undefined and null values', () => {
      expect(cn('class1', undefined, 'class2', null)).toBe('class1 class2');
      expect(cn(undefined, null)).toBe('');
    });

    it('should handle empty strings and whitespace', () => {
      expect(cn('', 'class1', '  ', 'class2')).toBe('class1 class2');
      expect(cn('   class1   ', '  class2  ')).toBe('class1 class2');
    });

    it('should handle Tailwind CSS class conflicts', () => {
      // Test that tailwind-merge resolves conflicts correctly
      expect(cn('p-4', 'p-6')).toBe('p-6'); // Later padding wins
      expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500'); // Later text color wins
      expect(cn('bg-red-100', 'bg-blue-200')).toBe('bg-blue-200'); // Later background wins
    });

    it('should handle responsive class conflicts', () => {
      expect(cn('text-sm', 'md:text-lg', 'text-base')).toBe('md:text-lg text-base');
      expect(cn('p-2', 'md:p-4', 'lg:p-6')).toBe('p-2 md:p-4 lg:p-6'); // No conflicts across breakpoints
    });

    it('should handle hover and state variants', () => {
      expect(cn('bg-blue-500', 'hover:bg-blue-600', 'focus:bg-blue-700'))
        .toBe('bg-blue-500 hover:bg-blue-600 focus:bg-blue-700');
    });

    it('should handle size class conflicts', () => {
      expect(cn('w-4', 'w-6')).toBe('w-6');
      expect(cn('h-4', 'h-6')).toBe('h-6');
      expect(cn('w-4', 'h-4', 'w-6', 'h-6')).toBe('w-6 h-6');
    });

    it('should handle margin and padding conflicts', () => {
      expect(cn('m-2', 'm-4')).toBe('m-4');
      expect(cn('mx-2', 'mx-4')).toBe('mx-4');
      expect(cn('my-2', 'my-4')).toBe('my-4');
      expect(cn('mt-2', 'mt-4')).toBe('mt-4');
      expect(cn('p-2', 'p-4')).toBe('p-4');
      expect(cn('px-2', 'px-4')).toBe('px-4');
      expect(cn('py-2', 'py-4')).toBe('py-4');
      expect(cn('pt-2', 'pt-4')).toBe('pt-4');
    });

    it('should handle border class conflicts', () => {
      expect(cn('border', 'border-2')).toBe('border-2');
      expect(cn('border-gray-200', 'border-blue-300')).toBe('border-blue-300');
      expect(cn('border-solid', 'border-dashed')).toBe('border-dashed');
    });

    it('should handle text alignment conflicts', () => {
      expect(cn('text-left', 'text-center')).toBe('text-center');
      expect(cn('text-center', 'text-right')).toBe('text-right');
    });

    it('should handle display conflicts', () => {
      expect(cn('block', 'inline')).toBe('inline');
      expect(cn('hidden', 'block')).toBe('block');
      expect(cn('flex', 'grid')).toBe('grid');
    });

    it('should handle position conflicts', () => {
      expect(cn('relative', 'absolute')).toBe('absolute');
      expect(cn('static', 'fixed')).toBe('fixed');
    });

    it('should handle flex direction conflicts', () => {
      expect(cn('flex-row', 'flex-col')).toBe('flex-col');
      expect(cn('flex-col', 'flex-row-reverse')).toBe('flex-row-reverse');
    });

    it('should handle justify-content conflicts', () => {
      expect(cn('justify-start', 'justify-center')).toBe('justify-center');
      expect(cn('justify-center', 'justify-end')).toBe('justify-end');
    });

    it('should handle align-items conflicts', () => {
      expect(cn('items-start', 'items-center')).toBe('items-center');
      expect(cn('items-center', 'items-end')).toBe('items-end');
    });

    it('should preserve non-conflicting classes', () => {
      expect(cn('text-lg', 'font-bold', 'bg-blue-500', 'p-4'))
        .toBe('text-lg font-bold bg-blue-500 p-4');
    });

    it('should handle complex real-world scenarios', () => {
      // Button component with base classes and conditional states
      const baseClasses = 'inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md';
      const variantClasses = 'text-white bg-blue-600 hover:bg-blue-700 focus:bg-blue-700';
      const sizeClasses = 'px-6 py-3 text-base'; // Override base padding and text size

      const result = cn(baseClasses, variantClasses, sizeClasses);
      expect(result).toContain('inline-flex items-center justify-center');
      expect(result).toContain('text-white bg-blue-600 hover:bg-blue-700 focus:bg-blue-700');
      expect(result).toContain('px-6 py-3 text-base'); // Should override px-4 py-2 text-sm
      expect(result).not.toContain('px-4');
      expect(result).not.toContain('py-2');
      expect(result).not.toContain('text-sm');
    });

    it('should handle form input scenarios', () => {
      const baseInput = 'block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm';
      const focusClasses = 'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500';
      const errorClasses = 'border-red-500 focus:ring-red-500 focus:border-red-500';

      const normalInput = cn(baseInput, focusClasses);
      const errorInput = cn(baseInput, focusClasses, errorClasses);

      expect(normalInput).toContain('border-gray-300');
      expect(normalInput).toContain('focus:ring-blue-500 focus:border-blue-500');

      expect(errorInput).toContain('border-red-500'); // Should override border-gray-300
      expect(errorInput).toContain('focus:ring-red-500 focus:border-red-500'); // Should override blue focus states
      expect(errorInput).not.toContain('border-gray-300');
      expect(errorInput).not.toContain('focus:ring-blue-500');
    });

    it('should handle grid layout scenarios', () => {
      const gridContainer = cn(
        'grid',
        'grid-cols-1', // base: 1 column
        'md:grid-cols-2', // medium: 2 columns
        'lg:grid-cols-3', // large: 3 columns
        'gap-4',
        'gap-6' // Should override gap-4
      );

      expect(gridContainer).toBe('grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6');
      expect(gridContainer).not.toContain('gap-4');
    });

    it('should handle dark mode variants', () => {
      const darkModeClasses = cn(
        'bg-white text-black', // Light mode
        'dark:bg-gray-800 dark:text-white', // Dark mode
        'bg-gray-100', // Should override bg-white
        'dark:bg-black' // Should override dark:bg-gray-800
      );

      expect(darkModeClasses).toContain('bg-gray-100');
      expect(darkModeClasses).toContain('text-black');
      expect(darkModeClasses).toContain('dark:bg-black');
      expect(darkModeClasses).toContain('dark:text-white');
      expect(darkModeClasses).not.toContain('bg-white');
      expect(darkModeClasses).not.toContain('dark:bg-gray-800');
    });

    it('should handle empty and falsy values gracefully', () => {
      expect(cn()).toBe('');
      expect(cn('')).toBe('');
      expect(cn(false)).toBe('');
      expect(cn(null)).toBe('');
      expect(cn(undefined)).toBe('');
      expect(cn(0)).toBe('');
      expect(cn([])).toBe('');
      expect(cn({})).toBe('');
    });

    it('should handle nested arrays', () => {
      expect(cn(['class1', ['class2', 'class3']], 'class4'))
        .toBe('class1 class2 class3 class4');
    });

    it('should handle arbitrary values', () => {
      expect(cn('bg-[#1da1f2]', 'bg-[#ff0000]')).toBe('bg-[#ff0000]');
      expect(cn('w-[100px]', 'w-[200px]')).toBe('w-[200px]');
    });

    it('should preserve important modifiers', () => {
      expect(cn('text-red-500', '!text-blue-500')).toBe('text-red-500 !text-blue-500');
    });

    it('should handle performance with many classes', () => {
      const manyClasses = Array.from({length: 100}, (_, i) => `class-${i}`);
      const result = cn(...manyClasses);
      expect(result).toContain('class-0');
      expect(result).toContain('class-99');
      expect(result.split(' ')).toHaveLength(100);
    });
  });
});
