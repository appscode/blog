#! /usr/bin/bash

kubectl delete -f job.yaml
kubectl delete -f pvc.yaml

kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml
kubectl apply -f job.yaml
kubectl apply -f pvc.yaml