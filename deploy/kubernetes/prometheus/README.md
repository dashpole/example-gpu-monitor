# Usage
`# extra step for GKE`  
`EMAIL=your.google.cloud.email@example.org`  
`kubectl create clusterrolebinding prometheus-admin --clusterrole=cluster-admin --user=$EMAIL`  

`# view prometheus console through service, after externalIP is created`  
`kubectl get svc --namespace=monitoring -w`  

`# navigate to externalIP:9090`  
