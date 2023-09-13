module.exports = {
	content: ['./site/src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
	theme: {
		extend: {
      colors: {
        pnk: '#f472b6',
        gld: '#FFD700',
      },
    },
	},
	plugins: [
    require('@tailwindcss/typography'),
  ],
};
