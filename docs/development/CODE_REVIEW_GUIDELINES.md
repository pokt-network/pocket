# Table of Contents <!-- omit in toc -->

- [Code Quality](#code-quality)
- [Code Reviews](#code-reviews)
- [Code Review Guidelines](#code-review-guidelines)
    - [Expectations](#expectations)
- [Best Practices](#best-practices)
    - [Reviewing](#reviewing)
    - [Starting a Review](#starting-a-review)
    - [Finishing a Review](#finishing-a-review)
    - [Merging](#merging)

# Code Quality

"Code Quality" can be a vague concept, as it usually addresses what is more the "art" side (vs. the "science" side) of software development.
There are often several *technically correct* ways to address a problem -- that is, the correct answer or behavior is produced.
Selecting the "*best*" solution is often a matter of style.
Sometimes the best solution is one that fits the surrounding code in the most cohesive way.
Terms like "maintainability" or "readability" are used; these address the ability of other contributors to understand and improve the code.

Unlike correctness or performance concepts, there's no single metric or mathematical solution that can be optimized to achieve better code quality.
Thus, we rely on human judgement.

In this document, we outline a framework to guide that human judgement towards -- collectively -- "better code".

For decades, the IETF (Internet Engineering Task Force) has used the motto "rough consensus and running code".
This motto encapsulates ("running code") that developers' core output is still software: if there is no code that runs and produces correct results, we have nothing.
It also encapsulates ("rough consensus") that we may not always precisely agree and that's okay.

## Code Reviews

One tactic often employed to produce "maintainable" or "more-readable" code is a "code review".
These can take many shapes and forms and often have goals beyond simply "code quality".

Broadly speaking, code reviews involve developers looking at some proposed new code (or code changes).
This is *usually* developers other than the author (although a "self-review" can also be employed).
In many projects, such attention is a scarce commodity: most programmers would rather write code than read someone else's.

## Code Review Guidelines

All participants must adhere to the overall idea that the review is an attempt to achieve "better" code.
This is a vague statement on purpose.
Participants must be cautious in their criticism and generous with praise.
Participants must remember the scarcity of another developer's attention.

### Expectations

**Pull Request Authors:**
The author is responsible for tracking up-to-date next actions for a Pull Request to progress towards being merged.

**Reviewers:**
Reviewers should prefer engaging in code review over starting new work (i.e. taking planned work items that haven't been started yet).

Reviewers (and prospective reviewers) are encouraged to engage in reviews of codebases outside the projects and technologies they use on a day-to-day basis (but not expected to provide an approving review).

## Best Practices

### Reviewing

TL;DR Use SOS, TECHDEBT(XXX), TECHDEBT, NIT if you think it'll help the author get context on next steps.

When you leave reviews, if it is unclear from the reviewer to the author, please make sure to mention if:

- NIT: Comment is a nitpick
- TECHDEBT: Comment should have a TECHDEBT comment w/o a ticket
- TECHDEBT(XXX): Comment should have a TECHDEBT comment but imporant enough that it requires a ticket to track the work in the near future
- SOS: Show Stopper. You feel strongly it needs to be addressed now

Only use the above if you think normal comments / TODOs are not self explanatory enough.

### Starting a Review

Consider if it could be broken into smaller Pull Requests. If it is clear that it can be, summarize your thinking on how in your Review.

If the commits be (re)organized (i.e. reordered and/or amended) such that there is a commit at which the tests are passing prior to the conclusion of the main change, that's a signal that there's likely a logic split which can be made at that point in such a (re)arrangement.

Use the following guidelines to evaluate whether a Pull Request should be approved:

- *Every* Pull Request should have tests or justification otherwise; esp. bug fixes.
- *Every* Pull Request should have at least 1 approval from a team member internal or external to the project. Exceptions made by repository maintainer(s), as necessary, on a case-by-case basis.

During review, submit feedback using line comments (A); prefer "Add\[ing a\] single comment" over "Start[ing] a review" (B).

*(Once a review has been started, the option to add single comments is removed. Preferring single comments allows feedback to happen even in the event of an interrupted review.)*

### A

![github_line_comment.png](./assets/github_line_comment.png)

### B

![line_comment_dialog_start.png](./assets/line_comment_dialog_start.png)

**Referencing Issues Across Repositories:**
When referencing issues from one repository, in another's Issues and Pull Requests, GitHub supports automatic links in markdown using the following format: `<org name>/<repo name>#<issue or PR number>`.`

### Finishing a Review

Write a summary of your observations and consider including positive remarks in addition to any constructive criticism.

If you observe a deviation from these practices or another reason that this change should not be merged, select "request changes" and include a summary of the observation(s), as well as any practice(s) you find them to be in conflict with, in the review body (C).

If you don't feel comfortable giving approval or requesting changes but want to share a summary or observations of larger patterns in the codebase or the company, select "Comment" and submit your review (C).

Confirm that all items in the "required checklist" are checked or not applicable.

If you believe the Pull Request looks good to merge, select "Approve" and submit your review (C).

### C

![submit_review.png](./assets/submit_review.png)

### Merging

1. Utilize the `Squash & Merge` feature (maintain a clean history)
2. Copy the `Github PR Description` into the commit message (add sufficient detail)

**Authors are core members or regular external contributors**

- Core member needs to approve PR
- Author should merge in the PR themselves (following instructions above)

**Authors are non-regular external contributors:**

- Core member needs to approve PR
- Core member can merge PR on behalf of contributor  (following instructions above)

**Pull Request Authors:**
Specify whether or not the Pull Request can be merged by the last approving reviewer by use of (or lack thereof) the "Reviewer can merge" label. If no such label exists for a given project, create it.

**Reviewers:**
If you would give the last approving review **and** the Pull Request has the "Reviewer can merge" label, **merge**.
