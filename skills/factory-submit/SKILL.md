---
name: factory:submit
description: Submit a PRD to YOLO Factory for autonomous execution
---

Extract a PRD from the current conversation and submit it to Factory for planning and execution.

## Steps

1. **Identify the project.** Ask: "Which project is this for?" If the user's cwd matches a Factory project path, suggest it.

2. **Extract PRD from conversation.** Look for:
   - Title: short description of the feature/change
   - Body: what to build and why (include context, constraints, boundaries)
   - Acceptance criteria: specific, testable conditions (minimum 3)
   - Design decisions: key choices made during /grill-me (if applicable)

3. **Show the structured PRD** for review:
   ```
   Project: plugin-webhooks
   Title: Webhook retry with exponential backoff
   
   Body:
   Add retry logic for failed webhook deliveries...
   
   Acceptance Criteria:
   - Retries up to configured max
   - Exponential backoff with jitter between retries
   - Dead letter queue after max retries
   
   Design Decisions:
   - RetryPolicy in core, not plugin
   - Exponential backoff prevents thundering herd
   ```

4. **Ask for confirmation:** "Submit to Factory? [Y/n]"

5. **Submit via MCP** (if Factory MCP is available):
   Call `factory_submit_prd` with the structured data.
   
   If MCP is not available, output the CLI command instead:
   ```bash
   yf prd:submit --project plugin-webhooks --title "..." --body "..." --criteria "..." --criteria "..."
   ```

6. **Trigger planning:** After submission, suggest:
   "PRD submitted. Run `yf prd:execute <id>` to start planning, or I can do it via MCP."

7. **Show task breakdown** (if execution triggered):
   After Planner runs, show the generated tasks and ask:
   "Factory created N tasks. Review them with `yf task:list --prd <id>`. Execute? [Y/n]"

## Rules
- Do NOT create GitHub issues -- Factory PRD entity is the source of truth
- Minimum 3 acceptance criteria (more = higher success rate)
- Criteria must be verifiable by reading code (the review agent checks with file:line evidence)
- If /grill-me was used earlier, extract design decisions from it
- If /write-prd was used earlier, extract the PRD body from it
- Keep body concise -- agents read it, long specs waste tokens
