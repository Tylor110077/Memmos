---
name: artful-frontend-design
description: Design visually distinctive, art-directed, production-minded web frontends that avoid generic template aesthetics. Use when Codex is asked to design or restyle a landing page, dashboard, marketing site, product UI, portfolio, campaign page, or component system where visual quality matters, including theme creation, palette building, typography pairing, composition, imagery direction, texture, motion, or translating a concept or brand into a functional frontend.
---

# Artful Frontend Design

## Overview

Create a strong visual direction before writing UI code. Build frontends that feel authored rather than templated while preserving usability, hierarchy, responsiveness, and implementation realism.

## Workflow

1. Extract the product goal, audience, emotional target, content type, and functional constraints.
2. Write a design thesis in 2-4 sentences. Name the visual tension clearly, such as "editorial luxury with operational precision" or "retro-futurist play anchored by sober navigation."
3. Choose one primary art direction and at most one supporting influence. Reject vague blends.
4. Build the visual system: palette, typography, spacing, geometry, surfaces, imagery, iconography, and motion.
5. Compose the page around one memorable move: hero treatment, navigation frame, card geometry, asymmetry, light treatment, or scroll reveal.
6. Implement with reusable tokens and consistent component rules.
7. Critique the result for originality, clarity, accessibility, responsiveness, and coherence. Refine weak decisions instead of adding more decoration.

## Non-Generic Standard

- Avoid default startup tropes unless the product explicitly calls for them: centered headline over pastel gradient, feature-card triplets, glassmorphism on white, random blobs, interchangeable dashboards.
- Prefer one strong visual idea carried through the whole interface over many unrelated flourishes.
- Use contrast beyond color: scale, density, silence, rhythm, edge hardness, saturation pockets, and material shifts.
- Make decorative moves earn their place by reinforcing brand, narrative, or hierarchy.
- Preserve task clarity. Beauty does not excuse vague navigation, weak contrast, or broken flows.

## Deliverable Contract

When the user asks for design help, produce all of the following unless the request is narrower:

1. A concise design thesis.
2. A visual direction summary covering palette, type, materials, imagery, motion, and layout attitude.
3. A component treatment summary for navigation, hero, sections/cards, forms, data display, and calls to action.
4. Implementation notes for tokens, responsiveness, and accessibility constraints.
5. The actual code changes. Do not stop at moodboard language unless the user asked only for ideation.

## Reference Loading

Read only the files needed for the current task.

- For theme invention, mood, and source-material selection, read [references/art-direction.md](references/art-direction.md).
- For palette, type, surface, depth, and texture decisions, read [references/visual-system.md](references/visual-system.md).
- For layout, component styling, and interaction rhythm, read [references/interface-composition.md](references/interface-composition.md).
- For critique and final polish, read [references/evaluation.md](references/evaluation.md).

## Working Rules

- Start from the product or content truth. A fintech tool, music festival, museum archive, and AI notebook should not share the same visual language.
- Preserve an existing design system when the codebase already has one. Push the work forward without breaking brand constraints unless the user explicitly asks for a rebrand.
- Prefer expressive typography and deliberate spacing systems over decorative clutter.
- Use color intentionally: one dominant family, one accent strategy, one neutral system.
- Introduce motion sparingly and with purpose. Favor staged reveals, spatial transitions, and emphasis cues over constant animation.
- If imagery or illustration is needed, specify the exact role it should play: atmosphere, credibility, narrative, texture, or instructional clarity.
- Design for desktop and mobile from the start. A striking desktop composition that collapses on mobile is incomplete.
- If the frontend still feels generic, change the concept, not just the polish.

## Request Patterns

- New page or app: define the thesis, then design the system before implementing.
- Restyle existing UI: identify what currently feels generic or incoherent, keep the useful structure, and replace weak visual assumptions.
- Component request: infer the surrounding art direction so the component feels native to a larger world.
- Design critique: list the strongest and weakest visual decisions first, then propose concrete revisions.
