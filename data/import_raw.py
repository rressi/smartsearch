"""
This script imports data from the original document to generate a consistent
text file made of a stream of JSON objects (each object is a dict at its
top-level).
"""

import argparse
import gzip
import json
import sys


def load_input(raw):
    """Takes the original JSON data structure and generates a stream of
    data items as dictionaries made of key/value pairs."""

    columns = [(column, column['dataTypeName'] == 'text')
               for column in raw['meta']['view']['columns']]

    print('')
    print("[COLUMNS]")
    for column, selected in columns:
        if selected:
            try:
                description = column['description'].strip()[:50]
            except KeyError:
                description = ''
            print("{}: {}".format(column['name'],
                                  description))
    print('')

    for item in raw['data']:
        yield {column['name']: value
               for (column, selected), value in zip(columns, item)
               if selected and type(value) is str}


def main(arguments):
    """It rakes an input file with the source data and converts it
    to a stream of JSON documents.
    """

    # The input file can be optionally encoded with gzip format:
    input_file = arguments.input_file[0]
    assert isinstance(input_file, str)
    if input_file.endswith(".gz"):
        _open = gzip.open
    else:
        _open = open
    with _open(input_file, "rt",
               encoding='utf-8') as fd:
        print("Loading JSON content into memory....")
        raw = json.load(fd)  # Parses all the input file.

    # Also the output file can be optionally encoded with gzip format:
    output_file = arguments.output_file[0]
    assert isinstance(output_file, str)
    uuid = 0
    if output_file.endswith(".gz"):
        _open = gzip.open
    else:
        _open = open
    with _open(output_file, "wt",
               encoding='utf-8') as fd:
        # for each element extracted from the input
        print("Generating distilled file")
        for item in load_input(raw):
            uuid += 1  # generates incremental uuid from 1
            item['uuid'] = uuid
            fd.write(json.dumps(item,
                                sort_keys=True))
            fd.write("\n")  # one encoded document per line

    print("{} documents imported".format(uuid))


if __name__ == "__main__":
    description = " ".join(map(str.strip,
                               __doc__.strip().splitlines()))
    parser = argparse.ArgumentParser(prog="import_raw",
                                     description=description)
    parser.add_argument(
        'input_file',
        nargs=1,
        default="raw.json",
        type=str,
        help='File containing raw data from https://datasf.org/')
    parser.add_argument(
        'output_file',
        nargs=1,
        default="documents.json",
        type=str,
        help='Distilled JSON file to be generated.')
    main(parser.parse_args(sys.argv[1:]))
