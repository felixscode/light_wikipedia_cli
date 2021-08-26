"""
Wikipedia command line tool
"""
from genericpath import exists
import os
import optparse
import wikipedia as wp
import pdfkit
from pathlib import Path
import subprocess


class wikipedia_term:

    def __init__(self):
        pass

    def search_article(self, query, num_suggestions):
        suggestions = wp.search(query)
        print('Suggestions:')
        print(80*'-')
        for num, sugg in enumerate(suggestions):
            print("{}: {}".format(num+1, sugg))
        print(80*'-')
        choosen_sugg = input("Choose Aritcle (1-{}):".format(len(suggestions)))
        try:
            return wp.page(suggestions[int(choosen_sugg)-1], auto_suggest=False)
        except:
            print(
                'Sorry bad option page couldnt be found. Searching again for selected option')
            self.search_article(
                suggestions[int(choosen_sugg)-1], num_suggestions)

    def summarize(self, page):
        return (page.title, page.summary)

    def get_content(self, page):
        return page.title, page.content

    def get_random_article(self):
        try:
            return wp.page(wp.random())
        except:
            return self.get_random_article()

    def term_print(self, info_tupel):
        name, article = info_tupel
        messg = "\n\n{}\n\n{}".format(name, article)
        print(messg)
        return

    def render(self, page, pdf_viewer='zathura'):

        path = Path(__file__).parent / 'tmp/sample.pdf'
        if not os.path.exists(path.parent):
            os.makedirs(path.parent)
        url = page.url
        pdfkit.from_url(url, path)
        print(str(path.absolute()))
        subprocess.call([pdf_viewer, str(path.absolute())])


def add_parser_opts():
    parser = optparse.OptionParser()
    parser.add_option("--search", "-s", type='string',
                      help="specify an search query if empty or not providen random article is choosen")
    parser.add_option("--pdf", "-p", action="store_true", default=False,
                      help="renders wikipage to pdf and opens pdf viewer")
    parser.add_option("--full", "-f", action="store_true",
                      default=False, help="show initre content")
    parser.add_option("--deutsch", "-d", action="store_true",
                      default=False, help="search and show in german")
    return parser


def main():
    wpt = wikipedia_term()
    parser = add_parser_opts()
    (options, args) = parser.parse_args()
    if options.deutsch:
        wp.set_lang('de')
    if options.search != None:
        page = wpt.search_article(options.search, 5)
    else:
        page = wpt.get_random_article()

    if options.pdf:
        wpt.render(page)
    elif options.full:
        wpt.term_print(wpt.get_content(page))
    else:
        wpt.term_print(wpt.summarize(page))


if __name__ == '__main__':
    main()
