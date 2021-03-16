module.exports = {
  plugins: [
    require('tailwindcss')(require('path').resolve(__dirname, 'tailwind.config.js')),
    ...(
      // eslint-disable-next-line no-process-env
      process.env.HUGO_ENVIRONMENT === 'production'
        ? [ require('autoprefixer') ]
        : []
    ),
  ]
};
