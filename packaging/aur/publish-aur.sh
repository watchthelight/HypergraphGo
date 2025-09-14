#!/usr/bin/env bash

set -euo pipefail

# Usage: publish-aur.sh <version> [<pkgbuild-dir>]
# Default pkgbuild-dir: packaging/aur

if [ $# -lt 1 ]; then
    echo "Usage: $0 <version> [<pkgbuild-dir>]"
    exit 1
fi

VERSION=$1
PKGDIR=${2:-packaging/aur}

echo "Starting AUR publish for HypergraphGo $VERSION"

# Check makepkg availability
command -v makepkg >/dev/null 2>&1 || {
    echo "✖ Error: makepkg not available. Install pacman and makepkg."
    exit 1
}

# Check PKGBUILD exists
if [ ! -f "$PKGDIR/PKGBUILD" ]; then
    echo "✖ Error: PKGBUILD not found in $PKGDIR"
    exit 1
fi
echo "✔ Found PKGBUILD"

# Update pkgver in PKGBUILD
sed -i.bak -E 's/^(pkgver=).*/\1'"$VERSION"'/' "$PKGDIR/PKGBUILD" && rm -f "$PKGDIR/PKGBUILD.bak"
echo "✔ Updated pkgver to $VERSION"

# Regenerate .SRCINFO
cd "$PKGDIR"
makepkg --printsrcinfo > .SRCINFO
if [ ! -s .SRCINFO ]; then
    echo "✖ Error: Failed to generate .SRCINFO"
    exit 1
fi
echo "✔ Regenerated .SRCINFO"

# Validate updates
if ! grep -q "pkgver=$VERSION" PKGBUILD; then
    echo "✖ Error: pkgver not updated in PKGBUILD"
    exit 1
fi
if ! grep -q "pkgver = $VERSION" .SRCINFO; then
    echo "✖ Error: pkgver not updated in .SRCINFO"
    exit 1
fi
echo "✔ Validated updates"

# Prepare AUR clone
AUR_DIR="$HOME/.aur/hypergraphgo"
if [ ! -d "$AUR_DIR" ]; then
    echo "Cloning AUR repository..."
    git clone ssh://aur@aur.archlinux.org/hypergraphgo.git "$AUR_DIR"
fi

cd "$AUR_DIR"
# Ensure correct origin (optional check)
if ! git remote get-url origin | grep -q "aur.archlinux.org/hypergraphgo.git"; then
    echo "✖ Error: AUR repo has incorrect origin"
    exit 1
fi

# Copy files
cp "$OLDPWD/$PKGDIR/PKGBUILD" .
cp "$OLDPWD/$PKGDIR/.SRCINFO" .
echo "✔ Copied files to AUR working copy"

# Commit and push
git add PKGBUILD .SRCINFO
if ! git diff --cached --quiet; then
    git commit -m "release: v$VERSION"
    if git push; then
        echo "✔ Pushed to AUR"
    else
        echo "✖ Error: git push failed. Ensure SSH key is set up for AUR and passphrase is cached or not required."
        exit 1
    fi
else
    echo "No changes to push."
fi

echo "✅ AUR publish for HypergraphGo $VERSION completed."
