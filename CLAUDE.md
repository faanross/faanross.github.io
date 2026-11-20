# faanross-svelte

Personal site for faanross.com, built with SvelteKit (static adapter, deployed to
GitHub Pages on every push to `main`).

The site is intentionally tiny: one landing page that points to aionsec.ai, plus the
archive for the AntiSyphon course "Building a C2 Framework in Go", which paid students
still reach at `/courses/course01`. Do not remove or rename that route while the course
is still in AntiSyphon's catalogue.

## Structure

| Path | Purpose |
|------|---------|
| `src/routes/+page.svelte` | Landing page: short bio, socials, course-archive note |
| `src/routes/+error.svelte` | 404/error page (old URLs land here) |
| `src/routes/courses/course01/+page.svelte` | Course archive index (table of contents) |
| `src/routes/courses/course01/*/+page.md` | Lessons (mdsvex, `layout: course01` in frontmatter) |
| `src/lib/layouts/course01.svelte` | Lesson layout: breadcrumbs, prose and code styling |
| `src/lib/components/EyeField.svelte` | Full-screen canvas of ASCII eyes |
| `src/lib/eyes/field.ts` | The eye engine (layout, gaze, blinks, click wave) |
| `src/lib/components/ScrollProgress.svelte` | Section dots on lesson pages |
| `src/lib/utils/addCopyButtons.ts` | Adds copy buttons to every `<pre>` |
| `src/lib/styles/global.css` | Tokens, Geist Mono `@font-face`, base styles |

## The eye field

- `<EyeField />` renders behind the page. Mark text blocks with `data-eye-avoid`: the
  big eye keeps clear of them; every other eye may sit behind them, dimmed to ~20%
  inside a soft invisible frame (`dimRects`/`dimAlpha` on the scene).
- The homepage cast is `HOME_CAST` in `src/lib/eyes/field.ts` (design ids from the
  catalogue in `src/lib/eyes/designs.ts`; `'C02:aion'` pins a palette). The big eye has a fixed anchor; the others
  are randomised per load and keep their places on resize.
- `visible={false}` keeps the eyes hidden; flipping it to true fades them in, big
  eye first, the rest staggered (the homepage does this after its typing intro).
- A press anywhere (except links/buttons) calls `poke(x, y)`: an ASCII ping expands from
  the point and each eye turns toward it as the wave reaches it. Eyes close to the point
  flinch (blink).
- Eye kinds: `sigil` (the large AIONSEC eye with ring and dial), `owl`, `serpent` (slit
  pupil), `hooded`, `goat` (bar pupil), `almond`, plus tiny glyph eyes like `(o)` that
  often come in pairs. Recipes live in `variantSpec()` and `layout()`.
- Glyphs are blitted from per-size/per-colour sprite atlases (much cheaper than
  `fillText`); a full desktop frame is ~2.5k glyphs at ~2 ms.
- Respects `prefers-reduced-motion` (static eyes; a click still turns them).

## Style

- Dark only. Colours are CSS variables in `global.css`: `--bg`, `--fg`, `--fg-dim`,
  `--fg-faint`, `--line`, `--purple` (#bd93f9), `--yellow` (#f1fa8c).
- Everything is set in Geist Mono (self-hosted in `static/fonts`, OFL licence alongside).
- Svelte 5 runes, TypeScript, no Tailwind.
