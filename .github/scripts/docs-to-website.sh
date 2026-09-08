#!/usr/bin/env bash
# Mirrors the tfplugindocs output into the kestra.io website docs collection.
# Usage: docs-to-website.sh <provider-docs-dir> <website-terraform-docs-dir>
set -euo pipefail

SRC="${1:?provider docs dir required}"
DEST="${2:?website terraform docs dir required}"

# The transform runs from a temp copy, so both paths have to be absolute.
SRC="$(cd "$SRC" && pwd)"
mkdir -p "$DEST"
DEST="$(cd "$DEST" && pwd)"

FRONTMATTER='NR==1 && $0!="---" {exit} NR==1 {next} $0=="---" {exit} {print}'
BODY='NR==1 {if ($0=="---") {fm=1; next} else {out=1}}
      fm && !out {if ($0=="---") out=1; next}
      out {print}'

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp -R "$SRC"/. "$WORK/"
cd "$WORK"

# The provider index page is published as the "configurations" guide.
mkdir -p guides
sed -Ei 's/page_title: "([^ ]+).*"/title: Provider configurations/' index.md
sed -Ei 's/^# kestra Provider/# Provider configurations/' index.md
mv index.md guides/configurations.md

# Frontmatter: tfplugindocs pages are generated, so they are not editable on GitHub.
find . -type f -name '*.md' -exec sed -i 's/subcategory: ""/editLink: false/g' {} +
find data-sources resources -type f -name '*.md' \
    -exec sed -Ei 's/page_title: "([^ ]+).*"/title: \1/' {} +
find guides -type f -name '*.md' -exec sed -Ei 's/page_title: "([^"]+)"/title: \1/' {} +

# Body: hcl highlighting, Astro alert directives, links to the website page layout.
find . -type f -name '*.md' -exec sed -i \
    -e 's/```terraform/```hcl/g' \
    -e 's/^-> \(.*\)/:::alert{type="info"}\n\1\n:::/' \
    -e 's/^~> \(.*\)/:::alert{type="warning"}\n\1\n:::/' \
    -e 's/^!> \(.*\)/:::alert{type="danger"}\n\1\n:::/' \
    -e 's#](\(\.\./[^)]*\)\.md)#](../\1/index.md)#g' \
    -e 's#](\([A-Za-z0-9._-]*\)\.md)#](../\1/index.md)#g' \
    {} +

skipped=""
while read -r file; do
    page="${file#./}"
    out="$DEST/${page%.md}/index.md"

    # Guides are prose the website team edits in place; only publish new ones.
    if [ -f "$out" ] && [ "${page%%/*}" = guides ]; then
        skipped="$skipped $page"
        continue
    fi

    # The website owns the frontmatter (titles, descriptions); only the body is mirrored.
    if [ -f "$out" ]; then
        frontmatter="$(awk "$FRONTMATTER" "$out")"
    else
        frontmatter="$(awk "$FRONTMATTER" "$file")"
    fi

    body="$(awk "$BODY" "$file" | awk '
        /^```/ {fenced = !fenced}
        # The website renders the H1 from frontmatter, so demote the generated one.
        !fenced && !demoted && /^# / {
            demoted = 1
            title = substr($0, 3)
            if (sub(/ \(Resource\)$/, "", title)) print "## Terraform Resource: " title
            else if (sub(/ \(Data Source\)$/, "", title)) print "## Terraform Data Source: " title
            else print "## " title
            next
        }
        {print}
    ' | sed -e 's/^## Example Usage$/## Example usage/' -e '/./,$!d')"

    # The docs collection requires a title, and upstream guides do not always carry one.
    if ! grep -q '^title:' <<<"$frontmatter"; then
        heading="$(sed -n 's/^## //p' <<<"$body" | head -1)"
        frontmatter="$(printf 'title: %s\n%s' "${heading:-$(basename "${page%.md}")}" "$frontmatter")"
    fi

    mkdir -p "$(dirname "$out")"
    printf -- '---\n%s\n---\n\n%s\n' "$frontmatter" "$body" > "$out"
done < <(find data-sources resources guides -type f -name '*.md' | sort)

echo "Mirrored the provider docs into $DEST"
[ -n "$skipped" ] && echo "Guides already on the website, left untouched:$skipped"
exit 0
