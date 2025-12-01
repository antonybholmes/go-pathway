import collections
import os
import re
import string
import uuid_utils as uuid
import pandas as pd
from nanoid import generate

# good enough for Planetscale
# print(generate("0123456789abcdefghijklmnopqrstuvwxyz", 12))


printable = set(string.printable)


def clean(text):
    ret = "".join(filter(lambda x: x in printable, text))
    ret = re.sub(r"[\&]", "_", ret)
    ret = re.sub(r"_+", "_", ret)
    return ret


with open("../../data/modules/pathway/pathway.sql", "w") as f:

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

                    genes = list(sorted(set(tokens[2:])))
                    geneCount = len(genes)
                    genes = ",".join(genes)
                    genes = clean(genes)

                    organization = (
                        "Staudt Lab SignatureDB" if "signaturedb" in file else "MSigDB"
                    )

                    id = uuid.uuid7()
                    #     generate("0123456789abcdefghijklmnopqrstuvwxyz", 12)

                    # add a comma at the end of tags for exact search e.g. exactly BCL6 => 'BCL6,'
                    print(
                        f"INSERT INTO pathway (id, organization, dataset, name, source, gene_count, genes) VALUES ('{id}', '{organization}', '{db}', '{pathway}', '{source}', {geneCount}, '{genes}');",
                        file=f,
                    )

            print("COMMIT;", file=f)

    df = pd.read_csv("hugo_genes.tsv", sep="\t", header=0, index_col=0)

    genes = sorted(df.index.values)

    print("BEGIN TRANSACTION;", file=f)

    for gene in genes:
        id = uuid.uuid7()

        # add a comma at the end of tags for exact search e.g. exactly BCL6 => 'BCL6,'
        print(
            f"INSERT INTO genes (id, gene_symbol) VALUES ('{id}', '{gene}');",
            file=f,
        )

    print("COMMIT;", file=f)
