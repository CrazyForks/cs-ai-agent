---
name: release-version
description: Create a new repository release using semantic version tags in the form `vx.y.z`, generate Chinese and English upgrade notes by comparing the new tag against the previous release tag, update `docs/zh/changelog.md` and `docs/en/changelog.md`, commit and push the `docs` submodule, then commit the parent repo's submodule pointer and push the new tag. Use when Codex needs to perform release preparation, changelog drafting, Git tag analysis, submodule release updates, or end-to-end version publishing.
---

# Release Version

## Overview

Create releases with a strict `vx.y.z` tag, produce bilingual changelog entries from actual Git history, and publish both the `docs` submodule update and the repository tag in one controlled workflow.

Run the workflow from the repository root. Read [references/changelog-style.md](references/changelog-style.md) before drafting the human-facing update notes.

## Workflow

1. Validate the requested version.
2. Inspect the repository and determine the comparison range.
3. Draft bilingual changelog entries from the actual diff.
4. Commit and push the `docs` submodule.
5. Commit the parent repository update if the submodule pointer changed.
6. Create and push the annotated tag.

Do not skip the repository inspection step. Release notes must come from the real diff between tags, not from guesswork.

## Validate The Version

- Accept only tags that match `^v\d+\.\d+\.\d+$`.
- Reject date-style tags such as `v20260414`.
- Confirm the target tag does not already exist locally or on any configured remote.
- Prefer the latest reachable semver tag as the previous release tag.
- If no earlier semver tag exists, fall back to the latest reachable tag of any format and state that fallback in the changelog drafting notes.

Use the helper script first:

```bash
python3 ~/.codex/skills/release-version/scripts/collect_release_context.py \
  --repo . \
  --tag v1.2.3
```

If the caller already specifies the previous tag, pass it explicitly:

```bash
python3 ~/.codex/skills/release-version/scripts/collect_release_context.py \
  --repo . \
  --tag v1.2.3 \
  --previous-tag v1.2.2
```

## Inspect The Repository

- Check `git status --short` in the parent repo.
- Check `git -C docs status --short` in the `docs` submodule.
- Read the JSON output of `collect_release_context.py`.
- Use the commit list, changed files, and insertions/deletions to decide what is user-visible.
- Prioritize behavior changes, new features, fixes, migrations, API changes, configuration changes, and documentation changes that matter to adopters.
- Ignore pure formatting churn unless it changes usage.

If the working tree contains unrelated changes that would be risky to include in the release, stop and ask the user before proceeding.

## Draft The Changelog

Update these files:

- `docs/zh/changelog.md`
- `docs/en/changelog.md`

Prepend a new entry using this exact structure:

```md
## ${tag} (${yyyy-MM-dd})

### 更新内容

${content}

### 发布地址

- Github: <https://github.com/huabeitech/cs-ai-agent/releases/tag/${tag}>
- Gitee: <https://gitee.com/mlogclub/bbs-go/releases/tag/${tag}>
```

For the English file, keep the same links and heading level, but translate the section heading and content naturally:

```md
## ${tag} (${yyyy-MM-dd})

### Updates

${content}

### Release Links

- Github: <https://github.com/huabeitech/cs-ai-agent/releases/tag/${tag}>
- Gitee: <https://gitee.com/mlogclub/bbs-go/releases/tag/${tag}>
```

Changelog writing rules:

- Write concise, user-facing summaries instead of raw commit subjects.
- Keep Chinese and English entries semantically aligned.
- Prefer 3-6 bullets unless the release is extremely small.
- Group related changes into a single bullet when that reads better.
- Mention compatibility-sensitive changes explicitly.
- If the comparison baseline is a non-semver fallback tag, note that in your private reasoning, not in the public changelog unless the user asks for it.

## Commit And Push The Docs Submodule

After editing the changelog files:

1. Run `git -C docs status --short`.
2. Review the diff with `git -C docs diff -- docs/zh/changelog.md docs/en/changelog.md` or the actual file paths present in the submodule.
3. Commit inside the `docs` submodule with a focused message such as `docs: update changelog for v1.2.3`.
4. Push the `docs` submodule commit to its remote branch.

Branch rule:

- If `docs` is on a local branch, push that branch.
- If `docs` is detached, push `HEAD` to `origin/main` unless the repository clearly uses another default branch.

## Commit The Parent Repository

If the `docs` submodule pointer changed in the parent repository, commit it before tagging. Otherwise the release tag will not reference the new changelog revision.

Recommended flow:

```bash
git status --short
git add docs
git commit -m "chore: update docs submodule for v1.2.3"
```

Only include unrelated parent-repo changes if the user explicitly wants them in the release commit.

## Create And Push The Tag

Create an annotated tag after the repository state is ready:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
```

Push the commit branch first if needed, then push the tag to every configured remote that should publish releases:

```bash
git push github HEAD
git push origin HEAD
git push github v1.2.3
git push origin v1.2.3
```

Adjust the branch name if `HEAD` is not tracking the intended release branch.

## Final Verification

- Confirm `git rev-parse v1.2.3^{tag}` succeeds.
- Confirm `git ls-remote --tags github v1.2.3` and `git ls-remote --tags origin v1.2.3` show the new tag.
- Confirm the `docs` submodule remote contains the changelog commit.
- Summarize the previous tag used for comparison, the files updated, the commit hashes created, and the remotes pushed.
