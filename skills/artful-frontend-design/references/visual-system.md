# Visual System

## Purpose

Use this file when converting the art direction into concrete design decisions and implementation rules.

## Palette Strategy

Build color around roles, not random swatches.

- Anchor: the main emotional color family
- Accent: the high-energy or high-importance contrast
- Neutrals: background, surface, border, and text scales
- Signal colors: success, warning, error, info

Practical rules:

- Let one family dominate.
- Use accent color sparingly so interactions still feel special.
- Vary saturation and lightness more than hue count.
- Make sure the palette still works when imagery is absent.

Useful palette structures:

- Ink + paper + one vivid accent
- Deep nocturne + electric accent + soft fog neutrals
- Warm stone + copper/red accent + charcoal text
- Clinical light neutrals + mineral blue/green + sparse coral alert

## Typography

Assign explicit jobs to type.

- Display face: voice, atmosphere, hero moments
- Text face: body copy, controls, data density
- Mono or condensed face: metadata, code, stats, navigation tags

Pairing rules:

- Combine one expressive face with one disciplined workhorse.
- Increase contrast through width, x-height, stress, or terminal style.
- If the display face is loud, make spacing and layout calmer.
- If the UI is dense, keep body typography plain and generous.

## Geometry And Shape Language

Shape decisions should echo the concept.

- Sharp rectangles signal precision or severity.
- Large radii feel approachable, toy-like, or soft.
- Capsules suggest momentum and controls.
- Irregular cuts or diagonal edges create tension and speed.
- Repeating corner rules create a recognizable component family.

Apply the same geometry to cards, buttons, media masks, dividers, and overlays.

## Surfaces And Depth

Choose a material model and stay loyal to it.

Options:

- Flat print-like planes
- Frosted/translucent layers
- Matte paper with grain
- Hard polished lacquer
- Dark luminous glass
- Industrial paneling with seams

Depth tools:

- Layered panels
- Border contrast
- Noise or grain
- Soft or hard shadows
- Specular highlights
- Inner glows

Use depth to separate information, not to decorate every box.

## Texture And Image Treatment

Texture works best when localized and intentional.

- Paper grain for editorial warmth
- Halftone or photocopy noise for cultural grit
- Subtle mesh or moire for digital atmosphere
- Soft vignettes for focus
- Repeated linework for technical feeling

Control image treatment:

- Desaturate or tint for integration
- Crop aggressively for drama
- Use blur or mask only when it supports hierarchy
- Keep image language consistent across the page

## Token Skeleton

Translate the visual system into tokens early.

```css
:root {
  --color-bg: #f4f0e8;
  --color-surface: #fffaf4;
  --color-ink: #161410;
  --color-accent: #b84c2a;
  --color-accent-2: #375e7a;

  --font-display: "Canela", "Iowan Old Style", serif;
  --font-body: "Suisse Intl", "Helvetica Neue", sans-serif;
  --font-mono: "IBM Plex Mono", monospace;

  --radius-sm: 8px;
  --radius-lg: 24px;
  --shadow-panel: 0 18px 60px rgba(22, 20, 16, 0.08);
  --texture-strength: 0.08;
}
```

Do not ship magic numbers everywhere if the design language can be tokenized once.

## Accessibility Guardrails

- Preserve text contrast even when the palette is atmospheric.
- Reserve extreme display styling for short text, not paragraphs.
- Keep focus states clearly visible within the chosen aesthetic.
- Check motion choices against reduced-motion needs.
- Make decorative layers ignore pointer and reading order when appropriate.
