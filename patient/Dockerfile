FROM golang:bullseye

WORKDIR /app

# copy go mod files and download depedencies
# do with as a seperate step to build to cache it
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# waits (1s) before the healthcheck starts
HEALTHCHECK  --interval=5s --timeout=3s --start-period=1s \  
    CMD curl --fail -s http://localhost/api/healthcheck || exit 1

CMD ["go", "run", "main.go"]