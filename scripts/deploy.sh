#!/bin/sh
helm install my-redpanda redpanda-data/redpanda --version 25.1.1 -f helm/redpanda/values.yaml > /dev/null
helm install pguser oci://registry-1.docker.io/bitnamicharts/postgresql -f helm/pguser/values.yaml > /dev/null
kubectl wait --for=condition=Ready pod/pguser-postgresql-primary-0 --timeout=2m

kubectl exec my-redpanda-0 -it -- rpk topic create -p 1 new_users
kubectl exec my-redpanda-0 -it -- rpk topic create -p 1 new_orders
kubectl exec my-redpanda-0 -it -- rpk topic create -p 1 order_is_paid_topic
kubectl exec my-redpanda-0 -it -- rpk topic create -p 1 order_payment_failed_topic

export POSTGRES_PASSWORD=$(kubectl get secret --namespace default pguser-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
cat ./scripts/db_init/initdb.sql | kubectl exec -i pguser-postgresql-primary-0 -- sh -c "PGPASSWORD='$POSTGRES_PASSWORD' psql -U postgres -d postgres"

helm install userapp-ingress ingress-nginx/ingress-nginx -f helm/nginx-ingress/values.yaml
kubectl wait --for=condition=Ready deployment/userapp-ingress-ingress-nginx-controller --timeout 2m

helm install userapp helm/userapp
helm install orderapp helm/orderapp
helm install billingapp helm/billingapp
helm install notificationapp helm/notificationapp
