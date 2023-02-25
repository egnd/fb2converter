#!make

MAKEFLAGS += --always-make --no-print-directory

static/dictionaries:
	cd ../static/dictionaries && \
		wget -r -l1 --no-parent -nd -A.pat.txt http://ctan.math.utah.edu/ctan/tex-archive/language/hyph-utf8/tex/generic/hyph-utf8/patterns/txt
	cd ../static/dictionaries && \
		wget -r -l1 --no-parent -nd -A.hyp.txt http://ctan.math.utah.edu/ctan/tex-archive/language/hyph-utf8/tex/generic/hyph-utf8/patterns/txt
	cd ../static/dictionaries && for item in $$(ls *.txt); do gzip $$item; done

static/sentences:
	cd ../static/sentences && \
		curl -L https://api.github.com/repos/neurosnap/sentences/tarball | tar xz --wildcards '*/data/*.json' --strip-components=2
	$(MAKE) _russian-senteces
	cd ../static/sentences && for item in $$(ls *.json); do gzip $$item; done

# https://github.com/Mottl/ru_punkt
# https://github.com/mhq/train_punkt
_russian-senteces:
ifeq ($(wildcard russian.pickle),)
	wget https://github.com/Mottl/ru_punkt/archive/refs/heads/master.tar.gz -O ru_punkt.tar.gz \
        && mkdir -p ru_punkt && tar --strip-components=1 -xzf ru_punkt.tar.gz -C ru_punkt
	cp ru_punkt/nltk_data/tokenizers/punkt/PY3/russian.pickle russian.pickle
	rm -rf ru_punkt*
endif
	python3 convert.py | jq "." > ../static/sentences/russian.json
