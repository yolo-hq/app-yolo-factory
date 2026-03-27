# MERGE SESSION

You are resolving merge conflicts for issue #{{ISSUE_NUM}} in a git worktree.

## Situation

A `git merge origin/{{TARGET_BRANCH}}` produced conflicts. Your job:

1. Identify all conflicting files
2. Read each conflicting file and understand both sides
3. Resolve every conflict — preserve intent from both sides
4. Stage resolved files with `git add`
5. Complete the merge with `git commit --no-edit`
6. Run feedback loops to verify the merge is clean

## Rules

- Do NOT discard changes from either side unless they are truly redundant
- If both sides modified the same function, merge the logic — don't pick one
- After resolving, the code must compile and tests must pass
- Do NOT modify files that are not conflicting

## Feedback loops

Run these after resolving all conflicts:

{{FEEDBACK_LOOPS}}

If any feedback loop fails, fix the issue and re-run until all pass.

## Output

If merge is resolved and all feedback loops pass, output: <promise>MERGED</promise>
If you cannot resolve the conflicts, output: <promise>CONFLICT</promise>
