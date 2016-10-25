import json
import itertools
import os
import random
import shutil
import subprocess
import sys
import traceback
import time
import urllib
import urllib.request


TEST_NAME = os.path.basename(__file__)[:-3]

CONTENT_FIELDS = "Actor 1,Actor 2,Actor 3,Distributor,Director,Fun Facts," \
                 "Locations,Production Company,Release Year,Title,Writer"

BULK_DOCUMENTS = [
    12, 13, 14, 15, 16, 19, 22, 25, 27, 28, 29, 30, 37, 42, 49, 52, 53, 59, 86,
    87,
    88, 90, 91, 92, 94, 96, 100, 108, 110, 112, 113, 114, 115, 116, 143, 158,
    159, 160, 161, 163, 164, 165, 166, 167, 168, 169, 170, 175, 176, 177, 178,
    182, 193, 195, 196, 197, 201, 204, 214, 218, 219, 220, 222, 223, 225, 227,
    228, 231, 238, 239, 240, 242, 245, 246, 248, 249, 250, 251, 252, 253, 254,
    259, 260, 261, 262, 263, 264, 265, 266, 267, 268, 269, 270, 271, 272, 273,
    274, 278, 279, 283, 285, 286, 289, 291, 292, 293, 294, 296, 299, 300, 304,
    306, 307, 308, 309, 310, 311, 312, 313, 315, 316, 317, 318, 319, 320, 325,
    326, 328, 331, 333, 334, 335, 337, 338, 339, 340, 341, 342, 343, 348, 355,
    358, 360, 364, 365, 366, 367, 368, 369, 370, 371, 372, 373, 374, 375, 376,
    377, 378, 379, 380, 381, 382, 385, 387, 388, 392, 405, 410, 411, 412, 413,
    414, 415, 427, 428, 429, 432, 434, 435, 438, 442, 443, 445, 448, 449, 450,
    451, 453, 454, 461, 463, 464, 472, 473, 475, 476, 477, 478, 479, 480, 481,
    482, 483, 484, 485, 486, 487, 488, 489, 490, 491, 492, 493, 494, 495, 496,
    498, 499, 501, 503, 505, 509, 510, 511, 512, 513, 514, 523, 524, 527, 530,
    543, 561, 565, 577, 580, 581, 582, 583, 584, 585, 586, 587, 588, 589, 590,
    591, 592, 593, 594, 595, 596, 597, 598, 601, 603, 604, 608, 609, 612, 613,
    614, 615, 617, 618, 622, 626, 627, 628, 629, 630, 631, 632, 633, 634, 635,
    636, 637, 638, 639, 650, 653, 656, 657, 663, 670, 671, 672, 673, 674, 677,
    679, 680, 686, 687, 688, 689, 690, 691, 692, 697, 698, 700, 705, 706, 708,
    743, 745, 751, 755, 765, 766, 767, 769, 771, 772, 773, 774, 775, 776, 777,
    778, 779, 780, 781, 782, 783, 784, 785, 786, 787, 788, 789, 790, 791, 792,
    794, 795, 796, 797, 799, 800, 801, 803, 805, 806, 807, 808, 809, 835, 836,
    837, 840, 842, 852, 853, 854, 855, 856, 857, 858, 859, 860, 861, 862, 863,
    864, 878, 879, 885, 887, 888, 890, 902, 903, 904, 907, 911, 913, 917, 918,
    919, 944, 945, 946, 947, 948, 949, 950, 951, 952, 953, 954, 955, 956, 957,
    958, 959, 960, 961, 962, 963, 964, 965, 966, 967, 968, 969, 970, 971, 972,
    973, 974, 975, 976, 977, 978, 979, 980, 981, 982, 983, 984, 985, 986, 987,
    988, 989, 990, 991, 992, 993, 994, 995, 996, 997, 998, 999, 1000, 1001,
    1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010, 1011, 1012, 1013,
    1014, 1015, 1016, 1017, 1018, 1019, 1020, 1021, 1022, 1023, 1024, 1025,
    1026, 1027, 1028, 1029, 1030, 1031, 1032, 1033, 1034, 1035, 1036, 1037,
    1038, 1039, 1040, 1041, 1042, 1043, 1044, 1045, 1046, 1047, 1048, 1049,
    1050, 1051, 1052, 1053, 1054, 1055, 1056, 1057, 1058, 1059, 1060, 1061,
    1062, 1063, 1064, 1065, 1066, 1067, 1068, 1069, 1070, 1071, 1072, 1073,
    1074, 1075, 1076, 1077, 1078, 1079, 1080, 1081, 1082, 1083, 1084, 1085,
    1086, 1087, 1088, 1089, 1090, 1091, 1092, 1093, 1094, 1095, 1096, 1097,
    1098, 1099, 1100, 1102, 1103, 1104, 1105, 1106, 1107, 1108, 1109, 1110,
    1111, 1112, 1113, 1114, 1115, 1116, 1117, 1118, 1119, 1120, 1121, 1122,
    1123, 1124, 1125, 1127, 1128, 1129, 1130, 1131, 1132, 1133, 1134, 1135,
    1136, 1137, 1138, 1139, 1140, 1141, 1142, 1143, 1144, 1145, 1146, 1147,
    1148, 1149, 1150, 1151, 1162, 1164, 1165, 1167, 1172, 1173, 1174, 1177,
    1179, 1181, 1183, 1192, 1193, 1194, 1195, 1200, 1205, 1207, 1208, 1209,
    1210, 1212, 1214, 1215, 1216, 1217, 1218, 1222, 1223, 1226, 1233, 1234,
    1235, 1236, 1237, 1238]


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
        time.sleep(2.0)
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

        expected_content = b"".join(docs[id_]
                                    for id_ in BULK_DOCUMENTS)
        try:
            fetch_docs(5987, BULK_DOCUMENTS, expected_content)
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