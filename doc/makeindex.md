# makeindex

It is a tool that takes JSON streams like the following:

```json
{"i":10, "t":"Title to beindexed", "c":"Some content to be indexed..."}
{"i":2, "t":"Another title", "c":"More content..."}
...
```

And is able to generate one single index if executed in this way:

```sh
makeindex -i inputstream.txt -id i -content t,c -o output.idx
```

It can also be used this way:

```
zcat inputstream.txt.gz | makeindex -id i -content t,c > output.idx
```


## Command line usage

```sh
$ ./makeindex.exe --help
Usage of makeindex:
  -content string
        Json attributes to be indexed, comma separated (default "content")
  -i string
        Input file (default "-")
  -id string
        Json attribute for document ids (default "id")
  -o string
        Output file (default "-")
```


## How to build

The first time you need to fetch the prerequisites, you can execute `init.sh` or
just do the following:

```sh
export GOPATH=$(pwd)
go get golang.org/x/text
```

It downloads and compiles `golang.org/x/text` that we use for text 
normalization.

Then to compile `makeindex` on your machine you can run `build.sh` or just 
type the following:

```sh
export GOPATH=$(pwd)
go build src/makeindex.go
```

The generated binary file should be placed in the root folder of the project.
