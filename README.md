# transaction-latency

docker build -t transaction-latency .
docker run -v (pwd)/data:/data --env-file .env --rm -it  transaction-latency