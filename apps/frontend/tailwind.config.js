/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      // Custom breakpoints for better performance
      screens: {
        'xs': '0px',
        'sm': '600px',
        'md': '960px',
        'lg': '1280px',
        'xl': '1920px',
      },
      // Custom containers
      container: {
        center: true,
        padding: '1rem',
        screens: {
          '3xs': '16rem',
          '2xs': '18rem',
          'xs': '20rem',
          'sm': '24rem',
          'md': '28rem',
          'lg': '32rem',
          'xl': '36rem',
          '2xl': '42rem',
          '3xl': '48rem',
          '4xl': '56rem',
          '5xl': '64rem',
          '6xl': '72rem',
          '7xl': '80rem',
        }
      },
      // Custom font sizes for performance
      fontSize: {
        'xs': ['0.5625rem', { lineHeight: '1.77' }],   // 9px
        'sm': ['0.6875rem', { lineHeight: '1.82' }],   // 11px
        'md': ['0.75rem', { lineHeight: '1.67' }],     // 12px
        'base': ['0.8125rem', { lineHeight: '1.85' }], // 13px
        'lg': ['0.875rem', { lineHeight: '1.71' }],    // 14px
        'xl': ['1rem', { lineHeight: '1.5' }],         // 16px
        '2xl': ['1.125rem', { lineHeight: '1.56' }],   // 18px
        '3xl': ['1.375rem', { lineHeight: '1.27' }],   // 22px
        '4xl': ['1.5rem', { lineHeight: '1.33' }],     // 24px
        '5xl': ['1.75rem', { lineHeight: '1.14' }],    // 28px
        '6xl': ['2.25rem', { lineHeight: '0.89' }],    // 36px
        '7xl': ['3rem', { lineHeight: '0.67' }],       // 48px
        '8xl': ['3.5rem', { lineHeight: '0.57' }],     // 56px
        '9xl': ['4rem', { lineHeight: '0.5' }],        // 64px
        '10xl': ['5rem', { lineHeight: '0.4' }],       // 80px
      }
    },
  },
  plugins: [
    // Custom icon-size plugin
    function ({ addUtilities, theme, matchUtilities }) {
      const spacingScale = theme('spacing');

      const createIconStyles = (value) => ({
        width: value,
        height: value,
        minWidth: value,
        minHeight: value,
        fontSize: value,
        lineHeight: value,
        'svg': {
          width: value,
          height: value
        }
      });

      // Standard spacing scale utilities
      const iconUtilities = Object.entries(spacingScale).reduce((acc, [key, value]) => {
        // Replace dots with underscores in class names to avoid CSS parsing errors
        const sanitizedKey = key.replace(/\./g, '_');
        acc[`.icon-size-${sanitizedKey}`] = createIconStyles(value);
        return acc;
      }, {});

      addUtilities(iconUtilities);

      // Arbitrary value support
      matchUtilities(
        {
          'icon-size': (value) => createIconStyles(value)
        },
        { values: spacingScale }
      );
    }
  ],
  // Performance optimizations
  corePlugins: {
    // Disable unused core plugins for better performance
    preflight: true,
  },
}