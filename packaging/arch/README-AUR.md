# Publishing to Arch Linux AUR

To publish a new version to the AUR:

1. Clone the AUR repository:
   ```
   git clone ssh://aur@aur.archlinux.org/hottgo-bin.git
   cd hottgo-bin
   ```

2. Copy the updated files:
   ```
   cp /path/to/HypergraphGo/packaging/arch/PKGBUILD .
   cp /path/to/HypergraphGo/packaging/arch/.SRCINFO .
   ```

3. Commit and push:
   ```
   git add PKGBUILD .SRCINFO
   git commit -m "Update to version X.Y.Z"
   git push
