module.exports = {
  content: ["./core/views/**/*.tmpl"],
  future: {
    hoverOnlyWhenSupported: true,
  },
  theme: {
    extend: {
      keyframes: {
        "fade-in": {
          '0%': { opacity: '0%' },
          '100%': { opacity: '100%' },
        }
      },
      animation: {
        "fade-in": 'fade-in 0.075s ease-in-out',
      }
    }
  },
  plugins: [],
}
