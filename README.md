# redesigned-computing-machine

## documentation links

https://github.com/gin-gonic/gin
https://github.com/gorilla/websocket?tab=readme-ov-file

docker build -t peterjbishop/torrent-server:latest .

docker push peterjbishop/torrent-server:latest

minikube start

kubectl apply -f deployment.yaml

kubectl get pods -l app=ws-server

kubectl get svc server-headless

kubectl expose deployment torrent-server --type=NodePort --port=8080

minikube service torrent-server

kubectl get pods -l app=torrent-server -o wide

kubectl describe pod //name

kubectl describe pod torrent-server-69bfc7fcb9-drkll 

kubectl logs -l app=torrent-server -f --max-log-requests=10
