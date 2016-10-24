import json
import itertools
import os
import random
import shutil
import subprocess
import sys
import urllib
import urllib.request
import traceback


TEST_NAME = os.path.basename(__file__)[:-3]

CONTENT_FIELDS = "Actor 1,Actor 2,Actor 3,Distributor,Director,Fun Facts," \
                 "Locations,Production Company,Release Year,Title,Writer"


def main():

    try:
        shutil.rmtree(_p(""))
    except FileNotFoundError:
        pass
    os.mkdir(_p(""))

    docs = {}
    documents_file = _p("../documents.json")
    with open(documents_file, "r",
              encoding="utf-8") as fd:
        for line in fd:
            doc = json.loads(line)
            docs[doc["uuid"]] = line.encode(encoding="utf-8")

    srv = None
    try:
        srv = subprocess.Popen([_p("../../searchservice" + _EXE),
                                "-d", documents_file,
                                "-id", "uuid",
                                "-content", CONTENT_FIELDS,
                                "-n", "localhost",
                                "-p", "5987"])
        failures = 0
        # time.sleep(0.5)
        sample = random.sample(docs.items(), 10)
        for doc_id, expected_content in sample:
            try:
                fetch_docs(5987, [doc_id], expected_content)
            except:
                failures += 1
                traceback.print_exception(*sys.exc_info())
                pass

        sample = random.sample(docs.keys(), 4)
        for doc_ids in itertools.combinations(sample, 2):
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
    print("Success!")


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