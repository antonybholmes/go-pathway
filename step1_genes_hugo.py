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

df = pd.read_csv("/ifs/scratch/cancer/Lab_RDF/ngs/references/hugo/hugo_20240524.tsv", sep="\t", header=0, index_col=0)

df = df[df["Status"] == "Approved"]

 
df.to_csv("hugo_genes.tsv", sep="\t", header=True, index=False)
