/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: '#2965ff',
        'brand-hover': '#5289ff',
        'brand-active': '#1848d9',
        'brand-light': '#e6f0ff',
        navy: '#0e2029',
        'navy-dark': '#071015',
        'navy-mid': '#0f2d3c',
        page: '#ffffff',
        surface: '#f5f5f5',
        line: '#e8e8e8',
        ink: '#0e2029',
        'ink-muted': 'rgba(0, 0, 0, 0.65)',
        'ink-light': 'rgba(255, 255, 255, 0.99)',
        'ink-muted-light': 'rgba(255, 255, 255, 0.77)',
      },
      fontFamily: {
        sans: ['"PT Sans Caption"', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
      },
      backgroundImage: {
        'hero-gradient': 'linear-gradient(102.61deg, #0f2d3c 20.68%, #2965ff 85.35%, #cccccc 101.01%)',
      },
      boxShadow: {
        btn: '0 2px 0 rgba(0, 0, 0, 0.045)',
      },
    },
  },
  plugins: [],
}
