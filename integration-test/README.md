# Data Population Script

This script is used to **populate test data** into the Explore Service so you can stress-test and verify the **ListLikedYou** and **ListNewLikedYou** endpoints.  

It creates **500 fake users** who all "like" a given recipient (`endy` by default).


## Prerequisites

- A running instance of the Explore Service (via `make up` or `go run main.go`)
- [`grpcurl`](https://github.com/fullstorydev/grpcurl) installed

### Install `grpcurl`

**On macOS (Homebrew):**
```bash
brew install grpcurl
