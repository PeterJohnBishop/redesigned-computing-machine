# redesigned-computing-machine

Once deployed, a cluster of 11 identical server pods is launched and each immediately beings connecting to the others via websocket connections.

## documentation links

https://github.com/gin-gonic/gin
https://github.com/gorilla/websocket?tab=readme-ov-file

docker build -t peterjbishop/torrent-server:latest .

docker push peterjbishop/torrent-server:latest

minikube start

kubectl apply -f deployment.yaml

minikube service torrent-server

kubectl get pods -l app=torrent-server -o wide

kubectl logs

kubectl describe pod //name

kubectl describe pod torrent-server-69bfc7fcb9-drkll 

kubectl logs -l app=torrent-server -f --max-log-requests=10
