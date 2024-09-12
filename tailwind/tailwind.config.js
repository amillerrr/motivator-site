/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    '/templates/**/*.{gohtml, html}', // Tell Tailwind to scan these files for classes
  ],
  darkMode: 'media',
  theme: {
    container: {
      center: true,
      padding: '2rem',
      screens: {
        '2xl': '2000px', // Customizing the max width beyond 1536px
      },
    },
  },
  plugins: [],
}

