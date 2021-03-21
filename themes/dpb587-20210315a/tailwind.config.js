const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  purge: {
    enabled: process.env.HUGO_ENVIRONMENT === 'production',
    content: [
      './hugo_stats.json',
      './layouts/**/*.html',
    ],
    extractors: [
      {
        extractor: (content) => {
          let els = JSON.parse(content).htmlElements;

          return els.tags.concat(els.classes, els.ids);
        },
        extensions: ['json']
      },
    ],
    mode: 'all',
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
  theme: {
    extend: {
      typography: (theme) => ({
        DEFAULT: {
          css: {
            color: theme('colors.gray.800'),
            a: {
              color: theme('colors.yellow.600'),
              '&:hover': {
                color: '#2c5282',
              },
            },
            ol: {
              textAlign: 'justify',
            },
            p: {
              textAlign: 'justify',
            },
            ul: {
              textAlign: 'justify',
            },
          },
        },
      }),
    },
    // fontFamily: {
    //   sans: [
    //     'Quicksand',
    //     ...defaultTheme.fontFamily.sans,
    //   ],
    // },
  },
};
