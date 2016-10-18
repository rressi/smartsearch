# SmartSearch

A small yet effective framework to implement micro search services.

This framework is build around a prefix tree (also know as trie) that is 
decoded and traversed for each term of the query.

Our trie is encoded in a space efficient manner with the express purpose to 
allow us to have many of them placed in a blob storage (a DB, some archive
 files etc.).


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
key-blending.

We plain to implement the search also in JavaScript with this use case in mind.


## Implemented features

The following feature have been implemented up to now:
- Modules *triebuilder* and *triereader* that are the core of our search 
framework. They will allow us to implement efficiently and with realtive 
simplicity many fancy algorithms. They share a binary format that is space 
efficient and fast (based on UVarint and delta encoding).
- A tool called [*makeindex*](doc/makeindex.md) to index streams of JSON 
documents and create 
serialized indices.
- A basic *normalization* layer and tokenizer that removes any decorations from
 our text strings keeping only letters and digits (for example Ä becomes a).
- The classic *AND query* that covers most of the use cases.
