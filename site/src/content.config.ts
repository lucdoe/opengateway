import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

// The pages live in ../docs, so they also read well on GitHub and change in
// the same PR as the code. docs/README.md is GitHub's folder index; the site
// has its own home page instead.
const home = 'site/src/home/index.mdx';

export const collections = {
  docs: defineCollection({
    loader: glob({
      base: '..',
      pattern: ['docs/*.md', '!docs/README.md', home],
      generateId: ({ entry }) => (entry === home ? 'index' : entry.replace(/^docs\//, '').replace(/\.md$/, '')),
    }),
    schema: docsSchema(),
  }),
};
