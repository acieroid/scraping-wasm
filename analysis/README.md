
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
Can be sorted with `sort -g`
# What are the imports of a Wasm module?
Requires [wassail](https://github.com/acieroid/wassail).
```
$ ./imports.sh
...
Produced 2 file:
	import-modules.txt lists imports of each module, one import per line
	import-usages counts total number of usage per import, one import per line
```

# What are the exports of a Wasm module
Requires [wassail](https://github.com/acieroid/wassail).
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
# How is the code loaded from JavaScript
TODO
# Are there similar WebAssembly modules?
TODO
