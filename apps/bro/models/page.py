from pydantic import BaseModel


class Page(BaseModel):
    content: str
    domain: str
    index: int
    kind: str
    meta: object
    screenshot_key: str
    title: str
    url: str
