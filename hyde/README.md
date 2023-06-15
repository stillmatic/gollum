# HyDE

This module is an implementation of [HyDE: Precise Zero-Shot Dense Retrieval without Relevance Labels](https://github.com/texttron/hyde).

We differ from the reference Python implementation by using an in-memory exact search (instead of FAISS) and OpenAI Ada embeddings instead of Contriever. Realistically you should expect slightly worse performance (exact search is slower than approximate search, calling OpenAI is probably slower than a local model for smaller batch sizes). However, the results should still be valid.

However, this modules _only_ expects the interfaces, not the actual implementations -- so you could implement a FAISS interface that connects to a docker instance. Or the same for an embedding service, just make a wrapper compatible with the OpenAI interface.