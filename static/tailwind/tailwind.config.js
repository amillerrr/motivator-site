/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './templates/**/*.{gohtml,html}', // Tell Tailwind to scan these files for classes
    './static/**/*.{css,js}',      // If you have custom JavaScript, include this as well
  ],
  theme: {
    extend: { // Customize the default theme if needed
      colors: {
        brand: '#1c3d5a',
      },
      spacing: {
        '128': '32rem',
      },
    },
  },
  plugins: [],
}

