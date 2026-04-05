# Learning Approach

This document describes the pedagogical model used in this project. It's versioned alongside the code so it can be iterated on or rolled back like anything else.

## Why this approach

The original approach was guided transcription: Claude explains a concept and provides the exact code to write, the learner types it out and verifies it compiles. This works well for complete novices — when you have no mental schema yet, studying worked examples reduces cognitive load and lets you focus on the concept rather than the problem-solving.

But it stops working once you've built enough schema. At that point, transcribing code you already understand produces low cognitive load _and_ low learning — the worst of both worlds. This is the **expertise reversal effect**, well-documented in Cognitive Load Theory research.

The approach below is the upgrade: keep worked examples for genuinely new concepts, replace transcription with exercises for everything else.

## The three-mode framework

### Mode 1 — New concept (first exposure)

1. Explain the concept, with analogies to known languages where useful
2. Show one complete worked example with commentary
3. Give an exercise: implement the next case yourself
4. Tests / compiler are the oracle — not the teacher

**Important:** A new third-party library or API always counts as a new concept, even if the surrounding Go pattern is familiar. Never give a spec that requires calling an unfamiliar library without first explaining what it does and showing how to use it. Sending someone to read docs they've never seen is not a spec — it's a scavenger hunt.

### Mode 2 — Applying a known pattern (2nd+ case)

1. Give a spec only: what it should do, inputs, outputs, error cases
2. Implement from scratch
3. Run tests to verify — no line-by-line checking needed
4. Debrief after: what worked, what was surprising, what could be cleaner

### Mode 3 — Mechanical steps (no learning value in guessing)

- CLI commands, config file keys, dependency versions, file paths
- Just provide the exact thing — no exercise, no friction

## The test-as-oracle principle

Frame exercises as "implement X so that test Y passes" wherever possible. The compiler and test runner give immediate, unambiguous feedback. The teacher's job is explaining outcomes, not checking correctness.

## One step at a time

Regardless of mode, never change more than one file before verifying it compiles and tests pass. This keeps failures easy to isolate.

## Rollback

If the exercise approach produces too much frustration or blocks progress, revert to guided transcription for that chapter and reassess. The goal is learning, not adherence to a method.

## Further reading

- [Worked-example effect — Wikipedia](https://en.wikipedia.org/wiki/Worked-example_effect)
- [Self-Explanation Effect of Cognitive Load Theory in Teaching Basic Programming (2024)](https://jise.org/Volume35/n3/JISE2024v35n3pp303-312.pdf)
- [Enhancing Teaching Strategies through Cognitive Load Theory (2024)](https://www.mdpi.com/2227-7102/14/8/813)
- [Is More Active Always Better for Teaching Introductory Programming?](https://pages.cs.wisc.edu/~gerald/papers/IsMoreActiveAlwaysBetter.pdf)
