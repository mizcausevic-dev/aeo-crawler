# Changelog

All notable changes to this project are documented here.

## [1.0.0] - 2026-05-12

### Released
- Shipped **aeo-crawler** as a public artifact for teams dealing with answer-engine visibility.
- Packaged the current implementation, documentation, validation flow, and proof surfaces into a repo that can be reviewed by technical and operating stakeholders.
- Clarified the core problem the project is addressing: weak semantic packaging, inconsistent structured data, and poor answer-system discoverability.

### Why this mattered
- Existing approaches in SEO crawlers, analytics platforms, and schema validators were useful for parts of the workflow.
- They still left out a review layer that connected technical content hygiene with answer readiness and citation potential.
- This release made the repo read like an operational capability rather than a narrow technical demo.

## [0.1.0] - 2026-03-20

### Shipped
- Cut the first coherent internal version of **aeo-crawler** with stable domain objects, review surfaces, and decision outputs.
- Established the first reviewable version of the architecture described as: BFS crawler for AEO Protocol v0.1 declaration graphs. Seed an origin, follow primary_source URIs, emit JSON Lines records of every fetch. Built on aeo-sdk-go. Concurrent, depth-limited, budget-capped, stdlib-only HTTP.
- Focused the repo around actionability instead of passive reporting.

## [Prototype] - 2025-04-08

### Built
- Built the first runnable prototype for the repo's main workflow and decision model.
- Validated the concept against pressure points such as answer-engine discoverability gaps, thin structured data, and inconsistent entity linking.
- Used the prototype phase to test whether the project could drive action, not just present information.

## [Design Phase] - 2024-11-08

### Designed
- Defined the system around operator-first and decision-legible outputs.
- Chose interfaces and examples that made sense for growth, search, content, and analytics teams.
- Avoided reducing the project to a generic dashboard, CRUD app, or fashionable wrapper around the stack.

## [Idea Origin] - 2024-03-08

### Observed
- The original idea surfaced while looking at how teams were handling weak semantic packaging, inconsistent structured data, and poor answer-system discoverability.
- The recurring pattern was that teams had data and tools, but still lacked a usable operating layer for the hardest decisions.

## [Background Signals] - 2022-08-09

### Context
- Earlier platform, governance, and operator-tooling work made one pattern hard to ignore: the systems that create the most drag are often the ones with partial controls and weak operational coherence, not the ones with no controls at all.
- That pattern shaped the thinking behind this repo well before the public version existed.