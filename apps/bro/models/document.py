from bs4 import BeautifulSoup
from pydantic import BaseModel


class Document(BaseModel):
    soup: BeautifulSoup