// https://github.com/tailwindlabs/tailwindcss-typography/blob/a86e6015694c3435ff6cef84f3dd61b81adf26e1/src/styles.js#L3-L9
const round = (num) =>
  num
    .toFixed(7)
    .replace(/(\.[0-9]+?)0+$/, '$1')
    .replace(/\.0$/, '')

// const rem = (px) => `${round(px / 16)}rem`
const em = (px, base) => `${round(px / base)}em`

/** @type {import('tailwindcss').Config} */
export default {
  theme: {
    extend: {
      // height: {
      //   'screen/80': '80vh',
      // },
      // rotate: {
      //   '135': '135deg',
      // },
      // borderRadius: {
      //   xs: '0.075rem', // 1px
      // },
      // scale: {
      //   '103': '1.03',
      // },
      typography: {
        DEFAULT: {
          css: {
            'p code': {
              fontWeight: 400,
            },
            'code::before': {
              content: '""'
            },
            'code::after': {
              content: '""'
            },
            h1: {
              fontSize: em(24, 14),
              fontWeight: '600',
              marginTop: em(48, 30),
            },
            h2: {
              fontWeight: '600',
            },
            'strong > a': {
              fontWeight: '600',
            },
          },
        },
      },
    },
  },
}
