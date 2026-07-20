# DOCKER
docker build -t drinks shop-scraper

# CRON
```
*/10 * * * * sleep $(awk 'BEGIN{srand(); print int(rand()*600)}'); docker rm -f drinks-app >/dev/null 2>&1; docker run -d --name drinks-app --add-host=host.docker.internal:host-gateway -e POSTGRES_URI=postgres://drinks@host.docker.internal:5432/drinks drinks
```
