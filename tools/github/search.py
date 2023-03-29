import os
import re
from datetime import datetime, timedelta

import click
from github import Github

iso_8601_format = "%Y-%m-%d"
now = datetime.now()
usage_text = """
usage: search [reviews|issues] [OPTIONS] <search text> 

OPTIONS:
    -r, --repo <owner/name>         GitHub repository to search, specified as `owner/name` (default: "pokt-network/pocket")
    -s, --since <date ISO 8601>     Earliest issue/review commment date to consider (default: 6 months ago)
    -p, --pulls <PR#>[, ...]        Specific pull request(s) to consider (default: all repo PRs)
    -i, --issues <issue#>[, ...]    Specific issue(s) to consider (default: all repo issues)
    -u, --users <user>[, ...]       Specific user(s) to consider (default: all participating users)
"""

access_token_key = "GITHUB_ACCESS_TOKEN"
default_repo = "pokt-network/pocket"
default_since = (now - timedelta(weeks=24)).strftime(iso_8601_format)  # 6 months ago


@click.command()
@click.option("--repo", default=default_repo,
              help="""GitHub repository to search, specified as `owner/name` (default: "pokt-network/pocket")""")
# TODO: support parsing shorthand (e.g. 6mo, 2yr, 10d, etc.)
@click.option("--since", default=default_since,
              help="""Earliest issue/review commment date to consider (default: 6 months ago)""")
@click.argument("search_text")
def search(repo: str, since: str, search_text: str):
    # TODO: better search methodology!
    searchRegex = re.compile(f".*{search_text}.*", flags=re.IGNORECASE)

    pocket_repo = gh.get_repo(repo)
    comments = pocket_repo.get_pulls_comments(
        # TODO: parameterize sort & direction
        sort="created_at",
        direction="desc",
        since=datetime.strptime(since, iso_8601_format)
    )

    # TODO: understand paging...
    # Naive search
    # TODO: parallelize w/ reactivex (see: https://rxpy.readthedocs.io/en/latest/)
    for comment in comments:
        matches = searchRegex.search(comment.body)
        if matches is None:
            continue

        print(f"user: {comment.user} | created: {comment.created_at} | url: {comment.url}")
        print(f"comment body: {comment.body}")


if __name__ == "__main__":
    try:
        access_token = os.environ[access_token_key]
    except KeyError:
        print(f"{access_token_key} env var must be set")
        exit(1)

    gh = Github(access_token)
    search()
