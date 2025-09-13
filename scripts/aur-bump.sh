#!/bin/bash

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# Update pkgver in PKGBUILD
sed -i "s/^pkgver=.*/pkgver=$VERSION/" packaging/arch/PKGBUILD

# Change to packaging/arch directory
cd packaging/arch

# Update checksums
updpkgsums

# Generate .SRCINFO
makepkg --printsrcinfo > .SRCINFO

echo "Updated PKGBUILD and .SRCINFO for version $VERSION"
