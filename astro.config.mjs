import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import mdx from '@astrojs/mdx';

import react from "@astrojs/react";

// https://astro.build/config
export default defineConfig({
  root: "./site",
  srcDir: "./site/src",
  publicDir: "./site/public",
  integrations: [tailwind(), mdx(), react()]
});