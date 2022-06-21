import requests
from bs4 import BeautifulSoup

URL = "https://trek.nasa.gov/tiles/apidoc/trekAPI.html?body=moon"
page = requests.get(URL)

print(page.status_code)

soup = BeautifulSoup(page.content, "html.parser")

print(soup.prettify())

res = soup.find_all("media-body")

print(res)
