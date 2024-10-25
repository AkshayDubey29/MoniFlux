docker stop /moniflux-api
docker stop /moniflux-loadgen
docker rm /moniflux-api
docker rm /moniflux-loadgen
docker build -t moniflux-loadgen -f Dockerfile.loadgen .
docker run -d --network atyaas-network --name moniflux-loadgen -p 9080:9080 moniflux-loadgen
docker build -t moniflux-api -f Dockerfile.api .
docker run -d --network atyaas-network --name moniflux-api -p 8080:8080 moniflux-api

