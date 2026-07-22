cd shop-scraper
git pull
docker build -t drinks .
docker rm -f drinks || true
docker run --name drinks \
    --add-host=host.docker.internal:host-gateway \
    -e POSTGRES_URI=postgres://drinks@host.docker.internal:5432/drinks \
    -e DOCKER=1 \
    -e LUCKIN
    drinks
cd ..
