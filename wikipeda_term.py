"""
Wikipedia command line tool
"""
import os
import optparse
import wikipedia as wp
import pdfkit
from pathlib import Path
import subprocess

class wikipedia_term:
    
    def __init__(self):
        pass

    def search_article(self,query,num_suggestions):
        suggestions = wp.search(query,num_suggestions)
        print('Suggestions:')
        print(80*'-')
        for num,sugg in enumerate(suggestions):
            print("{}: {}".format(num+1,sugg))
        print(80*'-')
        choosen_sugg = input("Choose Aritcle (1-{}):".format(len(suggestions)))
        return wp.page(suggestions[int(choosen_sugg)-1])

    def summarize(self,page):
        return (page.title,page.summary)

    def get_content(self,page):
        return page.title,page.content

    def get_random_article(self):
        return wp.page(wp.random())

    def term_print(self, info_tupel):
        name,article = info_tupel
        messg = "\n\n{}\n\n{}".format(name,article)
        print(messg)
        return

    def render(self,page,pdf_viewer='zathura'):

        path = Path(__file__).parent / 'tmp/sample.pdf'
        if not os.path.exists(path.parent):
            os.makedirs(path.parent)
        url = page.url
        pdfkit.from_url(url, path)
        print(str(path.absolute()))
        subprocess.call([pdf_viewer, str(path.absolute())])


def main():
    p = optparse.OptionParser()
    p.addoption('‑‑person', '‑p', default="world")
    options, arguments = p.parseargs()
    print('Hello %s' % options.person)

if __name__ == '__main__':
    wpt = wikipedia_term()
    wpt.render(wpt.get_random_article())


    