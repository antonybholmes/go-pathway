# -*- coding: utf-8 -*-
"""
Generate a tss distribution for a region file

Created on Thu Jun 26 10:35:40 2014

@author: Antony Holmes
"""

import gzip
import sys
import collections
import re
import os
import pandas as pd
import numpy as np

gene_map = collections.defaultdict(lambda: collections.defaultdict(str))

with gzip.open("/ifs/scratch/cancer/Lab_RDF/ngs/references/gencode/grch38/gencode.v42.basic.annotation.gtf.gz", "rt") as f:
    for line in f:
        if line.startswith("#"):
            continue

        tokens = line.strip().split("\t")

        level = tokens[2]

        if level == "exon":
            continue

        chr = tokens[0]
        start = tokens[3]
        end = tokens[4]
        strand = tokens[6]

        matcher = re.search(r'gene_name "(.+?)";', tokens[8])

        if matcher:
            name = matcher.group(1)
        else:
            name = "NA"

        gene_map[name]["chr"] = chr
        gene_map[name]["start"] = start
        gene_map[name]["end"] = end
        gene_map[name]["strand"] = strand

        matcher = re.search(r'gene_id "(.+?)";', tokens[8])

        if matcher:
            gene_id = matcher.group(1)
            gene_id = re.sub(r"\..+", "", gene_id)
            gene_map[name]["gene_id"] = gene_id

        matcher = re.search(r'gene_type "(.+?)";', tokens[8])

        if matcher:
            gene_map[name]["gene_type"] = matcher.group(1)

        matcher = re.search(r'tag "(.+?)";', tokens[8])

        if matcher:
            gene_map[name]["tags"] += matcher.group(1) + ","


for gene in sorted(gene_map):
    gene_map[gene]["tags"] = ";".join(
        sorted(set([x for x in gene_map[gene]["tags"].strip().split(",") if x]))
    )

cols = ["chr", "start", "end", "strand", "gene_id", "gene_type", "tags"]

d = []

for gene in sorted(gene_map):
    d.append([gene_map[gene][c] for c in cols])

df = pd.DataFrame(d, index=sorted(gene_map), columns=cols)
df.to_csv("gencode.v42.basic.chrmt.tsv", sep="\t", header=True, index=True)
