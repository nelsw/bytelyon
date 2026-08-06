from xml.etree.ElementTree import Element, fromstring

from aiohttp import ClientSession


async def fetch_xml(session: ClientSession, url: str) -> Element[str] | None:
    print(f"[ ] fetch_xml {url}")
    async with session.get(url) as response:
        if response.status >= 300:
            print(f"[!] fetch_xml {url} - {response.status}")
            return None

        print(f"[+] fetch_xml {url}")
        return fromstring(text=await response.text(encoding="utf-8"))
