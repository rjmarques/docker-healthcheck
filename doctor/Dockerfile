FROM golang:bullseye

WORKDIR /app

# copy go mod files and download depedencies
# do with as a seperate step to build to cache it
COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "run", "main.go"]