# SmartSearch

A small yet effective framework to implement micro search services.

This framework is build around a prefix tree (also know as *trie*) that is 
decoded and traversed for each term of the query.

Our trie is encoded in a space efficient manner with the express purpose to 
allow us to have many of them placed in a blob storage (a DB, some archive
 files etc.).
 
You can give a look to demonstrative web application 
[smartsearch.cc](http://smartsearch.cc) just to have an idea to what you can 
do with this framework. The code of it is on GitHub's repo called 
[smartsearch-demo](https://github.com/rressi/smartsearch-demo).


## Components

### smartsearch framework
It is a search library, with many modules like the following:
- *triebuilder* and [*triereader*](doc/trie.md) that are the core of 
our search framework. They will allow us to implement efficiently and with 
relative simplicity many fancy algorithms. They share a binary format that is 
space efficient and fast (based on UVarint and delta encoding).
- *normalizer* and *tokenizer* where we normalize queries removing unnecessary
characters decorations (for example Ä becomes a), all irrelevant characters and
where we isolate all the single pure tokens before indexing or searching.
- *indexbuilder* and *index* that are on top of all other components are 
meant to make easy and effective doing search.

### makeindex
[*makeindex*](doc/makeindex.md) preprocess streams of JSON documents and 
creates serialized indices.

### searchservice
[*searchservice*](doc/searchservice.md) is a production ready web service that
can be used as a backend for web apps that needs to search.

It implements the following features:
- it can take a pre-built index (built with *makeindex*) and offer a RESTful 
  method to search (the classic `/search?q=my+fancy+query`). It returns a 
  list of document ids.
- it can index a stream of JSON documents at boot and then serve `/search` 
  plus another method `/docs` that gets document ids and returns selected 
  documents as a stream of JSON documents.
- it can also serve one folder from the file system (method `/app/*`) so that
  can be used to host static web apps.


## Use cases 
 
### Many different set of documents
Imagine to have many users on your web-site, each of them need to search 
only their resources. You can do it with many mini pre-computed indices 
covering only the mini-universe of document that each user can access.

### A distributed index
There are cases (for example in a digital map) where time-to time we need to 
search only on an area of our DB and this area is different each time. In 
this case we need to select the indices we want to search, search on each of 
them and then unite the results.

The convenience of this approach is that we can iterate through indices, 
search on them and stop when we have collected enough results.

### Move the search to the client
With our micro-index we can move our search to the client (for example on 
the web browser).
In this case we can obtain:
- **Reduce server load** by letting client's devices do some computation for us.
- Drastically **reduce the latency** on client-side with auto-completion or 
key-blending. We plain to implement the search also in JavaScript with this 
use case in mind.
