import json
import os
import shutil
import subprocess
import sys
import time
import urllib
import urllib.request
import traceback


TEST_CASES = [

    # CASE 1
    ("""
     {"i":1, "t":"This is the first book I read"}
     {"i":2, "t":"I love to read a book"}
     {"i":3, "t":"The first time I tried hummus I liked it"}
     {"i":4, "t":"Spaceships are made of human dreams"}
     """,
     [("first", -1, [1, 3]),
      ("read book", -1, [1, 2]),
      ("the_i???", -1, [1, 3]),
      (" I ", 2, [1, 2])]
     )

]


def main():
    for test_case in TEST_CASES:
        doc_stream, queries = test_case
        run_test_case(doc_stream, queries)
    print("Success!")


def run_test_case(doc_stream, queries):

    try:
        shutil.rmtree(_p("tmp"))
    except FileNotFoundError:
        pass
    os.mkdir(_p("tmp"))

    documents_file = _p("tmp/documents.txt")
    with open(documents_file, "w") as fd:
        for line in doc_stream.strip().splitlines():
           fd.write(line.strip() + "\n")

    index_file = _p("tmp/index.raw")
    subprocess.check_call([_p("../makeindex" + _EXE),
                           "-i", documents_file,
                           "-o", index_file,
                           "-id", "i",
                           "-content", "t"])

    srv = None
    try:
        srv = subprocess.Popen([_p("../searchservice" + _EXE),
                                "-i", index_file,
                                "-n", "localhost",
                                "-p", "5987"])
        failures = 0
        # time.sleep(0.5)
        for query, limit, expected_postings in queries:
            try:
                execute_search(5987, query, limit, expected_postings)
            except:
                failures += 1
                traceback.print_exception(*sys.exc_info())
                pass
            assert failures == 0, "{} failures".format(failures)
    finally:
        if srv:
            srv.kill()
            srv.wait()


def execute_search(http_port, query, limit, expected_postings):

    postings = None
    http_query = None

    try:
        quoted_query = urllib.parse.quote_plus(query)
        if limit >= 0:
            http_query = "http://localhost:{}/search?l={}&q={}" \
                .format(http_port, limit, quoted_query)
        else:
            http_query = "http://localhost:{}/search?q={}" \
                .format(http_port, quoted_query)
        print("get:", http_query)
        response = urllib.request.urlopen(http_query).read()
        postings = json.loads(response.decode(encoding='utf-8'))
        assert postings == expected_postings
    except:
        print("Failed query: ['{}',{}] --({})--> [{}], expected [{}]".format(
            query, limit, http_query, postings, expected_postings))
        raise


def _p(rel_path):
    folder = os.path.dirname(__file__)
    path = os.path.join(folder, rel_path)
    return os.path.normpath(os.path.abspath(path))

_EXE = ""
if os.name == "nt":
    _EXE = ".exe"

if __name__ == "__main__":
    main()