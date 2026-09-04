---
name: feynman
description: Feynman technique. Explain a topic in plain words, hunt your own gaps, fill them from source, repeat until clean. If you cannot explain it simply, you do not understand it yet. Run with /feynman <topic>.
disable-model-invocation: true
---

The rule: if you cannot explain something in a simple way, you do not fully understand it yet. This skill runs that rule as a loop until the explanation holds.

## Unslop first

Before any response, run the /unslop skill on your output and apply its fixes. Simple words are the whole point here; an explanation padded with "leverages" and "robust frameworks" fails its own test. Unslop, then send.

## Invocation

Parse the topic from the invocation. Three forms:

- `/feynman <topic>`: you explain the topic.
- `/feynman this`: explain whatever the conversation was just about.
- `/feynman me <topic>`: role flip. The user explains, you listen and grade.

No topic? Ask once, wait for the answer.

## Grounding

Never explain from memory alone.

- Topic touches this repo: read the actual code first. Every claim about it cites `file:line`.
- Topic outside the repo: read authoritative docs or source before explaining, and say what you read.
- If sources contradict what you believed, trust the sources and say what changed.

## The loop

One pass has three moves. Repeat until the exit gate passes.

1. **Explain** it to a smart 12-year-old who shares none of your vocabulary. Short sentences. Concrete analogies. Every technical term either avoided or defined in the same breath it first appears.
2. **Audit** your own explanation for gaps. Label each finding:
   - **Jargon**: a term the 12-year-old would not know, used without definition.
   - **Hand-wave**: "it just handles that", "under the hood". No mechanism named.
   - **Circularity**: X explained by Y, Y explained by X.
   - **Name-dropping**: leaning on a famous name instead of saying what it does.
3. **Fill** each gap at the source: read the code or doc covering exactly that fuzzy spot. Then re-explain, tighter than last time.

Exit gate (all must hold):

- Zero audit findings on the latest pass. Not fewer than last time. Zero.
- Every step names a mechanism, not a reputation.
- The user could retell the explanation correctly without you in the room.

Report the pass count. "Clean in 3 passes" means something; "clean" alone does not.

## Grading mode (`/feynman me <topic>`)

You become the listener.

- Let the user finish a full attempt before interrupting.
- Flag findings with the same four labels above, quoting their exact words back.
- After flags, drill one question at a time: "why does that work?" until they hit bedrock or hit a wall. A wall is the method working, never mock it.
- Point every gap to the file or doc that closes it.
- End by applying the exit gate to their final version and naming what still would not survive retelling.