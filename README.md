# Muzz App Explore Service

A gRPC service that powers **"Liked You"** functionality:

- See who liked you.
- See new likes that you haven’t reciprocated yet.
- Count likes.
- Record decisions (like/pass) and detect mutual likes.

## Tech Stack

- **Go** (gRPC + GORM)
- **MySQL** (persistent storage)
- **Redis** (sorted sets for fast pagination of likes)
- **Docker Compose** (local orchestration)


## Requirements

- [Go 1.25+](https://go.dev/doc/install)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [make](https://www.gnu.org/software/make/) (optional, but recommended)


## Setup

### Clone the repo

```bash
git clone https://github.com/endyapina/muzzapp.git
cd muzzapp
```

## Running the App

### Docker Compose (recommended)

Start the Go service together with **MySQL**, **Redis**, and **Adminer**:

```bash
make start-services
