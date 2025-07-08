# DEVLOG.md

This is the development log for the `goswaggen` project.
It captures ideas, experiments, technical investigations, decisions, and todos made during development.

---

## 📅 [YYYY-MM-DD] - [Short Title]

### 🎯 Goal
What are you trying to achieve or figure out today?

> Example: "Extract handler function names using Go AST"

---

### 🧪 Experiments / What I Tried
- Tried [method/tool/approach]
- Observations or results (✅ worked / ❌ didn’t work)
- Include code snippets or error messages if needed

> Example:
> - ✅ `go/ast` correctly detects `func (c *XController)` methods
> - ❌ Can't infer route path from Echo statically

---

### 🧠 Observations
What you discovered or realized — even if unrelated to the original goal.

> Example:
> - Echo does not expose registered routes at compile-time
> - Comments are stripped during AST traversal unless parsed in a specific mode

---

### ✅ Decisions
What you’ve decided to do based on your experiment. Include "for now" decisions too.

> Example:
> - Use `@Route` comment as fallback for unsupported routers
> - Only support exported methods attached to a `*Controller` struct

---

### 🔜 TODO / Next Steps
Any follow-up tasks or features that need to be tackled later.

> Example:
> - [ ] Add CLI flag for `--tag`
> - [ ] Test behavior when handler function returns different types
> - [ ] Investigate runtime reflection fallback

---

## 📅 [YYYY-MM-DD] - [Optional Subtitle]

_Repeat the above structure for each dev session or day._
