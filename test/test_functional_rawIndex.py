import os
import shutil
import subprocess
import time
import urllib
import urllib.request

TEST_NAME = os.path.basename(__file__)[:-3]

TEST_DOCUMENTS = """
     {"i":1, "t":"This is the first book I read"}
     {"i":2, "t":"I love to read a book"}
     {"i":3, "t":"The first time I tried hummus I liked it"}
     {"i":4, "t":"Spaceships are made of human dreams"}
     """


def main():

    try:
        shutil.rmtree(_p(""))
    except FileNotFoundError:
        pass
    os.mkdir(_p(""))

    documents_file = _p("documents.txt")
    with open(documents_file, "w",
              encoding="utf-8") as fd:
        for line in TEST_DOCUMENTS.strip().splitlines():
            fd.write(line.strip() + "\n")

    index_file = _p("index.raw")
    subprocess.check_call([_p("../../makeindex" + _EXE),
                           "-i", documents_file,
                           "-o", index_file,
                           "-id", "i",
                           "-content", "t"])
    with open(index_file, "rb") as fd:
        expected_raw_index = fd.read()

    srv = None
    try:
        srv = subprocess.Popen([_p("../../searchservice" + _EXE),
                                "-i", index_file,
                                "-n", "localhost",
                                "-p", "5987"])

        http_query = "http://localhost:5987/rawIndex"
        print("get:", http_query)
        time.sleep(2.0)
        raw_index = urllib.request.urlopen(http_query).read()
        assert raw_index == expected_raw_index, \
            "\nraw_index: [{}]\n".format(" ".join(hex(b)
                                                  for b in raw_index)) \
            + "expected:  [{}]\n".format(" ".join(hex(b)
                                                  for b in expected_raw_index))
    finally:
        if srv:
            srv.kill()
            srv.wait()
    print("Success!")


def _p(rel_path):
    folder = os.path.dirname(__file__)
    path = os.path.join(folder, TEST_NAME, rel_path)
    return os.path.normpath(os.path.abspath(path))

_EXE = ""
if os.name == "nt":
    _EXE = ".exe"

if __name__ == "__main__":
    main()