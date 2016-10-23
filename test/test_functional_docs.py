import json
import itertools
import os
import shutil
import subprocess
import sys
import urllib
import urllib.request
import traceback


TEST_NAME = os.path.basename(__file__)[:-3]

DOCS = """
     {"id":1, "title":"This is the first book I read"}
     {"id":26, "title":"I love to read a book"}
     {"id":13, "title": "The first time I tried hummus", "p": [10, 1]}
     {"id":4, "title":"Spaceships are made of human dreams"}
     """


def main():

    try:
        shutil.rmtree(_p(""))
    except FileNotFoundError:
        pass
    os.mkdir(_p(""))

    docs = {}
    documents_file = _p("documents.txt")
    with open(documents_file, "w",
              encoding="utf-8") as fd:
        for line in DOCS.strip().splitlines():
            line = line.strip() + "\n"
            fd.write(line)
            doc = json.loads(line)
            docs[doc["id"]] = line.encode(encoding="utf-8")

    srv = None
    try:
        srv = subprocess.Popen([_p("../../searchservice" + _EXE),
                                "-d", documents_file,
                                "-id", "id",
                                "-content", "title",
                                "-n", "localhost",
                                "-p", "5987"])
        failures = 0
        # time.sleep(0.5)
        for doc_id, expected_content in docs.items():
            try:
                fetch_docs(5987, [doc_id], expected_content)
            except:
                failures += 1
                traceback.print_exception(*sys.exc_info())
                pass
        for doc_ids in itertools.combinations(docs.keys(), 2):
            expected_content = b"".join(docs[id_]
                                        for id_ in doc_ids)
            try:
                fetch_docs(5987, doc_ids, expected_content)
            except:
                failures += 1
                traceback.print_exception(*sys.exc_info())
                pass
        assert failures == 0, "{} failures".format(failures)

    finally:
        if srv:
            srv.kill()
            srv.wait()


def fetch_docs(http_port, doc_ids, expected_content):

    http_query = None

    try:
        quoted_ids = urllib.parse.quote_plus(" ".join(str(id_)
                                                      for id_ in doc_ids))
        http_query = "http://localhost:{}/docs?ids={}".format(http_port,
                                                              quoted_ids)
        print("get:", http_query)
        content = urllib.request.urlopen(http_query).read()
        assert content == expected_content
    except:
        print("\nFAILED")
        print("ids:      [{}]".format(" ".join(str(id_)
                                               for id_ in doc_ids)))
        print("URL:      [{}]".format(http_query))
        print("got:      [{}]".format(content.decode(encoding="'utf-8")))
        print("expected: [{}]".format(expected_content
                                      .decode(encoding="'utf-8")))
        print("")
        raise


def _p(rel_path):
    folder = os.path.dirname(__file__)
    path = os.path.join(folder, TEST_NAME, rel_path)
    return os.path.normpath(os.path.abspath(path))

_EXE = ""
if os.name == "nt":
    _EXE = ".exe"


if __name__ == "__main__":
    main()