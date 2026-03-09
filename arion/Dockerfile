FROM golang:1.24-alpine

WORKDIR /src
COPY . ./
RUN go mod download
RUN go build -o /arion

EXPOSE 4000
CMD ["/arion"]
