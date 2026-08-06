from bs4 import BeautifulSoup, Tag
from bs4.element import NavigableString


class Doc:
    def __init__(self, html_content: str | bytes):
        self.soup = BeautifulSoup(html_content, "html.parser")
        self.meta = dict[str, list[str]]()
        for tag in self.soup.find_all("meta"):
            k = tag.get("name") or tag.get("property")
            if not isinstance(k, str):
                continue

            v = tag.get_attribute_list("content")
            if not k or not v:
                continue

            k = k.lower()
            if k not in self.meta:
                self.meta[k] = []

            self.meta[k].extend([val.strip().lower() for val in v if val.strip()])

    def value(self, *keys: str) -> str:
        for k in keys:
            vals = self.meta.get(k.lower())
            if vals:
                for v in vals:
                    v = v.strip()
                    if v:
                        return v
        return ""

    def title(self) -> str:
        v = self.value("twitter:title", "og:title", "title")
        if v:
            return v
        title_tag = self.soup.find("title")
        return title_tag.get_text().strip() if title_tag else ""

    def img_url(self) -> str:
        return self.value(
            "twitter:image:src",
            "twitter:image",
            "og:image:secure_url",
            "og:image:url",
            "og:image",
            "image",
        )

    def img_alt(self) -> str:
        return self.value("twitter:image:alt", "og:image:alt")

    def source(self) -> str:
        return self.value("twitter:site", "og:site_name", "og:site")

    def description(self) -> str:
        return self.value(
            "twitter:description", "og:description", "description", "abstract"
        )

    def keywords(self) -> list[str]:
        kw = set()
        for opt in ["keywords", "news_keywords", "article:tag"]:
            vals = self.meta.get(opt)
            if vals is None:
                continue
            for val in vals:
                for v in val.split(","):
                    kw.add(v)

        return sorted(kw)

    def body(self) -> str:
        # Clone soup to avoid modifying the original if needed,
        # though in Go version it seems to modify the document if it falls back to 'body' or 'html'

        sel = self.soup.find("article")
        if not sel:
            sel = self.soup.find("main")
        if not sel:
            sel = self.soup.find("body")
            if not sel:
                # If body is missing, we clean up the whole soup and use it
                sel = self.soup

        unique_text = {}
        # The Go code does: sel.Find("*").Contents().Each(...)
        # and checks if node.Type == html.TextNode
        # In BeautifulSoup, we can iterate over all descendants and check if they are NavigableString

        # We need to maintain order based on appearance
        all_elements = sel.find_all(True)  # True finds all tags

        # We also need to check the selection itself if it's a tag and has direct text
        elements_to_check = [sel] + all_elements

        index = 0
        for element in elements_to_check:
            for content in element.contents:
                if isinstance(content, NavigableString) and not isinstance(
                    content, Tag
                ):
                    text = str(content).strip()
                    if text and text not in unique_text:
                        unique_text[text] = index
                    index += 1

        # Sort by index to preserve order
        ordered_text = sorted(unique_text.items(), key=lambda item: item[1])
        return "\n\n".join([item[0] for item in ordered_text])
