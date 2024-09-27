import collections
import os
import re
import string

import pandas as pd


printable = set(string.printable)


def clean(text):
    ret = "".join(filter(lambda x: x in printable, text))
    ret = re.sub(r"[\&]", "_", ret)
    ret = re.sub(r"_+", "_", ret)
    return ret


with open("pathway.sql", "w") as f:

    for file in os.listdir():
        if "gmt" in file:
            db = re.sub(r"\.Hs.+", "", file)
            print(file)

            print("BEGIN TRANSACTION;", file=f)

            with open(file, "r") as pf:
                for line in pf:
                    tokens = line.strip().split("\t")

                    pathway = tokens[0]
                    pathway = clean(pathway)

                    source = tokens[1]
                    source = clean(source)

                    genes = ",".join(tokens[2:])
                    genes = clean(genes)

                    # add a comma at the end of tags for exact search e.g. exactly BCL6 => 'BCL6,'
                    print(
                        f"INSERT INTO pathway (dataset, name, source, genes) VALUES ('{db}', '{pathway}', '{source}', '{genes}');",
                        file=f,
                    )

            print("COMMIT;", file=f)

    df = pd.read_csv(
        "gencode.v42.basic.restricted.chrmt.tsv", sep="\t", header=0, index_col=0
    )

    genes = sorted(df.index.values)

    print("BEGIN TRANSACTION;", file=f)

    for gene in genes:

        # add a comma at the end of tags for exact search e.g. exactly BCL6 => 'BCL6,'
        print(
            f"INSERT INTO genes (gene_symbol) VALUES ('{gene}');",
            file=f,
        )

    print("COMMIT;", file=f)
