import collections
import os
import re
import sqlite3
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


db = f"../../data/modules/pathway/pathway-20260527.db"

print(f"Writing to {db}")

rdfViewId = str(uuid.uuid7())


if os.path.exists(db):
    os.remove(db)

conn = sqlite3.connect(db)
conn.row_factory = sqlite3.Row

cursor = conn.cursor()

cursor.execute("PRAGMA journal_mode = WAL;")
cursor.execute("PRAGMA foreign_keys = ON;")

cursor.execute(
    f"""
    CREATE TABLE datasets (
        id INTEGER PRIMARY KEY,
        public_id TEXT NOT NULL UNIQUE,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    """,
)
cursor.execute("CREATE INDEX idx_datasets_name ON datasets (LOWER(name));")


cursor.execute(
    f"""
    CREATE TABLE collections (
        id INTEGER PRIMARY KEY,
        public_id TEXT NOT NULL UNIQUE,
        dataset_id INTEGER NOT NULL,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (dataset_id) REFERENCES datasets(id) ON DELETE CASCADE
    );
    """,
)
cursor.execute("CREATE INDEX idx_collections_name ON collections (LOWER(name));")
cursor.execute("CREATE INDEX idx_collections_dataset_id ON collections (dataset_id);")


cursor.execute(
    f"""
    CREATE TABLE genes (
        id INTEGER PRIMARY KEY,
        public_id TEXT NOT NULL UNIQUE,
        name TEXT NOT NULL UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    """,
)
cursor.execute("CREATE INDEX idx_genes_name ON genes (LOWER(name));")


cursor.execute(
    f"""
    CREATE TABLE pathways (
        id INTEGER PRIMARY KEY,
        public_id TEXT NOT NULL UNIQUE,
        collection_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
    );
    """,
)

cursor.execute("CREATE INDEX idx_pathways_collection_id ON pathways (collection_id);")
cursor.execute("CREATE INDEX idx_pathways_name ON pathways (name);")


cursor.execute(
    f"""
    CREATE TABLE pathway_genes (
        id INTEGER PRIMARY KEY,
        pathway_id INTEGER NOT NULL,
        gene_id INTEGER NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (pathway_id) REFERENCES pathways(id) ON DELETE CASCADE,
        FOREIGN KEY (gene_id) REFERENCES genes(id) ON DELETE CASCADE,
        UNIQUE (pathway_id, gene_id)
    );
    """,
)

cursor.execute(
    "CREATE INDEX idx_pathway_genes_pathway_id ON pathway_genes (pathway_id);"
)
cursor.execute("CREATE INDEX idx_pathway_genes_gene_id ON pathway_genes (gene_id);")


dataset_map = {}
collection_map = {}
pathway_map = {}
gene_map = {}

for file in os.listdir():
    if "gmt" not in file:
        continue

    dataset = "Staudt Lab SignatureDB" if "signaturedb" in file else "MSigDB"

    if dataset not in dataset_map:
        dataset_id = str(uuid.uuid7())
        dataset_map[dataset] = len(dataset_map) + 1

        cursor.execute(
            f"INSERT INTO datasets (id, public_id, name, description) VALUES (?, ?, ?, ?);",
            (
                dataset_map[dataset],
                dataset_id,
                dataset,
                None,
            ),
        )

    dataset_id = dataset_map[dataset]

    collection = re.sub(r"\.Hs.+", "", file)

    if collection not in collection_map:
        collection_id = str(uuid.uuid7())
        collection_map[collection] = len(collection_map) + 1

        cursor.execute(
            f"INSERT INTO collections (id, public_id, dataset_id, name) VALUES (?, ?, ?, ?);",
            (
                collection_map[collection],
                collection_id,
                dataset_id,
                collection,
            ),
        )

    collection_id = collection_map[collection]

    with open(file, "r") as pf:
        for line in pf:
            tokens = line.strip().split("\t")

            pathway = tokens[0]
            pathway = clean(pathway)

            if pathway not in pathway_map:
                pathway_id = str(uuid.uuid7())
                pathway_map[pathway] = len(pathway_map) + 1

                cursor.execute(
                    f"INSERT INTO pathways (id, public_id, collection_id, name) VALUES (?, ?, ?, ?);",
                    (
                        pathway_map[pathway],
                        pathway_id,
                        collection_id,
                        pathway,
                    ),
                )

            pathway_id = pathway_map[pathway]

            source = tokens[1]
            source = clean(source)

            genes = [clean(g) for g in list(sorted(set(tokens[2:])))]
            geneCount = len(genes)

            for g in genes:
                if g not in gene_map:
                    gene_id = str(uuid.uuid7())
                    gene_map[g] = len(gene_map) + 1

                    cursor.execute(
                        f"INSERT INTO genes (id, public_id, name) VALUES (?, ?, ?);",
                        (
                            gene_map[g],
                            gene_id,
                            g,
                        ),
                    )

                gene_id = gene_map[g]

                print(g, gene_id, pathway, pathway_id)

                cursor.execute(
                    f"INSERT INTO pathway_genes (pathway_id, gene_id) VALUES (?, ?);",
                    (
                        pathway_id,
                        gene_id,
                    ),
                )

conn.commit()
