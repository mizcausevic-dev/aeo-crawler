# Why We Built This

**aeo-crawler** came from a familiar enterprise pattern: important systems were technically functional but operationally under-explained. People could often find the inputs, but still struggle to form a clear next move.

The recurring pressure in this space showed up around weak semantic packaging, inconsistent structured data, and poor answer-system discoverability. In practice, that meant teams could collect logs, metrics, workflow state, documents, or events and still not have a good answer to the hardest questions: what is drifting, what matters first, who owns the next move, and what evidence supports that move? Once a system reaches that point, the problem is no longer only technical. It becomes operational.

That is why **aeo-crawler** was built the way it was. The repo is a deliberate attempt to model a real operating layer for growth, search, content, and analytics teams. It is not just trying to present data attractively or prove that a stack can be wired together. It is trying to show what happens when evidence, prioritization, and next-best action are treated as first-class product concerns.

Existing tools helped with adjacent workflows. SEO crawlers, analytics platforms, and schema validators covered storage, reporting, scanning, or execution in pieces. What they still missed was a review layer that connected technical content hygiene with answer readiness and citation potential. That left operators reconstructing the story manually at exactly the moment they needed clarity.

That shaped the design philosophy:

- **operator-first** so the riskiest or most time-sensitive signal is surfaced early
- **decision-legible** so the logic behind a recommendation can be understood by humans under pressure
- **review-friendly** so the repo supports discussion, governance, and iteration instead of hiding the reasoning
- **CI-native** so checks and narratives can live close to the build and change process

This repo also avoids trying to be a vague platform for everything. Its value comes from being opinionated about a real problem: BFS crawler for AEO Protocol v0.1 declaration graphs. Seed an origin, follow primary_source URIs, emit JSON Lines records of every fetch. Built on aeo-sdk-go. Concurrent, depth-limited, budget-capped, stdlib-only HTTP.

What comes next is practical. The roadmap is about deeper citation-gap detection, scheduled probes, and stronger semantic publishing workflows. That is the discipline this repo is trying to model in a practical, reviewable way.