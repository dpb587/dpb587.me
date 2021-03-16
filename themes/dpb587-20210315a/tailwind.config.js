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
    fontFamily: {
      sans: [
        'Quicksand',
        ...defaultTheme.fontFamily.sans,
      ],
    },
  },
};
