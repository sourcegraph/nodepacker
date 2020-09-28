# nodepacker ![CI](https://github.com/ggilmore/kickstart/workflows/CI/badge.svg)

WIP

Small utility packs pods onto nodes, computes optimal machine type and number of nodes necessary to run a k8s
Sourcegraph installation.

Provides a REPL for a human operator to assist in packing decisions.

```text
~/work/src/github.com/sourcegraph/nodepacker/cmd/nodepacker(main*) » go run .                                                                                                                                  uwe@maxx
nodepacker> manifests_read /Users/uwe/tmp/base
got the manifests
nodepacker> nodes_pack
replica count for indexed search is 15
cluster with 15 nodes of machine type {n1-standard-16, cpu: 16, mem: 60 GB}
Pod assignment as follows:
node-0: [indexed-search-0, gitserver-3], free {cpu: 5.4, mem: 27.9 GB}
node-1: [indexed-search-1, repo-updater, prometheus, precise-code-intel-worker, grafana], free {cpu: 8.3, mem: 34.07 GB}
node-2: [indexed-search-2, sourcegraph-frontend-1, searcher, precise-code-intel-bundle-manager], free {cpu: 8.3, mem: 40.8 GB}
node-3: [indexed-search-3, gitserver-7], free {cpu: 5.4, mem: 27.9 GB}
node-4: [indexed-search-4, gitserver-0], free {cpu: 5.4, mem: 27.9 GB}
node-5: [indexed-search-5, gitserver-6], free {cpu: 5.4, mem: 27.9 GB}
node-6: [indexed-search-6, gitserver-4], free {cpu: 5.4, mem: 27.9 GB}
node-7: [indexed-search-7, gitserver-2], free {cpu: 5.4, mem: 27.9 GB}
node-8: [indexed-search-8, gitserver-8], free {cpu: 5.4, mem: 27.9 GB}
node-9: [indexed-search-9, redis-store, redis-cache, query-runner, jaeger], free {cpu: 8.38, mem: 29.308 GB}
node-10: [indexed-search-10, gitserver-9], free {cpu: 5.4, mem: 27.9 GB}
node-11: [indexed-search-11, pgsql], free {cpu: 7.49, mem: 41.801 GB}
node-12: [indexed-search-12, sourcegraph-frontend-0, symbols, syntect-server, github-proxy], free {cpu: 8.35, mem: 38.95 GB}
node-13: [indexed-search-13, gitserver-1], free {cpu: 5.4, mem: 27.9 GB}
node-14: [indexed-search-14, gitserver-5], free {cpu: 5.4, mem: 27.9 GB}
nodepacker> 
```
