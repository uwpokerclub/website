# api

To build the image, run:

```bash
docker-compose build
```

To start the container, run:

```bash
docker-compose up -d
```

To check the status of the container, run:

```bash
docker-compose ps
```

To stop the container, run: 

```bash
docker-compose down
```

To shell into the container, run:

```bash
docker exec -ti api_api_1 bash
```

To see the logs of the container, run:

```bash
# Replace 50 with number of lines requires
docker logs --tail 50 --follow --timestamps api_api_1
```
