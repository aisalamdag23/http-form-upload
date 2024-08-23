# HTTP Form Upload

1. This is a simple HTTP server and handler that serves a HTML form with 2 fields:

    - A hidden input field named "auth" that receives its input value passed from the Go code. This token should be passed as an environment variable to the application.

    - A file upload field named "data" (ie. `<input type="file"/>`) that uploads a file that the user selects

2. The form should POST data to the /upload handler, which should write the received file data to a temporary file.
3. Before accepting any data, this checks that the content type of the uploaded file is an image, and that the auth token matches.
   If the submission is bad, this returns a 403 HTTP error code. Images larger than 8 megabytes are also rejected.
4. The image metadata (content type, size, etc) including all relevant HTTP information is being written to Postgres database.

## Getting Started

### Prerequisites

- Go 1.20 or higher 
- Docker
- Make
- Goose

## Running the application
1. Run the setup command.
```
make setup
```
2. Wait for setup to complete
3. To start the application, run
```
make run
```
4. Once the application is running, open your web browser and navigate to:
```
http://localhost:8080
```
