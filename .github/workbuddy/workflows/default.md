---
name: default
description: Default 2-agent lifecycle for any workbuddy-tracked issue
trigger:
  issue_label: "workbuddy"
max_retries: 3
---

## Default Workflow

Two-agent state machine applied to every issue labeled `workbuddy`.

```yaml
states:
  developing:
    enter_label: "status:developing"
    agent: dev-agent
    transitions:
      - to: reviewing
        when: labeled "status:reviewing"
      - to: blocked
        when: labeled "status:blocked"

  reviewing:
    enter_label: "status:reviewing"
    agent: review-agent
    transitions:
      - to: done
        when: labeled "status:done"
      - to: developing
        when: labeled "status:developing"

  blocked:
    enter_label: "status:blocked"
    transitions:
      - to: developing
        when: labeled "status:developing"

  done:
    enter_label: "status:done"
    action: close_issue

  failed:
    enter_label: "status:failed"
```
