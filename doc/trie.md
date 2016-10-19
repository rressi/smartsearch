# trie

![Trie](https://upload.wikimedia.org/wikipedia/commons/b/be/Trie_example.svg)

Follows a nice description of a *trie* from en.wikipedia.org:

> In computer science, a trie, also called digital tree and sometimes radix tree
or prefix tree (as they can be searched by prefixes), is a kind of search 
tree—an ordered tree data structure that is used to store a dynamic set or 
associative array where the keys are usually strings.

> Unlike a binary search tree, no node in the tree stores the key associated 
with that node; instead, its position in the tree defines the key with which it
is associated.

> All the descendants of a node have a common prefix of the string associated 
with that node, and the root is associated with the empty string. Values are 
not necessarily associated with every node.

> Rather, values tend only to be associated with leaves, and with some inner 
nodes that correspond to keys of interest. For the space-optimized presentation 
of prefix tree, see compact prefix tree.

Our implementation encodes a trie in a space-efficient binary blob that is 
decoded and traversed during search.

## Exact match

To match the term `book` module *triereader* is used in the
following form:

```
[root node] 
 |--edge b--> [node]
               |--edge o--> [node] 
                             |--edge o--> [node]
                                           |--edge k--> [node]
                                                         |-> postings
```

At the node corresponding to the last letter of an indexed *term* we have all
 the *postings* (read document ids) matching that *term*.
 
We just do the same for each *term* then we have just to intersect all the 
*postings* coming from each *term*: only the postings found on all terms 
represents our final result.

As said, *triereader* is a finite state machine that decodes on the fly the 
nodes, edges, postings traversed making it very efficient in cases when each 
index is used few times but we have many indices to consume.


## Prefix match

Prefix match is pretty similar to *exact match* with the simple addition that
 at the first node where there are no edges matching we take recursively all 
 the postings from the last node and all its children.
 
The postings are then united, sorted and deduplicated.


## References
- https://en.wikipedia.org/wiki/Trie