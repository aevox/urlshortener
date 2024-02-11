# URL Shortener Service

The URL Shortener is a simple web service written in Go, designed to create short, memorable URLs from longer ones.
It utilizes PostgreSQL for persistent storage of URLs, ensuring that shortened URLs remain accessible over time.

## Features

- Shorten URLs to a manageable length
- Redirect to original URLs using short links
- Persistent storage with PostgreSQL

## Prerequisites

- Go (1.21 or later recommended)
- PostgreSQL
- Docker (optional, for containerization)

## Setting Up the Project

### Database Setup

1. Create a new database for the project:

   ```sql
   CREATE DATABASE urlshortener;
   ```

2. Create the `urls` table:

   ```sql
   CREATE TABLE urls (
       slug CHAR(6) PRIMARY KEY,
       original_url TEXT NOT NULL
   );
   ```

### Application Setup

1. Clone the repository to your local machine:

   ```bash
   git clone <repository-url>
   ```

2. Navigate to the project directory:

   ```bash
   cd urlshortener
   ```

3. Build the project:

   ```bash
   make build
   ```

4. Update the PostgreSQL connection details in `main.go` to match your setup.

## Running the Application

To start the URL Shortener service:

```bash
go run main.go
```

The service will start listening on `http://localhost:8080`. Use the `/shorten` endpoint to create short URLs and navigate to short URLs to be redirected to the original URL.

## Using Docker (Optional)

To build a Docker image for the application:

```bash
docker build -t urlshortener .
```

To run the application in a Docker container:

```bash
docker run -p 8080:8080 urlshortener
```

Alternatively you can use docker compose:
```bash
docker-compose up
```


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
