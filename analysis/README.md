Given the following files, produced by the [processing](../processing) phase:
 - the scraped data in the `bytecode` directory, as a list of `.wasm` files with their sha256 sum as name
 - the corresponding JavaScript files that loaded them in `source`, with their sha256 sum as name
 - `results.csv` containing some metadata

The following questions can be answered.

# How many Wasm modules do each domain use?
```sh
$ ./use-per-domain.sh
...
      4 yoursuper.com
      4 zaco.by
      5 abc7.com
...
```

# How many domains use the same Wasm module?
```sh
$ ./use-per-module.sh
4       add757886deecba0270184f127cc2a258e7ac4f72c23214f2604d1ec32fb2f82
1       b1b0cb0702b7678e4c526f144f58e4f549335c46f1f4785bfcf703afa5ef5f69
10      b49a74f9a31c0c5bd4d1a4254b1f5869a3af985b633362999b473570ad2b153c
4       b4be03291b936c621ba31c52baf19e7f9a2b6bb7fc1dfdcb984288b08593ff01
1       b4f7547ecdf2486059b2bb3cdeaa237e05794cd245acda77eea12439ff1b96db
...
```
Can be sorted with `sort -g`.
*Interpretation*: 4 different domains ure using the module `bytecode/add757886deecba0270184f127cc2a258e7ac4f72c23214f2604d1ec32fb2f82.wasm`.

# What are the imports of a Wasm module?
Requires [wassail](https://github.com/acieroid/wassail).

There are two interesting questions here:
  1. Are there any imports that are common across multiple modules? For this, we can count for each import (characterized by its name and type signature) how many modules have it. This does not mean that the actual implementation of the import will be the same though (it often happens that imports have stripped names, such as "a", and there is not a lot of variety in the types of WebAssembly). This is what is produced in the `import-usages.txt` file.
  2. How many imports do modules have? This is what is produced in the `imports-per-module.txt` file.

Finally, the `import-modules.txt` file contains the raw data, listing every import of every module.


```
$ ./imports.sh
...
Produced 3 file:
	import-modules.txt lists imports of each module, one import per line
	import-usages.txt counts total number of usage per import, one import per line
	imports-per-module.txt counts the number of imports per module, one module per line
```

## Excerpt from `import-usages.txt`
```
22 w	i32, i32, i32 ->
23 emscripten_memcpy_big	i32, i32, i32 -> i32
23 k	i32 -> i32
23 s	i32 -> i32
```

## Excerpt from `imports-per-module.txt`
```
5	038066b226120b2e49493a0206bf5a55d31c34bf9e4f5e88ae990df11be1ccf1
1	07d1047b81529c9e7040c4d56fec1b574f656db37badbe05cda7ae65d0d04b53
1	09dfe684cd5c12e0b28b536d612164e0fdcbb37caac35cd2c98c986783703ece
357	0f1da969cc0b52fb34b230fd815971c5f56f31bdeb599f2397252ddec5d60129
1	12917cf106a884bc016697fe24971eefe108bf5a4ba34519c83cc5ae57fffe10
```
# What are the exports of a Wasm module
Requires [wassail](https://github.com/acieroid/wassail).
This is similar to the imports, but this time for functions *exported* by WebAssembly to JavaScript.

```
$ ./exports.sh
...
Produced 2 file:
	export-modules.txt lists exports of each module, one export per line
	export-usages counts total number of usage per export, one export per line
```

# Which module use non-standard extensions of WebAssembly
Requires wasm2wat.
```
$ ./use-extended-features.sh
...
165 modules do not use extended features, while 2 do
```

# Which instructions are the most used?
Requires [wassail](https://github.com/acieroid/wassail).
```
$ ./count-instructions.sh
...
2320985 i32.add
2831347 load
5340205 i32.const
10168217 local.get
```

*Interpretation*: `local.get` is used 10168217 times in the data set.
# What is the statistical distribution of binary sizes of WebAssembly modules?
```
$ ./binary-size.h
min: 8 bytes
max: 18613979 bytes
median: 86053 bytes
mean: 917116 bytes
```
# What is the size of sections in WebAssembly module?
Requires [wassail](https://github.com/acieroid/wassail).
```
$ ./section-sizes.sh
...
type:
	min: 0
	max: 3443
	median: 99
	mean: 377.709
import:
	min: 0
	max: 35860
	median: 70
	mean: 1344.11
func:
	min: 0
	max: 15449
	median: 53
	mean: 1447.83
...
```
(Sizes are given in bytes)

# How many modules import their memory?
Requires [wassail](https://github.com/acieroid/wassail).
```
$ ./memory-imports.sh
...
66 modules import their memory, while 101 do not
```

# How many modules export their memory?
Requires [wassail](https://github.com/acieroid/wassail).
```
$ ./memory-exports.sh
...
66 modules import their memory, while 101 do not
```




# How is the code loaded from JavaScript

In progress.

1. Pick a file from `bytecode/`, e.g. `6599b93117bfcf25deff23fe88b6e5fc5f39c2bde522b83a040834bd0d64adbd.wasm`
2. Find how it is referenced from `result.csv`:
```
grep 6599b93117bfcf25deff23fe88b6e5fc5f39c2bde522b83a040834bd0d64adbd.wasm results.csv | cut -d, -f5
```
3. If the result is empty, that means that it's not explicitly loaded from JavaScript. It should be manually checked then...

We can count how many seem never loaded from JavaScript:
```bash
DIR=bytecode
UNKNOWN=0
JS=0
for module in $(ls $DIR/*.wasm)
do
    HASH=$(basename $module | cut -d'.' -f1)
    ISINRESULTS=$(grep $HASH results.csv)
    MATCHES=$(grep $HASH results.csv | cut -d, -f5 | grep -v '^$')
    if [ -z "$MATCHES" ]; then
        UNKNOWN=$(echo "$UNKNOWN+1" | bc)
        echo $HASH
    else
        JS=$(echo "$JS+1" | bc)
    fi
done
echo "Unknown: $UNKNOWN"
echo "JS: $JS"
```

Result:
Unknown: 147
JS: 20

# Are there similar WebAssembly modules?
TODO

Previous work used cosine similarity.
Instead, we could look at the following survey: https://arxiv.org/pdf/1909.11424.pdf
But there's no tool that would work out of the box.

# What compiler was used to produce a binary?
Information from the names may help.
But that probably will not be enough.
Similarity can help too.

