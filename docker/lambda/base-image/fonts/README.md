# Optional: real Windows/Office fonts

This directory is baked into the image at `/usr/share/fonts/windows/` (see
`../Dockerfile`) and is empty by default — just this README, which
`fc-cache` silently ignores (not a font file).

The `Dockerfile` already installs `ttf-mscorefonts-installer` (Debian's
packaging of Microsoft's real "Core Fonts for the Web" — Arial, Times New
Roman, Courier New, Georgia, Verdana, etc. — fetched from Sourceforge at
build time under Microsoft's own redistribution EULA). That's a legitimate,
automatable source of *genuine* Microsoft font files, not a metric-compatible
substitute like `fonts-liberation`.

CloakBrowser's own guidance is that this still isn't enough on its own — a
real Windows install has 200+ font files, and Office adds ~50 more, versus
~13 in the core-fonts package. If you have legal access to those (e.g. a
`C:\Windows\Fonts\` folder and Office fonts from a machine/license you
control), drop the `.ttf`/`.otf` files directly into this directory before
building the base image:

```bash
cp /path/to/windows/Fonts/*.ttf docker/lambda/base-image/fonts/
cp /path/to/office/fonts/*.ttf  docker/lambda/base-image/fonts/
```

Then rebuild `../Dockerfile` as usual (see `../INSTRUCTIONS.md`) — the
`COPY fonts/ /usr/share/fonts/windows/` step picks up whatever's here, and
`fc-cache -f` registers it.

**Do not** commit real Microsoft font files to this repo — they're
proprietary and not freely redistributable. This directory is `.gitignore`d
except for this README specifically so each engineer can drop their own
locally, without them landing in version control or the shared team image
unless you specifically choose to publish a base image built with them.
