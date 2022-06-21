import bs4
import requests
import logging

# xpath = /html/body/div[2]/div[2]/div[2]/div/div[5]/div[2]/p[1]/span[1]/button

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    logging.info("Starting apiEndpointer.py")

    url = "https://trek.nasa.gov/tiles/apidoc/trekAPI.html?body=moon"

    logging.info("Getting url: " + url)
    response = requests.get(url)
    logging.info("Got response")
    soup = bs4.BeautifulSoup(response.text, "html.parser")
    logging.info("Got soup")
    # get all matching xpaths
    print(soup)

    logging.info("Endpointing")
