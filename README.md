# DOCKER
docker build -t drinks shop-scraper
docker run --name drinks --add-host=host.docker.internal:host-gateway -e POSTGRES_URI=postgres://drinks@host.docker.internal:5432/drinks -e DOCKER=1 drinks

# CRON
crontab -e
```
*/10 * * * * sleep $(awk 'BEGIN{srand(); print int(rand()*600)}'); docker restart drinks
```

```
select store_id, store_name, vendor, waiting_cups, waiting_time, created_at from entry;
```
```
sudo -i -u postgres psql -d drinks -c "select store_id, store_name, vendor, waiting_cups, waiting_time, created_at from entry ORDER BY created_at DESC;"
```