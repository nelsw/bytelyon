"""bytelyon glue: launch a stealthy SeleniumBase CDP Mode Chromium instance
and hand back a normal Playwright `Browser`, connected to it over the Chrome
DevTools Protocol.

This exists so every *-seleniumbase handler's lambda_handler.py (see
../../lambda/grab-seleniumbase) can be a near line-for-line port of its
CloakBrowser-based counterpart (../../lambda/grab, ../../lambda/page-scraper,
...): the launch call site becomes

    from seleniumbase_playwright import launch   # instead of: from cloakbrowser import launch

and every other line — every `page.goto()` / `page.content()` /
`page.screenshot()` / etc. call — is unchanged, because both return a
standard `playwright.sync_api.Page`.

How it works (see https://seleniumbase.io/ — "Stealthy Playwright Mode"):
  1. `seleniumbase.sb_cdp.Chrome(...)` launches SeleniumBase's patched,
     stealthy Chromium directly as a subprocess (CDP Mode talks to it over
     its own --remote-debugging-port — no chromedriver/webdriver binary is
     ever involved).
  2. `sb.get_endpoint_url()` returns that CDP endpoint's HTTP address.
  3. `playwright.sync_api.sync_playwright().chromium.connect_over_cdp(...)`
     attaches a *normal* Playwright `Browser` to that already-running,
     already-stealthy browser session.
From that point on, every Playwright API (contexts, pages, network
interception, etc.) behaves exactly as it would against any other
Playwright-launched browser — SeleniumBase is only responsible for how the
browser process itself was started and hardened against bot detection.

Interface differences from cloakbrowser.launch() (kept as close as possible
for handler-code compatibility, but these are NOT drop-in equivalent):
  - `humanize` / `human_preset`: cloakbrowser-specific synthetic
    mouse/keyboard/scroll timing. SeleniumBase's stealth comes from its CDP
    Mode launch flags and patches rather than replayed human-like input
    timing, so these two are accepted (purely so handler code doesn't need
    an `if` for which provider it's using) but have no effect here.
  - `geoip`: cloakbrowser auto-derives timezone/locale from the proxy's
    exit IP. SeleniumBase has no built-in equivalent lookup. Accepted but
    unused; pass `timezone=`/`locale=` explicitly via **kwargs instead if
    you already know the values you want (forwarded straight to
    `seleniumbase.sb_cdp.Chrome()` — see its docstring / cdp_util.start()
    for the full list of supported kwargs).
"""

from __future__ import annotations

import asyncio
import logging
import os
import time
from contextlib import suppress
from typing import Any

from playwright.sync_api import Browser, Playwright, sync_playwright
from seleniumbase import sb_cdp

__all__ = ["launch"]

logger = logging.getLogger("seleniumbase_playwright")

# Empirically, a fresh Lambda execution environment's *first* Chromium
# launch attempt occasionally fails outright (observed failure modes range
# from an early, silent exit -- see _last_chromium_output -- to the
# generic "Failed to connect to the browser" once SeleniumBase's own single
# internal retry, in Browser.create(), is also exhausted), while a second
# attempt in the very same (now-warm) container succeeds reliably and
# quickly. This looks like resource/scheduling contention specific to a
# microVM's very first process spawn rather than anything wrong with the
# launch arguments -- retrying here, a couple of times, costs nothing on
# the (common) case where the first attempt just works.
_LAUNCH_ATTEMPTS = 3
_LAUNCH_RETRY_DELAY_S = 1.0

# Rolling buffer of the most recently launched Chromium subprocess's own
# stdout+stderr, captured by _patch_subprocess_for_diagnostics() below.
# SeleniumBase's Browser.start() (seleniumbase/undetected/cdp_driver/browser.py)
# always redirects the Chromium subprocess's stdout/stderr to DEVNULL, so a
# launch failure normally only ever surfaces as a generic, contentless
# "Failed to connect to the browser" exception -- exactly the kind of
# failure that's otherwise near-impossible to debug from CloudWatch Logs
# alone (this surfaced for real: the exact same image that works fine under
# a plain `docker run` fails to connect when actually deployed to Lambda,
# and DEVNULL hid the reason why).
_last_chromium_output: list[str] = []
_MAX_CAPTURED_LINES = 200


def _patch_subprocess_for_diagnostics() -> None:
    """Monkeypatch `asyncio.create_subprocess_exec` (module-level, process-wide,
    applied once) so every subprocess it launches -- in practice, just the
    Chromium process started by seleniumbase/undetected/cdp_driver/browser.py's
    `Browser.start()` -- has its stdout+stderr piped and drained into
    `_last_chromium_output` instead of silently discarded via DEVNULL. This
    only ever *adds* an observer; it doesn't change how the caller's own
    stdin/stdout/stderr kwargs behave otherwise (they're overridden here
    specifically because DEVNULL is what browser.py always passes).
    """
    if getattr(asyncio, "_bytelyon_subprocess_diagnostics_patched", False):
        return
    orig_create_subprocess_exec = asyncio.create_subprocess_exec

    async def _patched_create_subprocess_exec(*args, **kwargs):
        kwargs["stdout"] = asyncio.subprocess.PIPE
        kwargs["stderr"] = asyncio.subprocess.STDOUT
        process = await orig_create_subprocess_exec(*args, **kwargs)
        _last_chromium_output.clear()

        async def _drain() -> None:
            assert process.stdout is not None
            while True:
                line = await process.stdout.readline()
                if not line:
                    break
                _last_chromium_output.append(
                    line.decode("utf-8", "replace").rstrip("\n")
                )
                del _last_chromium_output[:-_MAX_CAPTURED_LINES]

        asyncio.get_event_loop().create_task(_drain())
        return process

    asyncio.create_subprocess_exec = _patched_create_subprocess_exec
    asyncio._bytelyon_subprocess_diagnostics_patched = True


# SeleniumBase assumes it can create relative-path directories/lock files in
# the current working directory -- e.g. "./downloaded_files/" (its default
# download folder, per https://seleniumbase.io/) and a pip/driver "find
# lock" file under it, both created unconditionally on every CDP Mode launch
# (seleniumbase/undetected/cdp_driver/cdp_util.py, __activate_virtual_display_as_needed()
# -- runs even when headless=True). That's fine for a normal dev machine,
# but AWS Lambda's runtime filesystem is read-only everywhere except /tmp,
# so a launch() call from the handler's own CWD (/app) fails with
# "[Errno 30] Read-only file system: b'downloaded_files'" the moment
# SeleniumBase tries to create it. Switching the process's CWD to /tmp
# (Lambda's one writable path; also always writable in plain `docker run`
# testing) once, here, sidesteps every one of these relative-path
# assumptions at the source instead of chasing them down individually.
def _ensure_writable_cwd() -> None:
    with suppress(Exception):
        if os.getcwd() != "/tmp":
            os.chdir("/tmp")


def _normalize_proxy(proxy: Any) -> str | None:
    """Convert a cloakbrowser/Playwright-style proxy value (a
    "http://[user:pass@]host:port" string, a bare "[user:pass@]host:port"
    string, or a Playwright `ProxySettings` dict) into the plain
    "[user:pass@]host:port" string that `seleniumbase.sb_cdp.Chrome(proxy=...)`
    expects (see `seleniumbase/undetected/cdp_driver/cdp_util.py`'s
    `start()` docstring: `proxy: "host:port" or "user:pass@host:port"`).

    Unlike Chromium's raw --proxy-server flag, SeleniumBase's CDP Mode
    handles the embedded credentials itself — Browser.get() installs a CDP
    Fetch.authRequired handler using the username/password parsed back out
    of this same string — so passing "user:pass@host:port" straight through
    is sufficient. No separate proxy-auth extension is needed (contrast
    with ../../lambda/grab/lambda_handler.py's `_normalize_proxy()`, which
    has to split credentials into Playwright's own ProxySettings shape for
    cloakbrowser/plain Playwright's launch(proxy=...) to authenticate).
    """
    if not proxy:
        return None
    if isinstance(proxy, str):
        # Strip a "http://"/"https://"/"socks5://" scheme prefix, if any.
        if "://" in proxy:
            proxy = proxy.split("://", 1)[1]
        return proxy
    if isinstance(proxy, dict):
        server = proxy.get("server", "")
        if "://" in server:
            server = server.split("://", 1)[1]
        username = proxy.get("username")
        password = proxy.get("password")
        if username and password:
            return f"{username}:{password}@{server}"
        return server
    raise ValueError(f"'proxy' must be a string or dict, got {type(proxy).__name__}")


class _ManagedBrowser:
    """Thin wrapper so callers can treat this exactly like a Playwright
    `Browser` (`.new_page()` / `.close()`, plus anything else via
    delegation), while a single `.close()` also tears down the underlying
    SeleniumBase Chromium process and the Playwright client connection —
    mirroring cloakbrowser's `Browser.close()` contract, where one call
    releases everything.
    """

    def __init__(self, playwright: Playwright, browser: Browser, sb: Any):
        self._playwright = playwright
        self._browser = browser
        self._sb = sb

    def __getattr__(self, item):
        # Delegate everything else (new_context, contexts, version, ...)
        # straight through to the real Playwright Browser object.
        return getattr(self._browser, item)

    def new_page(self, **kwargs) -> Any:
        # SeleniumBase's own CDP Mode launch already opened one tab
        # (navigated to "about:blank"). Reuse it by default -- same as the
        # seleniumbase.io "Stealthy Playwright Mode" example
        # (`page = browser.contexts[0].pages[0]`) -- instead of opening a
        # second, redundant tab. Any kwargs (e.g. a custom viewport) mean
        # the caller wants a genuinely new page/context, so fall through to
        # Playwright's normal `new_page()` in that case.
        if not kwargs and self._browser.contexts and self._browser.contexts[0].pages:
            return self._browser.contexts[0].pages[0]
        return self._browser.new_page(**kwargs)

    def close(self) -> None:
        with suppress(Exception):
            self._browser.close()
        with suppress(Exception):
            self._playwright.stop()
        with suppress(Exception):
            self._sb.quit()


def launch(
    *,
    headless: bool = True,
    humanize: bool = False,  # accepted for interface parity; see module docstring
    human_preset: str = "careful",  # accepted for interface parity; unused
    proxy: Any = None,
    geoip: bool = False,  # accepted for interface parity; see module docstring
    args: list[str] | None = None,
    **kwargs: Any,
) -> _ManagedBrowser:
    """Launch a stealthy SeleniumBase CDP Mode Chromium instance and return
    a Playwright `Browser` connected to it over CDP. See the module
    docstring for the full cloakbrowser.launch() compatibility notes.

    :param headless: Same meaning as cloakbrowser.launch(headless=...).
    :param proxy: "http://[user:pass@]host:port" string, a bare
        "[user:pass@]host:port" string, or a Playwright `ProxySettings` dict.
    :param args: Extra Chromium command-line flags (e.g. for Lambda -- see
        ../../lambda/grab-seleniumbase/lambda_handler.py). "--no-zygote" is
        always added regardless (see below); no need to pass it yourself.
    :param kwargs: Forwarded as-is to `seleniumbase.sb_cdp.Chrome()`, e.g.
        `locale=`, `timezone=`, `user_agent=`, `incognito=`, `guest=` --
        see https://seleniumbase.io/ and `cdp_util.start()`'s docstring
        (in the `seleniumbase` package) for the full list.

    Retries the actual Chromium launch up to `_LAUNCH_ATTEMPTS` times before
    giving up -- see that constant's comment for why (in short: a fresh
    Lambda execution environment's very first launch attempt is observed to
    occasionally fail even with otherwise-correct arguments, while retrying
    in the same, now-warm container succeeds reliably).
    """
    _ensure_writable_cwd()
    _patch_subprocess_for_diagnostics()

    browser_args = list(args or [])
    # Lambda's restricted process/container model needs "--no-zygote"
    # regardless of caller-supplied args -- same rationale as every
    # CloakBrowser-based handler (see ../../lambda/base-image/Dockerfile and
    # ../../lambda/grab/lambda_handler.py): Lambda can't fork from
    # Chromium's zygote process. "--disable-gpu" avoids spawning a separate
    # GPU process no Lambda container can use anyway.
    # ("--disable-dev-shm-usage" is already one of SeleniumBase's own
    # default CDP Mode browser args -- see
    # seleniumbase/undetected/cdp_driver/config.py -- so it isn't repeated
    # here.)
    #
    # Deliberately NOT added: "--single-process". It looks like an obvious
    # extra hardening flag for Lambda's restricted process model (and is
    # what various third-party "Chromium on Lambda" projects use), but
    # confirmed by testing against a real deployed Lambda function, it's
    # actively harmful here: with it, Chromium starts (no crash, no fatal
    # error -- see the captured-output diagnostics below) but never opens
    # its CDP debug port, so every launch times out with a contentless
    # "Failed to connect to the browser"
    # (seleniumbase/undetected/cdp_driver/browser.py, start()). Removing it
    # fixed launches immediately. This reproduced only in real Lambda, not
    # in a plain `docker run` of the same image -- Lambda's extra syscall
    # sandboxing on top of the container apparently doesn't get along with
    # whatever `--single-process` needs internally. If you're tempted to
    # re-add it (e.g. chasing a different problem), test against a real
    # deployed Lambda function first, not just `docker run`/RIE locally.
    for flag in ("--no-zygote", "--disable-gpu"):
        if flag not in browser_args:
            browser_args.append(flag)

    sb = None
    last_exc: Exception | None = None
    for attempt in range(1, _LAUNCH_ATTEMPTS + 1):
        try:
            sb = sb_cdp.Chrome(
                headless=headless,
                proxy=_normalize_proxy(proxy),
                browser_args=browser_args,
                # Chromium's own internal sandbox needs namespace privileges
                # that Lambda's container runtime doesn't grant -- same
                # reason headless Chrome/Chromium in most container/CI
                # environments needs --no-sandbox. seleniumbase's Config
                # auto-disables this when running as uid 0, but Lambda's
                # runtime user isn't guaranteed to be root, so it's forced
                # off explicitly here instead of relying on that
                # auto-detection.
                sandbox=False,
                **kwargs,
            )
            break
        except Exception as e:
            last_exc = e
            if attempt < _LAUNCH_ATTEMPTS:
                logger.warning(
                    "seleniumbase.sb_cdp.Chrome() launch attempt %d/%d "
                    "failed (%s); retrying in %.1fs. Captured Chromium "
                    "stdout/stderr so far:\n%s",
                    attempt,
                    _LAUNCH_ATTEMPTS,
                    e,
                    _LAUNCH_RETRY_DELAY_S,
                    "\n".join(_last_chromium_output) or "(none captured)",
                )
                time.sleep(_LAUNCH_RETRY_DELAY_S)
    if sb is None:
        # SeleniumBase's own exception here is a generic, contentless
        # "Failed to connect to the browser" -- re-raise with whatever the
        # Chromium subprocess itself printed before dying/hanging on the
        # final attempt, captured by _patch_subprocess_for_diagnostics()
        # above, since that's almost always the actual actionable
        # information.
        if _last_chromium_output:
            raise RuntimeError(
                f"seleniumbase.sb_cdp.Chrome() failed to launch Chromium "
                f"after {_LAUNCH_ATTEMPTS} attempt(s). Captured Chromium "
                "stdout/stderr from the last attempt:\n"
                + "\n".join(_last_chromium_output)
            ) from last_exc
        raise last_exc
    endpoint_url = sb.get_endpoint_url()
    playwright = sync_playwright().start()
    try:
        browser = playwright.chromium.connect_over_cdp(endpoint_url)
    except Exception:
        with suppress(Exception):
            playwright.stop()
        with suppress(Exception):
            sb.quit()
        raise
    return _ManagedBrowser(playwright, browser, sb)
