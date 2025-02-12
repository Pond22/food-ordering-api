/** @type {import('tailwindcss').Config} */

const flowbite = require('flowbite-react/tailwind')

module.exports = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}', flowbite.content()],
  theme: {
    extend: {
      colors: {
        gold: '#D4AF37', // สีทองมาตรฐาน
        goldpre: '#DAA520',
        goldlight: '#E6C65C',
        // blackpremiem: '#2D232A',
        blackpremium: '#0A0A0A',
      },
    },
  },
  plugins: [flowbite.plugin()],
}
