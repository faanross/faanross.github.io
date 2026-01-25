# faanross-svelte

Personal website built with SvelteKit.

## Article Pages

**ALWAYS use `ArticleLayout` component for blog articles in `/routes/claude/`.**

### Usage

```svelte
<script lang="ts">
  import ArticleLayout from '$lib/components/ArticleLayout.svelte';
</script>

<ArticleLayout
  title="Article Title Here"
  date="2025-01-11"
  description="Meta description for SEO."
>
  <p>Article content goes here...</p>

  <h2>Section heading</h2>
  <p>More content...</p>

  <pre><code>{`code blocks work automatically`}</code></pre>
</ArticleLayout>
```

### Props

| Prop | Required | Default | Description |
|------|----------|---------|-------------|
| `title` | Yes | - | Article title (also used in `<title>` tag) |
| `date` | Yes | - | Publication date (YYYY-MM-DD) |
| `description` | Yes | - | Meta description for SEO |
| `backLink` | No | `/claude` | URL for back button |
| `backText` | No | `Back to Claude` | Text for back button |

### What ArticleLayout Provides

- Scroll progress indicator
- Back navigation with proper styling
- Date display
- Title with proper typography
- Copy buttons on all code blocks (auto-added)
- Proper spacing for copy button (44px top padding)
- Styled: paragraphs, headings, lists, tables, images, blockquotes, code, links
- Responsive design
- Fade-in animations

### Do NOT Use ArticleLayout For

- Index pages (`/claude/+page.svelte`)
- Dashboard/app pages (`/claude/memory/*`)
- Pages with custom layouts

## Lesson Pages (Courses)

Use `LessonLayout` component for course lesson pages in `/routes/courses/`.

## Components

| Component | Purpose |
|-----------|---------|
| `ArticleLayout.svelte` | Blog article wrapper with all styling |
| `LessonLayout.svelte` | Course lesson wrapper |
| `ScrollProgress.svelte` | Reading progress bar |
| `Nav.svelte` | Site navigation |
| `Background.svelte` | Animated background |

## Code Style

- Use Svelte 5 runes (`$state`, `$props`, etc.)
- TypeScript for all components
- Tailwind NOT used - custom CSS with CSS variables
- CSS variables defined in `src/lib/styles/global.css`
