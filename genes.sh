cat pathway.sql | cut -d" " -f12 | sed -r 's/"//g' | sed -r 's/\);//' | sed -r 's/,/\n/g' | sort | uniq > genes.txt
