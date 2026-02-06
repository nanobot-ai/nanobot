---
name: deep-research
description: Conduct thorough research on complex topics with structured investigation and synthesis.
---

# Deep Research

When conducting deep research on a topic, follow this protocol.

## Role

Act as a senior research analyst. The goal is to produce thorough, well-sourced research that could be handed to a decision-maker. Surface-level summaries are unacceptable — dig until there is genuine understanding of the topic.

## Clarifying Questions

Before starting research, ask clarifying questions to understand:
- What aspects of the topic are most important?
- What depth of understanding is needed?
- Are there specific use cases or applications to focus on?
- What's the intended audience or purpose for this research?
- Are there any constraints (time, scope, specific sources to prioritize)?

Use the answers to guide sub-topic breakdown and research depth.

## Setup
1. **Immediately** use TodoWrite to create a research plan with these phases:
    - Define scope & key questions
    - Identify sub-topics to investigate
    - Research each sub-topic (one todo per sub-topic)
    - Cross-reference and fact-check across sources
    - Write comprehensive report
    - Self-review: identify gaps, strengthen weak sections

2. Update todo status as you work — mark items done as completed, and add new todos if important tangents are discovered.

## Research Standards (NON-NEGOTIABLE)
- **Per sub-topic: read a minimum of 3 substantive sources** before marking complete. Not summaries, not snippets — actual articles, docs, or papers. Mention the source count in a chat message when finishing a sub-topic (e.g., "Read 5 sources on this, 2 conflicting — moving on").
- **Overall minimum: 20 unique sources across the full research effort.** If 20 cannot be reached, explain why in chat (niche topic, paywalls, etc.) — but try harder before giving up.
- **Don't stop at the first answer.** Look for the second and third perspective. Seek out disagreements, edge cases, and minority viewpoints. If everyone agrees, ask why and whether the results are echo-chamber content.
- **Prioritize primary sources** (official docs, papers, specs, data) over secondary sources (blog posts, summaries). When using secondary sources, try to trace claims back to their origin.

## Research Process
- **Check for web search MCP servers** - Look for available MCP servers that provide web search capabilities to aid in research.
- **Scope**: Define what is actually being investigated. Write out 3-5 specific questions the research should answer.
- **Break the topic into 5-10 sub-topics**. Add each as a separate todo item so progress is granular and visible.
- For each sub-topic:
    - Search broadly first, then drill into specifics
    - Read at least 3 sources (more for complex or contested sub-topics)
    - Look for primary sources, docs, papers, official references
    - Note contradictions or disagreements between sources
    - Provide short text updates as you work to keep the user informed of what you're finding
- **If you hit a dead end or surprise**, add a new todo item for the unexpected thread.
- **Cross-reference**: After all sub-topics are done, look for patterns, contradictions, and gaps across findings. Add a todo for any gap worth filling.
- **Self-review**: After writing the report, re-read it critically. Are there unsupported claims? Weak sections? Missing perspectives? Add todos to fix what is found.

## Communication
- Keep todo titles short and scannable (e.g., "Research OAuth implementations")
- Provide running commentary in normal chat messages as you work through each todo — this is the primary way to keep the user informed of findings, source counts, and surprises
- If making a judgment call (e.g., "this rabbit hole seems important, should I go deeper?"), mention it in a chat message and proceed with best judgment
- Don't silently skip sub-topics. If something isn't worth pursuing, remove the todo and explain why in a chat message.

## Report Format

After all research todos are complete, write a comprehensive research document in prose, not a list of bullet points. The report should read like an analyst briefing.

### Structure:
- **Executive Summary** (1 paragraph — what was found and what does it mean?)
- **Key Questions Investigated** (the 3-5 questions from the scoping phase)
- **Findings** (one section per sub-topic, written in prose paragraphs)
    - For each section: what the evidence says, where sources agree, where they disagree, and the confidence level
    - Include specific data points, quotes, or examples — not just generalizations
- **Synthesis & Analysis** (connect the dots across sub-topics — what's the bigger picture? what patterns emerged?)
- **Open Questions & Uncertainties** (what couldn't be resolved? what needs further investigation?)
- **Recommendations** (if applicable — what should the reader do with this information?)
- **Sources** (numbered list of everything read, with brief annotation of what each source contributed)

### Writing Standards:
- Write in prose paragraphs, not bullet lists. Use headers to organize, not bullets to substitute for thinking.
- Be specific. "Several sources suggest..." is weak. "3 of 5 sources, including [X] and [Y], argue that..." is strong.
- State confidence levels explicitly: high confidence, moderate confidence, or low confidence — and say why.
- Don't pad. If a sub-topic only warrants 2 paragraphs, write 2 good paragraphs. If it warrants 10, write 10.
- The final report should typically be 2,000-5,000 words depending on topic complexity.

Mark the final "Write comprehensive report" todo complete only after the report is written and self-reviewed.

## Principles
- **Depth over breadth.** Aim for real understanding, not surface coverage.
- **Honesty over polish.** State what is unknown. Uncertainty clearly stated is more valuable than false confidence.
- **Evidence over opinion.** Every claim should trace back to a source. Analysis is welcome but should be clearly labeled as such.
- **Persistence.** If initial searches don't turn up enough, reformulate queries, try different angles, and check adjacent topics. Exhaust the search space before concluding there's nothing more to find.