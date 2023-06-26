


REF: https://codefresh.io/docs/docs/integrations/docker-registries/github-container-registry/
REF: https://leeyoongti.medium.com/helm-in-kubernetes-part-4-publish-helm-chart-to-artifact-hub-using-github-pages-ab7f55904faa#:~:text=Add%20chart%20to%20Artifact%20Hub,in%20all%20the%20required%20information

```bash
version='v0.1.0'
docker login ghcr.io --username github-account
docker build -t ghcr.io/jmcgrath207/par:$version .
docker build -t ghcr.io/jmcgrath207/par:latest .
docker push ghcr.io/jmcgrath207/par:$version 
docker push ghcr.io/jmcgrath207/par:latest 
helm package chart --destination chart
helm repo index --merge index.yaml chart/.
```