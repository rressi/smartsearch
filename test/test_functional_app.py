import itertools
import os
import random
import shutil
import subprocess
import sys
import time
import traceback
import urllib
import urllib.request


TEST_NAME = os.path.basename(__file__)[:-3]


def main():

    try:
        shutil.rmtree(_p(""))
    except FileNotFoundError:
        pass
    os.mkdir(_p(""))

    # Generates random files:
    files = {}
    with open(_p("index.raw"), "wb") as fd:
        fd.write(bytes([0, 0]))
    with open(_p("index.html"), "wb") as fd:
        content = b"This is just a test\n"
        fd.write(content)
        files[""] = content
    for name, content in itertools.islice(zip(rand_file_names(),
                                              rand_file_content()),
                                              5):
        files[name] = content
        with open(_p(name), "wb") as fd:
            fd.write(content)

    srv = None
    try:
        srv = subprocess.Popen([_p("../../searchservice" + _EXE),
                                "-i", _p("index.raw"),
                                "--app", _p(""),
                                "-p", "5987"])
        failures = 0
        time.sleep(2.0)
        for name in sorted(files.keys()):
            try:
                fetch_file(5987, name, files[name])
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


def rand_file_names():
    files = set()
    while True:
        name = chr(random.randint(ord("a"),
                                  ord("z")))
        while random.randint(1, 6) != 6:
            name += chr(random.randint(ord("a"),
                                       ord("z")))
        if name not in files:
            yield name
            files.add(name)


def rand_file_content():
    while True:
        size = random.randint(0, 1024)
        content = bytes(random.randint(0, 255)
                        for _ in range(size))
        yield content


def fetch_file(http_port, file_name, expected_content):

    http_query = None
    content = None

    try:
        quoted_file = urllib.parse.quote_plus(file_name)
        http_query = "http://localhost:{}/app/{}".format(http_port, quoted_file)
        print("get:", http_query)
        content = urllib.request.urlopen(http_query).read()
        assert content == expected_content
    except:
        print("\nFAILED")
        print("file_name: [{}]".format(file_name))
        print("URL:       [{}]".format(http_query))
        print("got:       [{}]".format(show_bytes(content)))
        print("expected:  [{}]".format(show_bytes(expected_content)))
        print("")
        raise


def show_bytes(src):
    if type(src) is bytes:
        return " ".join(hex(value)
                        for value in src)
    else:
        return str(src)


def _p(rel_path):
    folder = os.path.dirname(__file__)
    path = os.path.join(folder, TEST_NAME, rel_path)
    return os.path.normpath(os.path.abspath(path))

_EXE = ""
if os.name == "nt":
    _EXE = ".exe"


if __name__ == "__main__":
    main()