import os, subprocess
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List


WIKI_DIR = "tools/wiki"
TEMP_WIKI = "tools/temp_wiki"

@dataclass
class DocumentationFile:
    file_name: str
    display_name: str
    path: Path


def get_all_raw_markdown_file_paths() -> List[str]:
    return os.popen('find . -name "*.md" | grep -v -e "vendor" -e "app"').readlines()


def get_all_markdown_file_paths() -> List[Path]:
    for raw_path in get_all_raw_markdown_file_paths():
        path = Path(raw_path.strip())
        if path.is_file():
            yield path


def get_file_to_wiki_comment(file_path: Path) -> Path:
    # gets the last line of of the file and parses it for the format
    with open(file_path) as file:
        wiki_comment = file.readlines()[-1]
        if "GITHUB_WIKI" in wiki_comment:
            mapping_format = wiki_comment.split(" ")[2].strip()
        return Path(mapping_format)


def categorize_paths() -> Dict[str, List[DocumentationFile]]:
    sidebar = defaultdict(list)
    paths = list()
    for path in get_all_markdown_file_paths():
        wiki_path_format = get_file_to_wiki_comment(path)

        # doc info parsing
        dirname = os.path.dirname(wiki_path_format)
        file = os.path.basename(wiki_path_format)
        categories = dirname.split("/")
        display_name = " ".join([s.title() for s in file.split("_")])
        filename = f"{categories[-1].title()} {display_name.title()}"

        sidebar[dirname].append(DocumentationFile(filename, display_name, path))

    return sidebar


def write_sidebar_file(sidebar: Dict[str, List[DocumentationFile]]) -> None:
    sidebar_format = "'''Contents'''\n"
    sidebar_format += "*'''[[ Home | Home ]]'''\n"

    level = 1
    for category, doc_files in sidebar.items():
        if category == "home":
            continue

        subcategories = category.split("/")
        for subcategory in subcategories:
            sidebar_format += ("*" * level) + f"'''{subcategory}'''\n"
            level += 1

        for doc in doc_files:
            sidebar_format += ("*" * level) + f"[[ {doc.file_name} | {doc.display_name} ]]\n"

        level = 1

    with open(f"{WIKI_DIR}/_Sidebar.mediawiki", "w") as f:
        f.write(sidebar_format)


def write_wiki_pages(sidebar: Dict[str, List[DocumentationFile]]) -> None:
    for category, doc_files in sidebar.items():
        for doc_file in doc_files:
            with open(doc_file.path) as source:
                target = f"{WIKI_DIR}/{doc_file.file_name}.md"
                if category == "home":
                    target = f"{WIKI_DIR}/Home.md"

                with open(target, "w") as dest:
                    dest.write(source.read())


def run_wiki_migration():
    os.makedirs(TEMP_WIKI, exist_ok=True)

    secret = os.environ['USER_TOKEN']
    user_name = os.environ['USER_NAME']
    user_email = os.environ['USER_EMAIL']
    owner = os.environ['OWNER']
    repo_name = os.environ['REPOSITORY_NAME']

    # init, pull, delete
    subprocess.call(["git", "init"], cwd=f"{TEMP_WIKI}")
    subprocess.call(["git", "config", "user.name",  f"{user_name}"], cwd=f"{TEMP_WIKI}")
    subprocess.call(["git", "config", "user.email", f"{user_email}"], cwd=f"{TEMP_WIKI}")
    subprocess.call(["git", "pull", f"https://{secret}@github.com/{owner}/{repo_name}.wiki.git"], cwd=f"{TEMP_WIKI}")

    # sync the new and old wiki files
    subprocess.call(["rsync", "-av",  "--delete", "tools/wiki/", "tools/temp_wiki", "--exclude", ".git"])

    # add, commit and push
    subprocess.call(["git", "add", "."], cwd=f"{TEMP_WIKI}")
    subprocess.call(["git", "commit", "-m", 'update wiki content'], cwd=f"{TEMP_WIKI}")
    subprocess.call(["git", "remote", "add", "master", \
        f"https://{secret}@github.com/{owner}/{repo_name}.wiki.git"], cwd=f"{TEMP_WIKI}")
    subprocess.call(["git", "push", "--set-upstream", "master", "master"], cwd=f"{TEMP_WIKI}")



if __name__ == "__main__":
    os.makedirs(WIKI_DIR, exist_ok=True)

    sidebar_format_dict = categorize_paths()
    write_sidebar_file(sidebar_format_dict)
    write_wiki_pages(sidebar_format_dict)

    # perform a migration for the git wiki
    run_wiki_migration()
