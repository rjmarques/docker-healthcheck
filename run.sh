#!/bin/bash

function finish {
  docker-compose down
}
trap finish EXIT

docker-compose up --build